from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    chroma_path: str = "chroma_data"
    collection_name: str = "knowledge_chunks_lc_v1"
    chunk_size: int = 800
    chunk_overlap: int = 120
    retrieval_top_k: int = 6
    min_similarity: float = 0.25
    embedding_backend: str = "sentence_transformers"
    embedding_model_name: str = "sentence-transformers/all-MiniLM-L6-v2"
    embedding_device: str = "cpu"
    embedding_hash_size: int = 384
    notion_token: str | None = None
    notion_version: str = "2022-06-28"
    notion_timeout_seconds: float = 20.0

    model_config = SettingsConfigDict(env_prefix="KQ_", extra="ignore")


settings = Settings()
