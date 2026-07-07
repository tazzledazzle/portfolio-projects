import re

import httpx

from app.core.config import settings
from app.domain.models import DocumentRecord


NOTION_PAGE_ID_RE = re.compile(r"([0-9a-fA-F]{32})")


def _extract_page_id(source_uri: str) -> str:
    maybe_id = source_uri.replace("-", "")
    if len(maybe_id) == 32 and maybe_id.isalnum():
        return maybe_id
    match = NOTION_PAGE_ID_RE.search(maybe_id)
    if not match:
        raise ValueError("source_uri must contain a valid Notion page id")
    return match.group(1)


def load_notion_document(source_uri: str, title: str | None = None) -> DocumentRecord:
    if not settings.notion_token:
        raise ValueError("KQ_NOTION_TOKEN is required for Notion ingestion")

    page_id = _extract_page_id(source_uri)
    headers = {
        "Authorization": f"Bearer {settings.notion_token}",
        "Notion-Version": settings.notion_version,
    }
    with httpx.Client(timeout=settings.notion_timeout_seconds) as client:
        page_resp = client.get(f"https://api.notion.com/v1/pages/{page_id}", headers=headers)
        page_resp.raise_for_status()
        blocks_resp = client.get(
            f"https://api.notion.com/v1/blocks/{page_id}/children?page_size=100",
            headers=headers,
        )
        blocks_resp.raise_for_status()

    page = page_resp.json()
    blocks = blocks_resp.json().get("results", [])
    block_texts: list[str] = []
    for block in blocks:
        block_type = block.get("type")
        typed = block.get(block_type, {}) if block_type else {}
        rich = typed.get("rich_text", [])
        if not rich:
            continue
        plain = "".join(str(part.get("plain_text", "")) for part in rich).strip()
        if plain:
            block_texts.append(plain)

    resolved_title = title
    if not resolved_title:
        title_prop = page.get("properties", {}).get("title", {})
        title_items = title_prop.get("title", [])
        resolved_title = "".join(str(item.get("plain_text", "")) for item in title_items).strip() or "Notion Page"

    body = "\n".join(block_texts).strip()
    if not body:
        raise ValueError("No text blocks found in Notion page")

    return DocumentRecord(
        document_id=f"notion::{page_id}",
        source_type="notion",
        source_uri=source_uri,
        title=resolved_title,
        text=body,
    )
