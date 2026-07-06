# Research — Features

**Researched:** 2026-05-19

## Table stakes (must have for credible demo)

- Temporal UI showing workflow history
- At least one multi-activity AI-shaped workflow
- Trace showing parent workflow span + child activity spans
- Grafana dashboard with workflow throughput and activity latency
- Correlation: find logs by `workflow_id`, jump to trace

## Differentiators (portfolio narrative)

- Kotlin worker + OTel (not Python-only AI demo)
- Explicit Jaeger → Tempo LGTM migration story
- Three distinct AI patterns (RAG, agent, batch)
- OPERATIONS.md with validated runbook steps

## Defer to v2

- Real vector DB + embeddings
- Temporal Cloud + mTLS
- SLO burn rate alerting in production
- Cost attribution per workflow type

## Feature → requirement mapping

| Feature | Requirements |
|---------|----------------|
| Compose platform | FOUND-01, FOUND-02 |
| Temporal worker | TEMP-01..03 |
| Sample workflows | WF-01..04 |
| OTel signals | OTEL-01..04 |
| Jaeger/Prometheus | BACK-01..03 |
| Grafana LGTM | LGTM-01..04 |
| On-call readiness | OPS-01..03 |
