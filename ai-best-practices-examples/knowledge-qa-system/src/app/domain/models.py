from pydantic import BaseModel, Field


class DocumentRecord(BaseModel):
    document_id: str
    source_type: str
    source_uri: str
    title: str
    text: str
    metadata: dict[str, str] = Field(default_factory=dict)


class ChunkRecord(BaseModel):
    chunk_id: str
    document_id: str
    chunk_index: int
    text: str
    source_type: str
    source_uri: str
    title: str


class Citation(BaseModel):
    chunk_id: str
    title: str
    source_uri: str
    snippet: str
    score: float
