from ai_app.models import ChatRequest
from ai_app.services.conversation import ConversationService
from ai_app.services.memory import MemoryService
from ai_app.services.tools import ToolService


def test_tool_service_registers_and_executes_tools() -> None:
    service = ToolService()

    def echo_tool(query: str) -> str:
        return f"echo:{query}"

    service.register("echo", echo_tool)
    output = service.execute("echo", "hello")
    assert output == "echo:hello"


def test_conversation_routes_weather_query_to_tool(tmp_path) -> None:
    memory = MemoryService(store_path=tmp_path / "memory.json")
    tools = ToolService()
    tools.register("weather", lambda city: f"Sunny in {city}")
    convo = ConversationService(memory_service=memory, tool_service=tools)

    result = convo.chat(ChatRequest(user_id="u1", message="What's the weather in Paris?"))
    assert result.metadata.tool_invoked == "weather"
    assert "Sunny in Paris" in result.text
