# Knowledge Q&A System

Demo-ready RAG MVP over private documents with:
- PDF, Notion, and web URL ingestion
- Chroma-backed retrieval
- citation-aware chat answers
- FastAPI SSE streaming endpoint

## Setup

```bash
python3 -m pip install -e ".[dev]"
```

## Run API

```bash
python3 -m uvicorn app.main:app --host 127.0.0.1 --port 8010
```

## Run tests

```bash
python3 -m pytest -q
```
