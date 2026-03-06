from __future__ import annotations

import re
from pathlib import Path

from canonical import CanonicalPolicySetting


def _normalize_key(value: str) -> str:
    normalized = re.sub(r"[^a-zA-Z0-9]+", "_", (value or "").strip().lower())
    return normalized.strip("_")


def parse_secedit_inf(path: str) -> list[dict]:
    content = Path(path).read_text(encoding="utf-8", errors="ignore")
    settings = []
    current_scope = "computer"
    for raw_line in content.splitlines():
        line = raw_line.strip()
        if not line or line.startswith(";"):
            continue
        if line.startswith("[") and line.endswith("]"):
            section = line[1:-1].strip().lower()
            if "user" in section:
                current_scope = "user"
            elif "system" in section or "security" in section:
                current_scope = "computer"
            continue
        if "=" not in line:
            continue

        key, value = [part.strip() for part in line.split("=", 1)]
        setting = CanonicalPolicySetting(
            setting_key=_normalize_key(key),
            setting_name=key,
            canonical_type="text",
            scope=current_scope,
            value_text=value,
            value_json={"raw": value},
        )
        if value.isdigit():
            setting.canonical_type = "numeric"
            setting.value_number = float(value)
        elif value.lower() in {"enabled", "disabled", "true", "false", "0", "1"}:
            setting.canonical_type = "boolean"
            setting.value_bool = value.lower() in {"enabled", "true", "1"}
        elif "," in value:
            values = [item.strip() for item in value.split(",") if item.strip()]
            setting.canonical_type = "set"
            setting.value_json = {"values": values}
        settings.append(setting.as_db_row())
    return settings
