from pydantic import BaseModel, Field


class CanonicalSafeguard(BaseModel):
    framework: str = Field(default="CIS Controls")
    version: str
    control_id: str
    safeguard_id: str
    title: str
    description: str = ""
    level: str = ""
    ig1: bool = False
    ig2: bool = False
    ig3: bool = False
