from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class CanonicalPolicySetting:
    setting_key: str
    setting_name: str
    canonical_type: str
    scope: str
    value_text: str = ""
    value_number: float | None = None
    value_bool: bool | None = None
    value_json: dict = field(default_factory=dict)

    def as_db_row(self) -> dict:
        return {
            "setting_key": self.setting_key,
            "setting_name": self.setting_name,
            "canonical_type": self.canonical_type,
            "scope": self.scope,
            "value_text": self.value_text,
            "value_number": self.value_number,
            "value_bool": self.value_bool,
            "value_json": self.value_json or {},
        }

