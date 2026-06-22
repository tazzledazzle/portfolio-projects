from pydantic import BaseModel, Field

from app.domain.models import Citation


class ChatRequest(BaseModel):
    message: str
    conversation_id: str | None = None
    top_k: int | None = None
    stream: bool = True


class ChatResponse(BaseModel):
    answer: str
    citations: list[Citation] = Field(default_factory=list)
    partial_context: bool = False
