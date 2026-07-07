import uuid

from app.core.config import settings
from app.domain.models import ChunkRecord, DocumentRecord
from app.ingestion.adapters.notion_adapter import load_notion_document
from app.ingestion.adapters.pdf_adapter import load_pdf_document
from app.ingestion.adapters.web_adapter import load_web_document
from app.ingestion.chunking import chunk_text
from app.schemas.ingest import IngestRequest, IngestResponse
from app.retrieval.vector_store import ChromaVectorStore


class IngestionService:
    def __init__(self, vector_store: ChromaVectorStore | None = None) -> None:
        self.vector_store = vector_store or ChromaVectorStore()

    def ingest(self, request: IngestRequest) -> IngestResponse:
        job_id = str(uuid.uuid4())
        failures: list[str] = []
        all_chunks: list[ChunkRecord] = []
        documents_ingested = 0

        for source in request.sources:
            try:
                document = self._load(source.source_type, source.source_uri, source.title)
                documents_ingested += 1
                all_chunks.extend(self._chunk_document(document))
            except Exception as exc:  # noqa: BLE001
                failures.append(f"{source.source_type}:{source.source_uri}: {exc}")

        self.vector_store.upsert_chunks(all_chunks)
        return IngestResponse(
            job_id=job_id,
            documents_ingested=documents_ingested,
            chunks_written=len(all_chunks),
            failures=failures,
        )

    def _load(self, source_type: str, source_uri: str, title: str | None) -> DocumentRecord:
        if source_type == "pdf":
            return load_pdf_document(source_uri, title)
        if source_type == "notion":
            return load_notion_document(source_uri, title)
        if source_type == "web":
            return load_web_document(source_uri, title)
        raise ValueError(f"Unsupported source_type: {source_type}")

    def _chunk_document(self, document: DocumentRecord) -> list[ChunkRecord]:
        pieces = chunk_text(document.text, settings.chunk_size, settings.chunk_overlap)
        return [
            ChunkRecord(
                chunk_id=f"{document.document_id}::{index}",
                document_id=document.document_id,
                chunk_index=index,
                text=piece,
                source_type=document.source_type,
                source_uri=document.source_uri,
                title=document.title,
            )
            for index, piece in enumerate(pieces)
        ]
