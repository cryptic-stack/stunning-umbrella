from dataclasses import dataclass


@dataclass
class SafeguardRecord:
    safeguard_id: str
    title: str
    description: str
    ig1: bool
    ig2: bool
    ig3: bool


@dataclass
class DiffResult:
    change_type: str
    safeguard_old: str
    safeguard_new: str
    old_text: str
    new_text: str
    similarity: float
