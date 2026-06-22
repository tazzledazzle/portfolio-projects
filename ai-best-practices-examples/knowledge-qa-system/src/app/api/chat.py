import json

from fastapi import APIRouter
from fastapi import Response
from fastapi.responses import JSONResponse, StreamingResponse

from app.domain.services.chat_service import ChatService
from app.schemas.chat import ChatRequest

router = APIRouter(prefix="", tags=["chat"])
chat_service = ChatService()


@router.post("/chat", response_model=None)
def chat(request: ChatRequest) -> Response:
    response = chat_service.answer(request.message, top_k=request.top_k)
    if not request.stream:
        return JSONResponse(response.model_dump())

    def event_stream():
        yield "event: retrieval_started\ndata: {}\n\n"
        for citation in response.citations:
            yield f"event: citation\ndata: {json.dumps(citation.model_dump())}\n\n"
        for token in chat_service.stream_answer(response.answer):
            yield f"event: token\ndata: {json.dumps({'token': token})}\n\n"
        yield f"event: final\ndata: {json.dumps(response.model_dump())}\n\n"

    return StreamingResponse(event_stream(), media_type="text/event-stream")
