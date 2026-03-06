from __future__ import annotations

import json
from datetime import datetime, timezone

from db import db_cursor

ALLOWED_STATUSES = {
    "compliant",
    "noncompliant",
    "unknown",
    "not_applicable",
    "partially_configured",
}


def _evaluate_exact(actual: dict, expected):
    if actual is None:
        return "unknown", "setting missing"
    value = actual.get("value_text")
    if value == "":
        value = actual.get("value_json", {}).get("raw")
    if value is None:
        return "unknown", "missing value"
    return ("compliant", "exact value match") if str(value).strip() == str(expected).strip() else ("noncompliant", "exact value mismatch")


def _evaluate_boolean(actual: dict, expected):
    if actual is None:
        return "unknown", "setting missing"
    if actual.get("value_bool") is None:
        raw = str(actual.get("value_text") or "").strip().lower()
        if raw in {"enabled", "true", "on"}:
            actual_bool = True
        elif raw in {"disabled", "false", "off"}:
            actual_bool = False
        else:
            return "unknown", "missing boolean value"
    else:
        actual_bool = bool(actual.get("value_bool"))
    expected_bool = bool(expected)
    return ("compliant", "boolean value match") if actual_bool == expected_bool else ("noncompliant", "boolean value mismatch")


def _evaluate_numeric_threshold(actual: dict, expected):
    if actual is None:
        return "unknown", "setting missing"
    value = actual.get("value_number")
    if value is None:
        try:
            value = float(str(actual.get("value_text") or "").strip())
        except ValueError:
            return "unknown", "missing numeric value"

    if isinstance(expected, (int, float)):
        threshold = float(expected)
        return ("compliant", "value meets minimum threshold") if float(value) >= threshold else ("noncompliant", "value below threshold")

    if not isinstance(expected, dict):
        return "unknown", "invalid expected threshold"

    if "min" in expected and float(value) < float(expected["min"]):
        return "noncompliant", "value below minimum"
    if "max" in expected and float(value) > float(expected["max"]):
        return "noncompliant", "value above maximum"
    return "compliant", "value within threshold bounds"


def _evaluate_set_membership(actual: dict, expected):
    if actual is None:
        return "unknown", "setting missing"
    actual_values = actual.get("value_json", {}).get("values")
    if not actual_values:
        raw = str(actual.get("value_text") or "")
        actual_values = [item.strip() for item in raw.replace(";", ",").split(",") if item.strip()]
    if not actual_values:
        return "unknown", "missing set values"

    expected_values = expected.get("values") if isinstance(expected, dict) else expected
    if not isinstance(expected_values, list):
        expected_values = [expected_values]

    actual_set = {str(item).strip().lower() for item in actual_values}
    expected_set = {str(item).strip().lower() for item in expected_values if str(item).strip()}
    if not expected_set:
        return "not_applicable", "no expected set values provided"

    matched = actual_set.intersection(expected_set)
    if matched == expected_set:
        return "compliant", "all expected set values present"
    if matched:
        return "partially_configured", "some expected set values present"
    return "noncompliant", "expected set values not found"


def _evaluate(check_type: str, actual: dict | None, expected):
    check = (check_type or "exact").strip().lower()
    if check == "exact":
        return _evaluate_exact(actual, expected)
    if check == "boolean":
        return _evaluate_boolean(actual, expected)
    if check == "numeric_threshold":
        return _evaluate_numeric_threshold(actual, expected)
    if check == "set_membership":
        return _evaluate_set_membership(actual, expected)
    return "unknown", f"unsupported check type: {check_type}"


def run_assessment(assessment_run_id: int) -> dict:
    with db_cursor() as (conn, cur):
        cur.execute(
            """
            SELECT id, policy_source_id, framework_id, version_id
            FROM assessment_runs
            WHERE id = %s
            """,
            (assessment_run_id,),
        )
        run_row = cur.fetchone()
        if not run_row:
            raise ValueError(f"assessment run {assessment_run_id} not found")

        _, policy_source_id, framework_id, version_id = run_row

        cur.execute(
            """
            SELECT setting_key, value_text, value_number, value_bool, value_json
            FROM policy_settings
            WHERE policy_source_id = %s
            """,
            (policy_source_id,),
        )
        settings = {}
        for key, value_text, value_number, value_bool, value_json in cur.fetchall():
            settings[key] = {
                "value_text": value_text or "",
                "value_number": float(value_number) if value_number is not None else None,
                "value_bool": value_bool,
                "value_json": value_json or {},
            }

        if framework_id and version_id:
            cur.execute(
                """
                SELECT id, rule_id, setting_key, check_type, expected_value
                FROM benchmark_policy_rules
                WHERE framework_id = %s AND version_id = %s
                ORDER BY id ASC
                """,
                (framework_id, version_id),
            )
        else:
            cur.execute(
                """
                SELECT id, rule_id, setting_key, check_type, expected_value
                FROM benchmark_policy_rules
                ORDER BY id ASC
                """,
            )

        rules = cur.fetchall()
        cur.execute("DELETE FROM assessment_results WHERE assessment_run_id = %s", (assessment_run_id,))

        counts = {status: 0 for status in ALLOWED_STATUSES}
        for rule_id, external_rule_id, setting_key, check_type, expected_value in rules:
            expected = expected_value or {}
            actual = settings.get(setting_key)
            status, details = _evaluate(check_type, actual, expected)
            if status not in ALLOWED_STATUSES:
                status = "unknown"
            counts[status] += 1
            cur.execute(
                """
                INSERT INTO assessment_results (
                    assessment_run_id, benchmark_policy_rule_id, rule_id, setting_key, status,
                    actual_value, expected_value, details
                ) VALUES (%s, %s, %s, %s, %s, %s::jsonb, %s::jsonb, %s)
                """,
                (
                    assessment_run_id,
                    rule_id,
                    external_rule_id or "",
                    setting_key or "",
                    status,
                    json.dumps(actual or {}),
                    json.dumps(expected or {}),
                    details,
                ),
            )

        cur.execute(
            "UPDATE assessment_runs SET status = %s, error = '', completed_at = %s WHERE id = %s",
            ("completed", datetime.now(timezone.utc), assessment_run_id),
        )
        conn.commit()

    return {"assessment_run_id": assessment_run_id, "counts": counts}
