from __future__ import annotations

import re
import xml.etree.ElementTree as ET
from pathlib import Path

from canonical import CanonicalPolicySetting


def _normalize_key(value: str) -> str:
    normalized = re.sub(r"[^a-zA-Z0-9]+", "_", (value or "").strip().lower())
    return normalized.strip("_")


def parse_gpmc_xml(path: str) -> list[dict]:
    xml_path = Path(path)
    root = ET.fromstring(xml_path.read_text(encoding="utf-8", errors="ignore"))
    settings: dict[str, CanonicalPolicySetting] = {}

    for node in root.findall(".//Policy"):
        name = (node.findtext(".//Name") or node.get("name") or "").strip()
        if not name:
            continue
        value = (
            (node.findtext(".//State") or "").strip()
            or (node.findtext(".//Value") or "").strip()
            or (node.findtext(".//SettingValue") or "").strip()
        )
        raw_text = ET.tostring(node, encoding="unicode").lower()
        scope = "user" if "user" in raw_text else "computer"
        key = _normalize_key(name)
        setting = CanonicalPolicySetting(
            setting_key=key,
            setting_name=name,
            canonical_type="text",
            scope=scope,
            value_text=value,
            value_json={"raw": value},
        )
        if value.strip().isdigit():
            setting.canonical_type = "numeric"
            setting.value_number = float(value.strip())
        elif value.strip().lower() in {"enabled", "disabled", "true", "false"}:
            setting.canonical_type = "boolean"
            setting.value_bool = value.strip().lower() in {"enabled", "true"}
        settings[key] = setting

    return [entry.as_db_row() for entry in settings.values()]
