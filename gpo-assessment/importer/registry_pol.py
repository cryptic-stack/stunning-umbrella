from __future__ import annotations

import re
from pathlib import Path

from canonical import CanonicalPolicySetting


def _normalize_key(value: str) -> str:
    normalized = re.sub(r"[^a-zA-Z0-9]+", "_", (value or "").strip().lower())
    return normalized.strip("_")


def parse_registry_pol(path: str) -> list[dict]:
    """
    Best-effort parser for text-form policy exports (e.g. LGPO /parse output).
    Binary Registry.pol parsing is out-of-scope for this initial implementation.
    """
    content = Path(path).read_text(encoding="utf-8", errors="ignore")
    settings = []
    for raw_line in content.splitlines():
        line = raw_line.strip()
        if not line or line.startswith(";"):
            continue
        if "=" not in line:
            continue
        key, value = [part.strip() for part in line.split("=", 1)]
        setting = CanonicalPolicySetting(
            setting_key=_normalize_key(key),
            setting_name=key,
            canonical_type="text",
            scope="computer",
            value_text=value,
            value_json={"raw": value},
        )
        if value.lower() in {"enabled", "disabled", "true", "false", "0", "1"}:
            setting.canonical_type = "boolean"
            setting.value_bool = value.lower() in {"enabled", "true", "1"}
        elif value.isdigit():
            setting.canonical_type = "numeric"
            setting.value_number = float(value)
        settings.append(setting.as_db_row())
    return settings
