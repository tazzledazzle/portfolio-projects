# Threat / failure notes — ENG-N-008

## T-4-06 Profile / workload ID injection (Tampering)

- **Mitigation (MVP):** profile names and workload IDs are validated with a
  conservative alphabet (`^[a-z0-9][a-z0-9-]*$`) that rejects path traversal
  (`..`, `/`, `\`). `Schedule` refuses unknown profiles so a workload can only
  be placed onto explicitly registered targets.
- **Deferred:** Production OIDC/RBAC on the profile API → Phase 5 (ENG-H-004).

## Boundary notes (D-03 / D-11)

- N-008 owns the profile abstraction and multi-profile scheduling only.
  Profiles: `k8s-standard`, `k8s-gpu`, `vm-bake`. Canary weights, burn-rate
  gates, and finalizers are out of scope.

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Failure modes: unknown workloads return 404; unknown profiles return 400;
  concurrent Schedule is mutex-safe (validated under `-race`).
