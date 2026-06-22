from fastapi import FastAPI
from fastapi.responses import HTMLResponse

from app.api.chat import router as chat_router
from app.api.ingest import router as ingest_router

app = FastAPI(title="Knowledge Q&A System MVP")
app.include_router(chat_router)
app.include_router(ingest_router)


@app.get("/health/live")
def health_live() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/", response_class=HTMLResponse)
def index() -> str:
    return """
<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Knowledge Q&A System</title>
  </head>
  <body>
    <h1>Knowledge Q&A System</h1>
    <p>Use POST /ingest then POST /chat to query with citations.</p>
  </body>
</html>
"""
