# Chat AI

Minimal FastAPI-based AI chat assistant MVP with:
- streaming chat endpoint
- persistent vector-like memory retrieval
- tool abstraction with weather integration
- lightweight browser chat UI

## Setup

```bash
python3 -m pip install -e ".[dev]"
```

## Run

```bash
python3 -m uvicorn ai_app.main:app --host 127.0.0.1 --port 8000
```

Open [http://127.0.0.1:8000](http://127.0.0.1:8000).

## Test

```bash
python3 -m pytest -q
```
