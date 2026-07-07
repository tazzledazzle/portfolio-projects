# Phase 2: Temporal Orchestration — Context

**Gathered:** 2026-05-19  
**Status:** Ready after Phase 1 verification

<domain>

## Phase boundary

Deliver **TEMP-01**, **TEMP-02**, **TEMP-03**:

- Kotlin worker registers workflow and activity implementations.
- **PingWorkflow** (single no-op activity) runs end-to-end.
- Temporal UI shows completed execution.
- Graceful shutdown documented and tested manually.

**Out of scope:** AI business logic (Phase 3), OTel (Phase 4).

</domain>

<decisions>

- **D-01:** Namespace `default`; task queue `ai-workflows`.
- **D-02:** Connection via env `TEMPORAL_HOST` default `localhost:7233`.
- **D-03:** Use Temporal Kotlin DSL / Java SDK patterns consistent with official samples.
- **D-04:** `starter` module provides `./gradlew :starter:run --args="ping"`.

</decisions>

<canonical_refs>

- `docs/adr/0001-temporal-for-ai-orchestration.md`
- `.planning/phases/01-foundation/01-VERIFICATION.md` (when exists)

</canonical_refs>
