# AI Chat Assistant MVP Design

## Goal

Build a polished, demo-ready AI chat assistant MVP with:
- multi-turn conversation
- persistent vector memory retrieval
- one external tool integration
- streaming assistant responses
- simple web chat UI

## Chosen Architecture

Approach A: thin UI + API-first backend.

- Lightweight browser chat client for fast demo iteration.
- FastAPI backend handles orchestration, memory retrieval, tool execution, and streaming.
- Tool interface supports one tool now and easy extension for a second tool.
- Vector memory service persists and retrieves semantically similar prior turns.

## Components

- `ChatUI`: browser client for sending messages and rendering streamed responses.
- `ChatAPI`: FastAPI routes for chat and health.
- `ConversationService`: orchestration layer for prompt context, memory lookup, tool flow, and final response.
- `MemoryService`: stores user/assistant turns, computes embeddings, retrieves top-k matches.
- `ToolService`: registry and execution abstraction for external tools.
- `LLMAdapter`: provider-agnostic interface for direct responses and tool decisions.
- `PromptBuilder`: deterministic prompt assembly from system prompt, recent turns, and memory snippets.

## Data Flow

1. User sends a message from the web UI.
2. API persists the user turn and loads recent context.
3. Memory service retrieves top-k relevant historical snippets.
4. Conversation service builds prompt context and asks the LLM adapter.
5. LLM returns either direct response or a tool call request.
6. If tool call: execute tool and re-run LLM with tool result.
7. Stream final assistant response back to UI.
8. Persist assistant turn and update embeddings.

## Error Handling

- LLM errors return a graceful fallback response and log details.
- Tool timeout/failure degrades gracefully with a clear assistant message.
- Memory misses are non-fatal and fall back to recent chat context.
- Streaming interruption preserves partial output and marks turn metadata.

## Testing Strategy

- Unit tests:
  - prompt assembly behavior
  - tool-routing decisions
  - memory ranking behavior
- Integration tests:
  - direct `/chat` path
  - tool `/chat` path
  - memory-augmented `/chat` path
- Smoke checks:
  - streaming works end-to-end
  - metadata records tool and memory usage

## MVP Boundaries

In scope:
- single-user demo workflow
- one external tool now, second-tool-ready abstraction
- reliability guardrails and explainable metadata

Out of scope:
- full production auth and permissions
- advanced evaluation pipelines
- multi-provider benchmarking
