# Phase 7: Operations & On-Call — Context

**Gathered:** 2026-05-19

<domain>

## Phase boundary

Deliver **OPS-01**, **OPS-02**, **OPS-03**:

- Complete `docs/OPERATIONS.md` for operator workflows.
- Example Prometheus alert rules (not necessarily wired to PagerDuty).
- Manual validation record for each failure mode.

Maps diagram bottom: **Operations / On-Call** receiving **end-to-end visibility** from Grafana.

**Out of scope:** 24/7 on-call rotation, production PagerDuty integration.

</domain>

<decisions>

- **D-01:** Four runbook sections: stuck workflow, missing traces, prometheus down, loki disk.
- **D-02:** Alerts in `deploy/prometheus/alerts.yml` as examples; `promtool check rules` in CI.
- **D-03:** Validation in `07-VERIFICATION.md` with date + environment (compose) + outcome table.
- **D-04:** Cross-link Temporal UI, Jaeger (optional), Grafana Explore in each runbook.

</decisions>

<canonical_refs>

- `.planning/REQUIREMENTS.md` — OPS-01..03
- `.planning/phases/06-lgtm/06-VERIFICATION.md`

</canonical_refs>
