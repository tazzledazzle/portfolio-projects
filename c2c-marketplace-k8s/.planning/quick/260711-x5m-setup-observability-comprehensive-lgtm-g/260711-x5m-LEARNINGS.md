---
phase: 260711-x5m
phase_name: "Setup Observability (LGTM)"
project: "C2C Marketplace K8s"
generated: "2026-07-12T09:00:00Z"
counts:
  decisions: 6
  lessons: 4
  patterns: 5
  surprises: 3
missing_artifacts:
  - "VERIFICATION.md"
  - "UAT.md"
---

# Phase 260711-x5m Learnings: Setup Observability (LGTM)

## Decisions

### Prometheus instead of Mimir for kind
Use Prometheus as the LGTM metrics backend on kind (Mimir-compatible PromQL for dashboards), not Mimir itself.

**Rationale:** Kind memory limits make a full Mimir deployment a poor fit for a portfolio demo; PromQL dashboards remain portable.
**Source:** 260711-x5m-PLAN.md (D-01); 260711-x5m-SUMMARY.md (key-decisions)

---

### Discrete LGTM + Alloy manifests (no all-in-one image)
Deploy Prometheus, Loki, Tempo, Grafana, and Alloy as separate single-replica manifests under `infra/k8s/observability/`.

**Rationale:** Avoid `grafana/otel-lgtm` all-in-one so the portfolio walkthrough shows real service topology and deploy wiring.
**Source:** 260711-x5m-PLAN.md (D-01/D-02); 260711-x5m-SUMMARY.md

---

### Shared `common/` Observability helper
Centralize Micrometer, OTLP, and `/metrics` wiring in `common/.../Observability.kt` instead of copying into each service.

**Rationale:** Four identical instrumentation paths would drift; plan explicitly allowed a shared helper as an acceptable deviation.
**Source:** 260711-x5m-SUMMARY.md (Decisions Made; Deviation 1)

---

### Manual OTel server spans over alpha Ktor instrumentation
Implement W3C-propagating server span interceptors in the shared helper rather than depending on `opentelemetry-ktor-2.0`.

**Rationale:** The alpha Ktor OTel artifact was unresolved/unstable with the chosen versions and blocked compile.
**Source:** 260711-x5m-SUMMARY.md (Deviation 2)

---

### Dashboard ConfigMap materialized at deploy time
Generate `grafana-dashboards` via `kubectl create configmap --from-file=...` in `deploy-kind.sh` rather than embedding twelve JSON files in a giant YAML.

**Rationale:** Keeps manifests readable while still ConfigMap-mounting provisioned dashboards.
**Source:** 260711-x5m-SUMMARY.md (Deviation 3); 260711-x5m-PLAN.md (Task 3)

---

### Anonymous Grafana Viewer + unauthenticated cluster-internal OTLP
Accept anonymous Viewer on localhost:3000 and unauthenticated Alloy OTLP inside the cluster for the kind demo trust model.

**Rationale:** Matches existing no-auth service posture; threats T-260711-01 and T-260711-04 explicitly accepted for local demo only.
**Source:** 260711-x5m-PLAN.md (threat_model); 260711-x5m-SUMMARY.md (Threat Flags)

---

## Lessons

### Micrometer 1.13 package rename bites imports
Micrometer relocated the Prometheus registry package to `io.micrometer.prometheusmetrics`, which broke first-pass imports during Task 2 compile.

**Context:** Instrumentation compile after adding micrometer-registry-prometheus.
**Source:** 260711-x5m-SUMMARY.md (Issues Encountered)

---

### Alpha Ktor OTel artifacts are not dependable for a quick task
Relying on `opentelemetry-ktor-2.0` blocked the build; a thin manual span interceptor was enough for demo traces.

**Context:** Task 2 compile failure and auto-fix deviation.
**Source:** 260711-x5m-SUMMARY.md (Deviation 2)

---

### Infra dashboards can start with kube proxies
Postgres/Redis/OpenSearch health panels can use cadvisor/kube-state proxies without dedicated exporters and still satisfy the kind footprint goal.

**Context:** Known stubs / Next Phase Readiness — richer exporters deferred.
**Source:** 260711-x5m-SUMMARY.md (Known Stubs; Next Phase Readiness)

---

### Metric name contract must be shared across rules and dashboards
Recording-rule metric names (`sli:*` / `slo:*`) need an explicit contract (header in `26-recording-rules.yaml`) so dashboard PromQL stays aligned.

**Context:** PLAN key_links and Task 3 wiring requirement; SUMMARY patterns-established.
**Source:** 260711-x5m-PLAN.md; 260711-x5m-SUMMARY.md (patterns-established)

---

## Patterns

### prometheus.io scrape annotations + Alloy kubernetes SD
Annotate app pods with `prometheus.io/scrape|port|path` and let Alloy discover/scrape them into Prometheus.

**When to use:** Adding any new Ktor service to this stack on kind.
**Source:** 260711-x5m-PLAN.md (key_links); 260711-x5m-SUMMARY.md (patterns-established)

---

### OTEL_EXPORTER_OTLP_ENDPOINT → Alloy → Tempo
Point each Deployment at Alloy `:4318`; Alloy forwards OTLP to Tempo.

**When to use:** Trace export from apps without embedding Tempo client config in each service.
**Source:** 260711-x5m-SUMMARY.md (patterns-established)

---

### Compose profile for optional observability
Keep default `docker compose up -d` as infra-only; gate LGTM/Alloy behind `profiles: observability`.

**When to use:** Inner-loop infra must stay light while still offering a local LGTM path.
**Source:** 260711-x5m-PLAN.md (Task 1 / D-07)

---

### Low-cardinality Micrometer labels only
Label metrics by service, status class, and route template — never raw userId/listingId.

**When to use:** Any new custom counter/timer (escrow, WebSocket, indexing).
**Source:** 260711-x5m-PLAN.md (threat T-260711-03; Task 2 action)

---

### Multi-window error-budget burn alerts + runbook
Fast (~14.4x) and slow (~6x) burn alerts with `runbook_url` pointing at `docs/runbooks/error-budget-burn.md`.

**When to use:** SLO-backed alerting for availability budgets on the four services.
**Source:** 260711-x5m-PLAN.md (D-06); 260711-x5m-SUMMARY.md

---

## Surprises

### Shared helper emerged as the cleanest path mid-execution
Plan listed per-service files; execution found duplication risk high enough to justify a `common/` module deviation without scope creep.

**Impact:** Cleaner instrumentation, one place to fix Micrometer/OTel issues, three atomic commits still mapped cleanly to tasks.
**Source:** 260711-x5m-SUMMARY.md (Deviation 1)

---

### Full stack landed in ~5 minutes of execution time
~50 files across three tasks completed in roughly five minutes once planned.

**Impact:** Confirms the 3-task quick-task split (stack → instrument → dashboards) was the right grain for this scope.
**Source:** 260711-x5m-SUMMARY.md (Performance)

---

### Twelve JSON dashboards do not belong inline in k8s YAML
Embedding dashboard JSON in manifests was impractical; deploy-time ConfigMap creation was required to unblock Task 3.

**Impact:** `deploy-kind.sh` became part of the Grafana provisioning path, not just apply-order wiring.
**Source:** 260711-x5m-SUMMARY.md (Deviation 3)
