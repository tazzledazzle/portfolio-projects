from pydantic import BaseModel, Field


class SourceInput(BaseModel):
    source_type: str
    source_uri: str
    title: str | None = None


class IngestRequest(BaseModel):
    sources: list[SourceInput] = Field(default_factory=list)
    mode: str = "full"


class IngestResponse(BaseModel):
    job_id: str
    documents_ingested: int
    chunks_written: int
    failures: list[str] = Field(default_factory=list)
