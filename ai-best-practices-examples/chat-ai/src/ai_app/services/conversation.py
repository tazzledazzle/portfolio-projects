from collections.abc import Iterator
from pathlib import Path

from ai_app.models import ChatMetadata, ChatRequest, ChatResult
from ai_app.services.memory import MemoryService
from ai_app.services.tools import ToolService


class ConversationService:
    def __init__(
        self,
        memory_service: MemoryService | None = None,
        tool_service: ToolService | None = None,
    ) -> None:
        self.memory_service = memory_service or MemoryService(store_path=Path("chroma_data/memory.json"))
        self.tool_service = tool_service or ToolService()

    def chat(self, request: ChatRequest) -> ChatResult:
        memories = self.memory_service.retrieve(request.user_id, request.message, top_k=3)
        memory_hits = len(memories)
        tool_name = None
        lower = request.message.lower()
        response_text = f"You said: {request.message}"
        if "weather in " in lower and self.tool_service.has("weather"):
            tool_name = "weather"
            city = request.message.split("weather in ", 1)[-1].strip(" ?.")
            if not city:
                city = "your location"
            tool_output = self.tool_service.execute("weather", city)
            response_text = f"{tool_output}."
        elif memories:
            response_text = f"You said: {request.message}. I remember: {memories[0].text}"

        self.memory_service.add_turn(request.user_id, request.message)
        return ChatResult(text=response_text, metadata=ChatMetadata(tool_invoked=tool_name, memory_hits=memory_hits))

    def stream_text(self, text: str) -> Iterator[str]:
        for chunk in text.split():
            yield f"{chunk} "
