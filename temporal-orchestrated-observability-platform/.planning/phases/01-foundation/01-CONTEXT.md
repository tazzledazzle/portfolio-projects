# Phase 1: Foundation — Context

**Gathered:** 2026-05-19  
**Status:** Ready for `/gsd-execute-phase 1`

<domain>

## Phase boundary

Deliver **FOUND-01**, **FOUND-02**, **FOUND-03**:

- Docker Compose stack for Temporal + LGTM components (services may be minimal/no-op configs initially but must **start healthy**).
- Gradle multi-module skeleton: `worker`, `workflows`, `starter` (empty main classes OK).
- `scripts/smoke.sh` validates health endpoints.
- GitHub Actions CI: `./gradlew build test`.

**Out of scope:** Business workflows, OTel wiring, provisioned Grafana dashboards (later phases).

</domain>

<decisions>

## Implementation decisions

- **D-01:** Single `deploy/docker-compose.yml` with profiles: `core` (Temporal, Postgres if needed), `observability` (Prometheus, Loki, Tempo, Grafana, collector), `dev` (Jaeger — optional until Phase 5).
- **D-02:** `.env.example` documents ports; `.env` gitignored.
- **D-03:** Gradle root `settings.gradle.kts` includes `:worker`, `:workflows`, `:starter`.
- **D-04:** Smoke script uses `curl` + retry loop; exits non-zero on timeout.
- **D-05:** README quick start matches actual compose service names.

</decisions>

<canonical_refs>

- `docs/ARCHITECTURE.md` — repository layout target
- `docs/adr/0005-compose-local-lgtm.md`
- `.planning/REQUIREMENTS.md` — FOUND-01..03
- `.planning/research/STACK.md`

</canonical_refs>
