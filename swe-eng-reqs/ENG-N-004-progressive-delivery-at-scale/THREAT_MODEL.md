# Threat / failure notes — ENG-N-004

## T-4-05 Criteria bypass on promote (Elevation of Privilege)

- **Mitigation (MVP):** `Promote` refuses to advance the environment when
  `criteria_passed` is false. The verdict is computed server-side by `Evaluate`
  from observed metrics — clients cannot supply their own pass/fail. After each
  promotion the verdict resets, forcing a fresh evaluation for the next
  environment. `/v1/demo` proves `promote_blocked_when_failing` and
  `auto_promoted`.
- **Deferred:** Production OIDC/RBAC on the plan API → Phase 5 (ENG-H-004).

## Boundary notes (D-03)

- N-004 owns multi-environment PD + promotion criteria only. Single-canary
  weight-step internals belong to ENG-E-024 and are not reproduced here.

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Input validation: environment IDs reject path traversal (`..`, `/`, `\`).
- Failure modes: unknown plan IDs return 404; concurrent Evaluate/Promote is
  mutex-safe (validated under `-race`); promotion never advances past the final
  environment.
