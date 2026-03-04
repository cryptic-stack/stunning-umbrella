from __future__ import annotations

import re
from pathlib import Path
from typing import List

import pdfplumber

try:
    from .models import CanonicalSafeguard
except ImportError:  # pragma: no cover
    from models import CanonicalSafeguard


SAFEGUARD_PATTERN = re.compile(r"^(?P<safeguard>\d+\.\d+)\s+(?P<title>.+)$")


def _ig_flags(text: str) -> tuple[bool, bool, bool]:
    lowered = text.lower()
    return ("ig1" in lowered, "ig2" in lowered, "ig3" in lowered)


def parse_pdf(path: str, framework: str, version: str) -> List[CanonicalSafeguard]:
    records: list[CanonicalSafeguard] = []
    current = None
    details: list[str] = []

    with pdfplumber.open(Path(path)) as pdf:
        for page in pdf.pages:
            page_text = page.extract_text() or ""
            for raw_line in page_text.splitlines():
                line = raw_line.strip()
                if not line:
                    continue

                match = SAFEGUARD_PATTERN.match(line)
                if match:
                    if current is not None:
                        description = " ".join(details).strip()
                        ig1, ig2, ig3 = _ig_flags(description)
                        records.append(
                            CanonicalSafeguard(
                                framework=framework,
                                version=version,
                                control_id=current["control_id"],
                                safeguard_id=current["safeguard_id"],
                                title=current["title"],
                                description=description,
                                ig1=ig1,
                                ig2=ig2,
                                ig3=ig3,
                            )
                        )
                    safeguard_id = match.group("safeguard")
                    current = {
                        "safeguard_id": safeguard_id,
                        "control_id": safeguard_id.split(".", 1)[0],
                        "title": match.group("title").strip(),
                    }
                    details = []
                    continue

                if current is not None:
                    details.append(line)

    if current is not None:
        description = " ".join(details).strip()
        ig1, ig2, ig3 = _ig_flags(description)
        records.append(
            CanonicalSafeguard(
                framework=framework,
                version=version,
                control_id=current["control_id"],
                safeguard_id=current["safeguard_id"],
                title=current["title"],
                description=description,
                ig1=ig1,
                ig2=ig2,
                ig3=ig3,
            )
        )

    return records
