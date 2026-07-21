# Threat / failure notes — ENG-E-025

- **T-3-12 stale Ready status:** reconcile atomically sets `Expired=True`,
  `Ready=False`, and `reclaimed=true` when TTL elapses.
- **T-3-13 unsafe IDs:** IDs reject path separators and `..`; TTL must be
  positive.
- POST bodies are capped at 1 MiB. The tick endpoint is a local logical-clock
  helper and must not be publicly exposed without authorization.
- Production controllers require API-server RBAC, admission policy, tenant
  quotas, and idempotent infrastructure cleanup. This phase gate is in-memory.
