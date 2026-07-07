# ADR-001: Use Temporal for AI Workflow Orchestration

## Status
Accepted

## Context
AI workflows involve long-running steps (LLM calls, retrieval, tool use) that fail transiently and must survive process restarts. Ad-hoc job queues and cron lack durable state, visibility into partial progress, and deterministic replay for debugging.

## Decision
Use **Temporal** as the orchestration engine. Workflow code expresses control flow; activities perform side effects. All sample AI patterns (RAG, agent loop, batch eval) run as Temporal workflows on Kotlin workers.

## Alternatives Considered

- **Apache Airflow** — Strong for batch ETL; weaker for interactive, per-user AI requests and fine-grained activity retries.
- **LangGraph alone** — Excellent agent graphs but does not replace durable infrastructure; can run *inside* activities later.
- **Celery / BullMQ** — Simpler ops; no built-in workflow history, replay, or first-class long-running child workflows.

## Consequences

### Positive
- Durable execution with automatic retries and timeouts
- Temporal UI for inspecting failed AI steps
- Clear separation: deterministic workflows vs non-deterministic activities

### Negative
- Operational overhead (Temporal Server + persistence)
- Learning curve for workflow determinism rules
- Kotlin SDK maturity slightly behind Java/Go

## Trade-offs
Operational complexity is accepted to gain **debuggability and reliability** for multi-step AI flows—core portfolio narrative.
