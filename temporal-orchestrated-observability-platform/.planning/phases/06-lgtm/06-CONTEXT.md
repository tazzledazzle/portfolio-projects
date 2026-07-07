# Phase 6: LGTM & Grafana — Context

**Gathered:** 2026-05-19

<domain>

## Phase boundary

Deliver **LGTM-01** through **LGTM-04**:

- **Loki** ingests worker JSON logs via Promtail.
- **Tempo** receives OTLP traces (collector export migration per ADR-003).
- **Grafana** provisioned datasources + 3 dashboards.
- Log → trace correlation via derived fields.

Maps diagram: Loki/Tempo → Grafana Dashboards → **end-to-end visibility**.

**Out of scope:** Production alerting routing (Phase 7); Jaeger removal optional.

</domain>

<decisions>

- **D-01:** Collector adds `otlp` exporter to Tempo; Jaeger exporter kept behind profile `jaeger` or removed from default.
- **D-02:** Promtail static config scrapes Docker logs or host file `logs/worker.log` — prefer Docker logging driver labels.
- **D-03:** Derived field in Loki: `trace_id` → Tempo datasource link.
- **D-04:** Dashboards as JSON in `deploy/grafana/dashboards/` provisioned via sidecar/watch.
- **D-05:** Grafana anonymous admin disabled; login from `.env`.

</decisions>

<canonical_refs>

- `docs/adr/0003-jaeger-dev-tempo-lgtm.md`
- `docs/adr/0005-compose-local-lgtm.md`

</canonical_refs>
