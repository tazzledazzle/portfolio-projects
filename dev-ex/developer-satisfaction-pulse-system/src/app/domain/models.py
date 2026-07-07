from pydantic import BaseModel, Field


class SurveyResponse(BaseModel):
    team: str = Field(min_length=1)
    score: int = Field(ge=0, le=10)
    comment: str | None = None
