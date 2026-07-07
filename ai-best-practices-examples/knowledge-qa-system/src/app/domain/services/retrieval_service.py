from app.core.config import settings
from app.domain.models import Citation
from app.retrieval.vector_store import ChromaVectorStore


class RetrievalService:
    def __init__(self, vector_store: ChromaVectorStore | None = None) -> None:
        self.vector_store = vector_store or ChromaVectorStore()

    def retrieve(self, query: str, top_k: int | None = None) -> tuple[list[Citation], list[str], bool]:
        citations, snippets = self.vector_store.retrieve(query, top_k or settings.retrieval_top_k)
        if not citations:
            return [], [], True
        filtered = [item for item in citations if item.score >= settings.min_similarity]
        if not filtered:
            return citations, snippets, True
        filtered_ids = {item.chunk_id for item in filtered}
        filtered_snippets = [snippet for item, snippet in zip(citations, snippets, strict=False) if item.chunk_id in filtered_ids]
        return filtered, filtered_snippets, False
