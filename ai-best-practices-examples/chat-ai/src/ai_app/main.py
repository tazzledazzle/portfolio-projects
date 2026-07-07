from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from fastapi.responses import StreamingResponse

from ai_app.models import ChatRequest
from ai_app.services.conversation import ConversationService
from ai_app.services.tools import ToolService
from ai_app.tools.weather import fetch_weather

app = FastAPI(title="AI Chat Assistant MVP")
tool_service = ToolService()
tool_service.register("weather", fetch_weather)
conversation_service = ConversationService(tool_service=tool_service)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/chat")
def chat(request: ChatRequest) -> StreamingResponse:
    result = conversation_service.chat(request)

    def event_stream() -> str:
        for token in conversation_service.stream_text(result.text):
            yield token

    return StreamingResponse(event_stream(), media_type="text/plain")


@app.get("/", response_class=HTMLResponse)
def index() -> str:
    return """
<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>AI Chat Assistant MVP</title>
    <style>
      body { font-family: Arial, sans-serif; max-width: 800px; margin: 2rem auto; }
      #messages { border: 1px solid #ccc; border-radius: 8px; padding: 1rem; min-height: 240px; }
      .msg { margin-bottom: 0.75rem; }
      .user { font-weight: bold; }
      .assistant { color: #1f4c9e; }
      form { display: flex; gap: 0.5rem; margin-top: 1rem; }
      input { flex: 1; padding: 0.5rem; }
      button { padding: 0.5rem 0.75rem; }
    </style>
  </head>
  <body>
    <h1>AI Chat Assistant MVP</h1>
    <div id="messages"></div>
    <form id="chat-form">
      <input id="message-input" type="text" placeholder="Ask something..." required />
      <button type="submit">Send</button>
    </form>
    <script>
      const userId = "demo-user";
      const messages = document.getElementById("messages");
      const form = document.getElementById("chat-form");
      const input = document.getElementById("message-input");

      function appendMessage(role, text) {
        const row = document.createElement("div");
        row.className = "msg " + role;
        row.textContent = (role === "user" ? "You: " : "Assistant: ") + text;
        messages.appendChild(row);
        messages.scrollTop = messages.scrollHeight;
        return row;
      }

      form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const message = input.value.trim();
        if (!message) return;
        appendMessage("user", message);
        input.value = "";
        const assistantRow = appendMessage("assistant", "");

        const response = await fetch('/chat', {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ user_id: userId, message }),
        });

        if (!response.ok || !response.body) {
          assistantRow.textContent = "Assistant: Chat request failed.";
          return;
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          assistantRow.textContent += decoder.decode(value, { stream: true });
        }
      });
    </script>
  </body>
</html>
"""
