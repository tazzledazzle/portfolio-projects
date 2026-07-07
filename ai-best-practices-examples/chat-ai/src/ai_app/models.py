from pydantic import BaseModel, Field


class ChatRequest(BaseModel):
    user_id: str = Field(default="demo-user")
    message: str = Field(min_length=1)


class ChatMetadata(BaseModel):
    tool_invoked: str | None = None
    memory_hits: int = 0


class ChatResult(BaseModel):
    text: str
    metadata: ChatMetadata
