# Error Budget Burn Runbook

Linked from Prometheus alerts via `runbook_url: docs/runbooks/error-budget-burn.md`.

## SLOs

| SLO | Target | Good event | Bad event |
|-----|--------|------------|-----------|
| Availability | 99.9% | HTTP responses with status not in 5xx | 5xx responses |
| Latency | 99% under 500ms | Requests with duration ≤ 500ms | Requests slower than 500ms |

4xx responses (including payments HTTP 409 illegal escrow transitions) do **not** burn the availability budget.

## Symptoms

- Alert `AvailabilityErrorBudgetBurnFast` / `LatencyErrorBudgetBurnFast` (14.4x, 5m+1h windows)
- Alert `AvailabilityErrorBudgetBurnSlow` / `LatencyErrorBudgetBurnSlow` (6x, 30m+6h windows)
- Grafana **SLO / Error Budget** dashboard shows budget remaining dropping

## PromQL to check

```promql
# Per-service availability (5m)
sli:http_requests:availability:ratio_rate5m

# Per-service latency SLI (5m)
sli:http_requests:latency_le_500ms:ratio_rate5m

# Error budget remaining
slo:availability:error_budget_remaining:ratio
slo:latency:error_budget_remaining:ratio

# Raw 5xx rate
sum by (service) (rate(http_server_requests_seconds_count{status=~"5.."}[5m]))
```

In Grafana Explore → Prometheus, or:

```bash
kubectl -n c2c port-forward svc/prometheus 9090:9090
# open http://localhost:9090
```

## Mitigation steps

1. **Identify the burning service** — check alert `service` label and Platform Overview dashboard.
2. **Check recent deploys / pod restarts** — `kubectl -n c2c get pods`, Events, Kubernetes Saturation dashboard.
3. **Availability burn** — inspect 5xx routes in the service deep-dive dashboard; check dependency health (Postgres, Redis, OpenSearch, Redpanda).
4. **Latency burn** — look at p95/p99 on the service dashboard; check DB pool saturation, OpenSearch indexing lag, Kafka consumer lag (Event Pipeline dashboard).
5. **Payments-specific** — escrow `server_error` counters vs expected 409s; illegal transitions must stay 409.
6. **Messaging-specific** — `messaging_ws_*` failure counters and Redis presence; see Messaging + Redis dashboard.
7. **Correlate traces** — Distributed Tracing dashboard / Tempo; follow `trace_id` from Loki JSON logs.
8. **Stabilize** — scale/restart only the failing Deployment; avoid cluster-wide restarts on kind.
9. **Validate recovery** — wait for fast-burn alert to clear; confirm `slo:*:error_budget_remaining:ratio` trending up.

## Local demo notes

- Prometheus retention is short (~2d) on kind; treat 1d windows as the long window proxy for 30d.
- Grafana anonymous Viewer is enabled for localhost demos only — not for production exposure.
