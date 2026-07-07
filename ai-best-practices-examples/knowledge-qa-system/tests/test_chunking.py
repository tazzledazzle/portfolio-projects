from app.ingestion.chunking import chunk_text


def test_chunk_text_overlap() -> None:
    text = "abcdefghijklmnopqrstuvwxyz"
    chunks = chunk_text(text=text, chunk_size=10, overlap=2)
    assert chunks[0] == "abcdefghij"
    assert chunks[1].startswith("ijkl")
    assert len(chunks) >= 3
