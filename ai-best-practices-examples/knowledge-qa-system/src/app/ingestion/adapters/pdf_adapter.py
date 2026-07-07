from pathlib import Path

from pypdf import PdfReader

from app.domain.models import DocumentRecord


def load_pdf_document(source_uri: str, title: str | None = None) -> DocumentRecord:
    path = Path(source_uri)
    if path.suffix.lower() == ".pdf":
        reader = PdfReader(str(path))
        pages: list[str] = []
        for page in reader.pages:
            pages.append((page.extract_text() or "").strip())
        text = "\n".join(part for part in pages if part).strip()
        if not text:
            raise ValueError(f"No extractable text found in PDF: {source_uri}")
    else:
        # Keep a text fallback for local fixtures and debugging.
        text = path.read_text(encoding="utf-8")

    return DocumentRecord(
        document_id=f"pdf::{path.name}",
        source_type="pdf",
        source_uri=source_uri,
        title=title or path.stem,
        text=text,
    )
