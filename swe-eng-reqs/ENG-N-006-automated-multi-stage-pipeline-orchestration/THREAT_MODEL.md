# Threat / failure notes — ENG-N-006

## T-4-04 Stage advance without gate (Elevation of Privilege)

- **Mitigation (MVP):** `Tick` ALWAYS calls `GateEvaluator.Evaluate` before any
  stage advance; a `deny` decision blocks advancement and marks the
  orchestration `blocked`. The live `/v1/demo` proves `blocked_on_deny` and
  `gate_required` so no path can advance a release stage without a gate check.
- **Deferred:** Production OIDC/RBAC on the orchestration API → Phase 5 (ENG-H-004).

## Boundary notes (D-09)

- N-006 orchestrates release/environment stages only. It must NOT reproduce CI
  DAG semantics (`lint`/`unit`/`build`/`publish`) — proof vocabulary is
  `stages_advanced` / `blocked_on_deny` / `gate_required`, never CI job names.
- The `GateEvaluator` is embedded (stub) — no HTTP coupling to ENG-N-005 (D-05).

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Input validation: stage and SLO IDs reject path traversal (`..`, `/`, `\`) and
  use a conservative alphabet.
- Failure modes: unknown orchestration IDs return 404; concurrent `Tick` is
  mutex-safe (validated under `-race`).
