import httpx
from bs4 import BeautifulSoup

from app.domain.models import DocumentRecord


def load_web_document(source_uri: str, title: str | None = None) -> DocumentRecord:
    with httpx.Client(timeout=15.0, follow_redirects=True) as client:
        response = client.get(source_uri)
        response.raise_for_status()
    soup = BeautifulSoup(response.text, "html.parser")
    text = " ".join(soup.stripped_strings)
    return DocumentRecord(
        document_id=f"web::{source_uri}",
        source_type="web",
        source_uri=source_uri,
        title=title or (soup.title.string.strip() if soup.title and soup.title.string else source_uri),
        text=text,
    )
