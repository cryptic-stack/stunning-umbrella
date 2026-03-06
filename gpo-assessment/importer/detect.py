from __future__ import annotations

from pathlib import Path

KNOWN_SOURCE_TYPES = {"gpresult_xml", "gpmc_xml", "secedit_inf", "registry_pol"}


def detect_source_type(path: str, declared_type: str = "") -> str:
    normalized_declared = (declared_type or "").strip().lower()
    if normalized_declared in KNOWN_SOURCE_TYPES:
        return normalized_declared

    source_path = Path(path)
    extension = source_path.suffix.lower()
    if extension == ".inf":
        return "secedit_inf"
    if extension == ".pol":
        return "registry_pol"

    try:
        sample = source_path.read_bytes()[:65536]
    except OSError as exc:
        raise ValueError(f"unable to read source file for type detection: {exc}") from exc

    if sample.startswith(b"PReg"):
        return "registry_pol"

    text = sample.decode("utf-8", errors="ignore").lower()

    if extension == ".xml" or "<?xml" in text or "<" in text:
        if "<rsop" in text or "<computerresults" in text or "<userresults" in text:
            return "gpresult_xml"
        if "<gpo" in text or "<grouppolicyobjects" in text or "<gpmc" in text:
            return "gpmc_xml"
        if "<policy" in text:
            return "gpresult_xml"

    if "[system access]" in text or "[event audit]" in text or "[registry values]" in text:
        return "secedit_inf"
    if "[registry.pol]" in text or "\\software\\policies\\" in text:
        return "registry_pol"

    raise ValueError(
        "unable to auto-detect policy source type. Supported inputs: gpresult XML, GPMC XML, secedit INF, and Registry.pol"
    )
