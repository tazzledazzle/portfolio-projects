# ADR-004: Workflow and Run ID Propagation in Telemetry

## Status
Accepted

## Context
Operators must connect Grafana panels to a specific failing AI workflow. HTTP `correlation_id` alone is insufficient when work is async and retried inside Temporal.

## Decision
Propagate these attributes on **every trace span** and **structured log line** emitted from workers:

| Field | Source |
|-------|--------|
| `workflow_id` | Temporal `WorkflowInfo.workflowId` |
| `run_id` | Temporal `WorkflowInfo.runId` |
| `workflow_type` | Workflow interface simple name |
| `activity_type` | Activity method name (activity spans only) |
| `task_queue` | Worker poll queue |
| `trace_id` / `span_id` | OTel context (W3C traceparent on outbound HTTP) |

Use **OpenTelemetry Baggage** or span attributes (prefer attributes for exporter compatibility). Log JSON must include `trace_id` when span context is active.

## Alternatives Considered

- **Logs only, no traces** — Insufficient for latency debugging across activities.
- **workflow_id as metric label** — Rejected: unbounded cardinality (see ADR-002 metrics rules).

## Consequences

### Positive
- End-to-end drill-down: metric anomaly → trace → logs
- Temporal UI run can be searched in Loki via `workflow_id`

### Negative
- Must guard against PII in workflow IDs if production uses user identifiers

## Trade-offs
Rich correlation is prioritized; **workflow_type** used for metrics, **workflow_id** for traces/logs only.
