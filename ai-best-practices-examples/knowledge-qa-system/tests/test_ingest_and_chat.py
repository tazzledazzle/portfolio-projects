from pathlib import Path

from fastapi.testclient import TestClient
from pypdf import PdfWriter

from app.main import app


def test_ingest_then_chat_flow() -> None:
    client = TestClient(app)
    pdf_path = Path(__file__).parent / "fixtures" / "sample_doc.pdf"
    if not pdf_path.exists():
        writer = PdfWriter()
        writer.add_blank_page(width=300, height=300)
        with pdf_path.open("wb") as handle:
            writer.write(handle)

    ingest_response = client.post(
        "/ingest",
        json={
            "sources": [
                {
                    "source_type": "pdf",
                    "source_uri": str(pdf_path),
                    "title": "Sample Fixture",
                }
            ],
            "mode": "full",
        },
    )
    assert ingest_response.status_code == 200
    ingest_json = ingest_response.json()
    assert ingest_json["documents_ingested"] in (0, 1)
    assert "failures" in ingest_json

    chat_response = client.post(
        "/chat",
        json={
            "message": "What is Chroma used for?",
            "stream": False,
        },
    )
    assert chat_response.status_code == 200
    payload = chat_response.json()
    assert "answer" in payload
    assert isinstance(payload.get("citations"), list)


def test_chat_stream_includes_events() -> None:
    client = TestClient(app)
    response = client.post("/chat", json={"message": "hello", "stream": True})
    assert response.status_code == 200
    body = response.text
    assert "event: retrieval_started" in body
    assert "event: token" in body
    assert "event: final" in body


def test_ingest_unsupported_source_returns_failure() -> None:
    client = TestClient(app)
    response = client.post(
        "/ingest",
        json={"sources": [{"source_type": "unknown", "source_uri": "x"}]},
    )
    assert response.status_code == 200
    payload = response.json()
    assert payload["documents_ingested"] == 0
    assert len(payload["failures"]) == 1
