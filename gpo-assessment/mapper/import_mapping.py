from __future__ import annotations

import csv
import json
from pathlib import Path

from db import db_cursor


def _normalize_check_type(value: str) -> str:
    normalized = (value or "").strip().lower()
    if normalized in {"exact", "equals"}:
        return "exact"
    if normalized in {"boolean", "bool"}:
        return "boolean"
    if normalized in {"numeric_threshold", "threshold", "numeric"}:
        return "numeric_threshold"
    if normalized in {"set_membership", "set", "contains_all"}:
        return "set_membership"
    return "exact"


def _parse_expected_value(raw_value):
    if isinstance(raw_value, (dict, list, int, float, bool)):
        return raw_value
    text = str(raw_value or "").strip()
    if text == "":
        return {}
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        lower = text.lower()
        if lower in {"true", "false"}:
            return lower == "true"
        try:
            if "." in text:
                return float(text)
            return int(text)
        except ValueError:
            return text


def _load_rows(path: Path) -> list[dict]:
    if path.suffix.lower() == ".json":
        payload = json.loads(path.read_text(encoding="utf-8-sig"))
        if isinstance(payload, dict) and "rules" in payload:
            payload = payload["rules"]
        return [dict(item) for item in payload]

    with path.open("r", encoding="utf-8-sig", newline="") as file:
        reader = csv.DictReader(file)
        return [dict(row) for row in reader]


def import_mapping_file(path: str, framework_id: int | None, version_id: int | None, source_label: str = "") -> int:
    source_path = Path(path)
    rows = _load_rows(source_path)
    inserted = 0
    with db_cursor() as (conn, cur):
        for row in rows:
            rule_id = (row.get("rule_id") or "").strip()
            setting_key = (row.get("setting_key") or "").strip()
            if not rule_id or not setting_key:
                continue
            cur.execute(
                """
                INSERT INTO benchmark_policy_rules (
                    framework_id, version_id, rule_id, benchmark_ref, title, description,
                    setting_key, check_type, expected_value, severity, source_label
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb, %s, %s)
                ON CONFLICT (framework_id, version_id, rule_id)
                DO UPDATE SET
                    benchmark_ref = EXCLUDED.benchmark_ref,
                    title = EXCLUDED.title,
                    description = EXCLUDED.description,
                    setting_key = EXCLUDED.setting_key,
                    check_type = EXCLUDED.check_type,
                    expected_value = EXCLUDED.expected_value,
                    severity = EXCLUDED.severity,
                    source_label = EXCLUDED.source_label
                """,
                (
                    framework_id,
                    version_id,
                    rule_id,
                    (row.get("benchmark_ref") or "").strip(),
                    (row.get("title") or "").strip(),
                    (row.get("description") or "").strip(),
                    setting_key,
                    _normalize_check_type(row.get("check_type") or ""),
                    json.dumps(_parse_expected_value(row.get("expected_value"))),
                    (row.get("severity") or "").strip(),
                    source_label or (row.get("source_label") or "").strip(),
                ),
            )
            inserted += 1
        conn.commit()
    return inserted
