from __future__ import annotations

from pathlib import Path
from typing import List

import pandas as pd

try:
    from .models import CanonicalSafeguard
except ImportError:  # pragma: no cover
    from models import CanonicalSafeguard


COLUMN_ALIASES = {
    "framework": ["framework", "benchmark", "standard"],
    "version": ["version", "framework_version"],
    "control_id": ["control_id", "control id", "control"],
    "safeguard_id": ["safeguard_id", "safeguard id", "safeguard", "recommendation"],
    "title": ["title", "name", "safeguard title"],
    "description": ["description", "details", "rationale", "text"],
    "ig1": ["ig1", "implementation group 1"],
    "ig2": ["ig2", "implementation group 2"],
    "ig3": ["ig3", "implementation group 3"],
}


def _normalize_column_map(columns: list[str]) -> dict[str, str]:
    normalized = {}
    for column in columns:
        key = str(column).strip().lower()
        normalized[key] = str(column)
    return normalized


def _resolve_column(column_map: dict[str, str], target: str) -> str | None:
    for alias in COLUMN_ALIASES[target]:
        if alias in column_map:
            return column_map[alias]
    return None


def _to_bool(value) -> bool:
    if isinstance(value, bool):
        return value
    if value is None:
        return False
    text = str(value).strip().lower()
    return text in {"1", "true", "yes", "y", "x"}


def parse_excel(path: str, framework: str, version: str) -> List[CanonicalSafeguard]:
    source = Path(path)
    if source.suffix.lower() == ".csv":
        df = pd.read_csv(source)
    else:
        df = pd.read_excel(source)

    df = df.fillna("")
    column_map = _normalize_column_map(list(df.columns))

    framework_col = _resolve_column(column_map, "framework")
    version_col = _resolve_column(column_map, "version")
    control_col = _resolve_column(column_map, "control_id")
    safeguard_col = _resolve_column(column_map, "safeguard_id")
    title_col = _resolve_column(column_map, "title")
    description_col = _resolve_column(column_map, "description")
    ig1_col = _resolve_column(column_map, "ig1")
    ig2_col = _resolve_column(column_map, "ig2")
    ig3_col = _resolve_column(column_map, "ig3")

    records: list[CanonicalSafeguard] = []
    for _, row in df.iterrows():
        safeguard_id = str(row.get(safeguard_col, "")).strip() if safeguard_col else ""
        if not safeguard_id:
            continue

        control_id = str(row.get(control_col, "")).strip() if control_col else ""
        if not control_id and "." in safeguard_id:
            control_id = safeguard_id.split(".", 1)[0]

        title = str(row.get(title_col, "")).strip() if title_col else ""
        description = str(row.get(description_col, "")).strip() if description_col else ""

        record = CanonicalSafeguard(
            framework=str(row.get(framework_col, framework)).strip() if framework_col else framework,
            version=str(row.get(version_col, version)).strip() if version_col else version,
            control_id=control_id,
            safeguard_id=safeguard_id,
            title=title,
            description=description,
            ig1=_to_bool(row.get(ig1_col, False)) if ig1_col else False,
            ig2=_to_bool(row.get(ig2_col, False)) if ig2_col else False,
            ig3=_to_bool(row.get(ig3_col, False)) if ig3_col else False,
        )
        records.append(record)

    return records
