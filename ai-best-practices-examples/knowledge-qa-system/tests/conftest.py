import os
from pathlib import Path


os.environ.setdefault("KQ_EMBEDDING_BACKEND", "hash")
os.environ.setdefault("KQ_EMBEDDING_HASH_SIZE", "384")
os.environ.setdefault("KQ_COLLECTION_NAME", "knowledge_chunks_test_v2")
os.environ.setdefault("KQ_CHROMA_PATH", str(Path(__file__).parent / ".chroma_test"))
