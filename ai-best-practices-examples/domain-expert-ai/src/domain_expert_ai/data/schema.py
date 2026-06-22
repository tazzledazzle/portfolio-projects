from typing import Literal

from pydantic import BaseModel, Field, field_validator


class DatasetRecord(BaseModel):
    question: str = Field(min_length=8)
    context: str = Field(min_length=4)
    answer: str = Field(min_length=8)
    jurisdiction: str = Field(min_length=2)
    risk_level: Literal["low", "medium", "high"]
    citations: list[str]

    @field_validator("citations")
    @classmethod
    def validate_citations(cls, value: list[str]) -> list[str]:
        cleaned = [item.strip() for item in value if item.strip()]
        if not cleaned:
            raise ValueError("citations must include at least one authority")
        return cleaned

