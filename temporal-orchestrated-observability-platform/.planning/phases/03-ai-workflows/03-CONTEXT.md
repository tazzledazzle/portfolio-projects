# Phase 3: AI Workflows — Context

**Gathered:** 2026-05-19

<domain>

## Phase boundary

Deliver **WF-01** through **WF-04** — three AI-shaped Temporal workflows with **stubbed** external dependencies (no live API keys in CI).

| Workflow | Activities | Demonstrates |
|----------|------------|--------------|
| `RagQaWorkflow` | embedQuery, vectorSearch, llmComplete | Sequential RAG |
| `AgentToolsWorkflow` | planStep, callTool, synthesize | Retry + branching |
| `BatchEvalWorkflow` | loadDataset, scoreItem×N, aggregate | Fan-out |

**Out of scope:** OTel spans (Phase 4), Grafana (Phase 6).

</domain>

<decisions>

- **D-01:** LLM stub HTTP service in Compose (`llm-stub:8080`) returning deterministic JSON.
- **D-02:** `vectorSearch` returns fixture chunks from classpath JSON.
- **D-03:** Agent workflow injects failure on first `callTool` when env `SIMULATE_TOOL_FAILURE=true`.
- **D-04:** Batch workflow N=5 default for CI speed.
- **D-05:** Starter subcommands: `rag`, `agent`, `batch`.

</decisions>

<canonical_refs>

- `docs/ARCHITECTURE.md` — Component 1 (AI Workflows)
- `.planning/research/PITFALLS.md` — non-determinism, live LLM

</canonical_refs>
