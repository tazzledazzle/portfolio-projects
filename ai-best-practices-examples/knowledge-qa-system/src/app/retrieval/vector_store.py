from langchain_chroma import Chroma
from langchain_core.documents import Document

from app.core.config import settings
from app.domain.models import ChunkRecord, Citation
from app.retrieval.embeddings import build_embedding_provider


class ChromaVectorStore:
    def __init__(self) -> None:
        self._store = Chroma(
            collection_name=settings.collection_name,
            persist_directory=settings.chroma_path,
            embedding_function=build_embedding_provider(),
        )

    def upsert_chunks(self, chunks: list[ChunkRecord]) -> None:
        if not chunks:
            return
        documents: list[Document] = []
        ids: list[str] = []
        for item in chunks:
            documents.append(
                Document(
                    page_content=item.text,
                    metadata={
                        "document_id": item.document_id,
                        "chunk_index": item.chunk_index,
                        "source_type": item.source_type,
                        "source_uri": item.source_uri,
                        "title": item.title,
                    },
                )
            )
            ids.append(item.chunk_id)
        self._store.add_documents(documents=documents, ids=ids)

    def retrieve(self, query: str, top_k: int) -> tuple[list[Citation], list[str]]:
        # Returns relevance scores in [0, 1].
        result = self._store.similarity_search_with_relevance_scores(query=query, k=top_k)
        citations: list[Citation] = []
        snippets: list[str] = []
        for doc, score in result:
            metadata = doc.metadata or {}
            chunk_id = f"{metadata.get('document_id', 'unknown')}::{metadata.get('chunk_index', 0)}"
            citations.append(
                Citation(
                    chunk_id=chunk_id,
                    title=str(metadata.get("title", "Untitled")),
                    source_uri=str(metadata.get("source_uri", "")),
                    snippet=doc.page_content[:220],
                    score=max(0.0, min(1.0, float(score))),
                )
            )
            snippets.append(doc.page_content)
        return citations, snippets
