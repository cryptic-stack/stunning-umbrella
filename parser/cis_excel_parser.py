from __future__ import annotations

import re
from pathlib import Path
from typing import Iterable, List

import pandas as pd

try:
    from .models import CanonicalSafeguard
except ImportError:  # pragma: no cover
    from models import CanonicalSafeguard


COLUMN_ALIASES = {
    "framework": ["framework", "benchmark", "standard"],
    "version": ["version", "framework_version"],
    "control_id": ["control_id", "control id", "control", "section #", "section"],
    "safeguard_id": ["safeguard_id", "safeguard id", "safeguard", "recommendation", "recommendation #"],
    "title": ["title", "name", "safeguard title"],
    "description": ["description", "details", "rationale", "text"],
    "profile": ["profile"],
    "ig1": ["ig1", "implementation group 1", "v8 ig1", "v7 ig1"],
    "ig2": ["ig2", "implementation group 2", "v8 ig2", "v7 ig2"],
    "ig3": ["ig3", "implementation group 3", "v8 ig3", "v7 ig3"],
}

LEVEL_PATTERN = re.compile(r"\((L1|L2)\)", re.IGNORECASE)
RECOMMENDATION_PATTERN = re.compile(r"\d+(?:\.\d+)+")


def _clean_text(value) -> str:
    if value is None:
        return ""
    text = str(value).strip()
    if text.lower() in {"nan", "none"}:
        return ""
    return text


def _normalize_column_map(columns: Iterable[str]) -> dict[str, str]:
    normalized = {}
    for column in columns:
        key = _clean_text(column).lower()
        if key:
            normalized[key] = str(column)
    return normalized


def _resolve_column(column_map: dict[str, str], target: str) -> str | None:
    for alias in COLUMN_ALIASES[target]:
        if alias in column_map:
            return column_map[alias]
    return None


def _to_bool(value) -> bool:
    text = _clean_text(value).lower()
    return text in {"1", "true", "yes", "y", "x"}


def _normalize_recommendation(value: str) -> str:
    text = _clean_text(value)
    if not text:
        return ""
    match = RECOMMENDATION_PATTERN.search(text)
    if not match:
        return ""
    return match.group(0)


def _derive_control_id(section_value: str, safeguard_id: str) -> str:
    section = _clean_text(section_value)
    if RECOMMENDATION_PATTERN.search(section):
        return section

    parts = safeguard_id.split(".")
    if len(parts) >= 2:
        return ".".join(parts[:-1])
    if parts:
        return parts[0]
    return ""


def _extract_level(title: str, recommendation: str, profile: str, sheet_name: str) -> str:
    for source in (title, recommendation, profile, sheet_name):
        match = LEVEL_PATTERN.search(_clean_text(source))
        if match:
            return match.group(1).upper()

    lowered_sheet = _clean_text(sheet_name).lower()
    if "level 1" in lowered_sheet:
        return "L1"
    if "level 2" in lowered_sheet:
        return "L2"
    return ""


def _parse_dataframe(df: pd.DataFrame, framework: str, version: str, sheet_name: str) -> list[CanonicalSafeguard]:
    df = df.fillna("")
    column_map = _normalize_column_map(list(df.columns))

    framework_col = _resolve_column(column_map, "framework")
    version_col = _resolve_column(column_map, "version")
    control_col = _resolve_column(column_map, "control_id")
    safeguard_col = _resolve_column(column_map, "safeguard_id")
    title_col = _resolve_column(column_map, "title")
    description_col = _resolve_column(column_map, "description")
    profile_col = _resolve_column(column_map, "profile")
    ig1_col = _resolve_column(column_map, "ig1")
    ig2_col = _resolve_column(column_map, "ig2")
    ig3_col = _resolve_column(column_map, "ig3")

    records: list[CanonicalSafeguard] = []

    for _, row in df.iterrows():
        raw_recommendation = _clean_text(row.get(safeguard_col, "")) if safeguard_col else ""
        safeguard_id = _normalize_recommendation(raw_recommendation)
        if not safeguard_id:
            continue

        title = _clean_text(row.get(title_col, "")) if title_col else ""
        description = _clean_text(row.get(description_col, "")) if description_col else ""
        profile = _clean_text(row.get(profile_col, "")) if profile_col else ""
        level = _extract_level(title, raw_recommendation, profile, sheet_name)

        control_value = _clean_text(row.get(control_col, "")) if control_col else ""
        control_id = _derive_control_id(control_value, safeguard_id)
        if not control_id:
            continue

        records.append(
            CanonicalSafeguard(
                framework=_clean_text(row.get(framework_col, framework)) if framework_col else framework,
                version=_clean_text(row.get(version_col, version)) if version_col else version,
                control_id=control_id,
                safeguard_id=safeguard_id,
                title=title,
                description=description,
                level=level,
                ig1=_to_bool(row.get(ig1_col, False)) if ig1_col else False,
                ig2=_to_bool(row.get(ig2_col, False)) if ig2_col else False,
                ig3=_to_bool(row.get(ig3_col, False)) if ig3_col else False,
            )
        )

    return records


def _select_sheets(sheet_names: list[str]) -> list[str]:
    combined = [name for name in sheet_names if "combined profiles" in name.lower()]
    if combined:
        return combined

    selected = []
    for name in sheet_names:
        lowered = name.lower()
        if "license" in lowered:
            continue
        selected.append(name)
    return selected


def _deduplicate(records: list[CanonicalSafeguard]) -> list[CanonicalSafeguard]:
    seen = set()
    unique: list[CanonicalSafeguard] = []
    for record in records:
        key = (record.framework.lower(), record.version, record.safeguard_id, record.level)
        if key in seen:
            continue
        seen.add(key)
        unique.append(record)
    return unique


def parse_excel(path: str, framework: str, version: str) -> List[CanonicalSafeguard]:
    source = Path(path)

    if source.suffix.lower() == ".csv":
        df = pd.read_csv(source)
        return _deduplicate(_parse_dataframe(df, framework, version, "csv"))

    workbook = pd.ExcelFile(source, engine="openpyxl")
    records: list[CanonicalSafeguard] = []

    for sheet_name in _select_sheets(workbook.sheet_names):
        df = pd.read_excel(source, sheet_name=sheet_name)
        records.extend(_parse_dataframe(df, framework, version, sheet_name))

    return _deduplicate(records)
