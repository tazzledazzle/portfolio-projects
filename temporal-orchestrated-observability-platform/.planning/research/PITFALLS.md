# Research — Pitfalls

**Researched:** 2026-05-19

## Orchestration

| Pitfall | Impact | Prevention |
|---------|--------|------------|
| Non-deterministic workflow code | Replay failures | Keep I/O in activities; no `Random()` in workflows |
| Unbounded activity timeouts | Stuck runs | Explicit `StartToCloseTimeout`; heartbeats for long LLM |
| Single task queue overload | Latency | Document split queues in Phase 3; optional `ai-activities` queue |

## Observability

| Pitfall | Impact | Prevention |
|---------|--------|------------|
| `workflow_id` as metric label | Prometheus OOM | ADR-004: use `workflow_type` only on metrics |
| Broken OTel collector loopback | Silent trace loss | Validate config in CI; integration test export |
| Logs without trace context | Broken log↔trace | Inject trace_id in log appender from OTel context |
| Dual Jaeger+Tempo confusion | Wrong drill-down UI | Document primary backend per phase in OPERATIONS.md |
| High-cardinality HTTP URLs on spans | Expensive traces | Template span names: `llm.complete` not full URL with query |

## AI / CI

| Pitfall | Impact | Prevention |
|---------|--------|------------|
| Live LLM in CI | Flaky tests | WireMock stub; feature flag for real keys locally |
| Large trace payloads (prompt text) | PII + cost | Record token counts, not full prompts in spans |

## Operations

| Pitfall | Impact | Prevention |
|---------|--------|------------|
| Paper runbooks | False confidence | OPS-03: manual validation record |
| Default Grafana password | Security embarrassment | Env-based admin password in Compose |
