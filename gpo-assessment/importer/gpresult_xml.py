from __future__ import annotations

import re
import xml.etree.ElementTree as ET
from pathlib import Path

from canonical import CanonicalPolicySetting


def _normalize_key(value: str) -> str:
    normalized = re.sub(r"[^a-zA-Z0-9]+", "_", (value or "").strip().lower())
    return normalized.strip("_")


def _to_setting(name: str, value: str, scope: str) -> CanonicalPolicySetting:
    clean = (value or "").strip()
    lower = clean.lower()

    if lower in {"enabled", "true", "on"}:
        return CanonicalPolicySetting(
            setting_key=_normalize_key(name),
            setting_name=name,
            canonical_type="boolean",
            scope=scope,
            value_text=clean,
            value_bool=True,
            value_json={"raw": clean},
        )
    if lower in {"disabled", "false", "off"}:
        return CanonicalPolicySetting(
            setting_key=_normalize_key(name),
            setting_name=name,
            canonical_type="boolean",
            scope=scope,
            value_text=clean,
            value_bool=False,
            value_json={"raw": clean},
        )
    if re.fullmatch(r"[-+]?\d+(\.\d+)?", clean):
        return CanonicalPolicySetting(
            setting_key=_normalize_key(name),
            setting_name=name,
            canonical_type="numeric",
            scope=scope,
            value_text=clean,
            value_number=float(clean),
            value_json={"raw": clean},
        )
    if ";" in clean or "," in clean:
        values = [item.strip() for item in re.split(r"[;,]", clean) if item.strip()]
        return CanonicalPolicySetting(
            setting_key=_normalize_key(name),
            setting_name=name,
            canonical_type="set",
            scope=scope,
            value_text=clean,
            value_json={"values": values},
        )

    return CanonicalPolicySetting(
        setting_key=_normalize_key(name),
        setting_name=name,
        canonical_type="text",
        scope=scope,
        value_text=clean,
        value_json={"raw": clean},
    )


def parse_gpresult_xml(path: str) -> list[dict]:
    xml_path = Path(path)
    root = ET.fromstring(xml_path.read_text(encoding="utf-8", errors="ignore"))
    settings: list[CanonicalPolicySetting] = []

    for policy in root.findall(".//Policy"):
        name = ""
        for tag in ("Name", "name"):
            name = (policy.findtext(tag) or "").strip()
            if name:
                break
        if not name:
            continue

        value = ""
        for tag in ("State", "state", "Value", "value", "Setting", "setting"):
            value = (policy.findtext(tag) or "").strip()
            if value:
                break

        scope = "computer"
        parent_text = ET.tostring(policy, encoding="unicode").lower()
        if "user" in parent_text:
            scope = "user"
        settings.append(_to_setting(name, value, scope))

    for extension in root.findall(".//ExtensionData"):
        name = (extension.findtext(".//Name") or "").strip()
        value = (extension.findtext(".//Value") or "").strip()
        if name:
            settings.append(_to_setting(name, value, "computer"))

    deduped: dict[str, CanonicalPolicySetting] = {}
    for entry in settings:
        if entry.setting_key:
            deduped[entry.setting_key] = entry

    return [entry.as_db_row() for entry in deduped.values()]
