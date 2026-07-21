# Threat / failure notes — ENG-N-007

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Input boundary: workload IDs reject traversal separators; replicas and image are validated.
- T-4-07 (finalizer never cleared / DoS): delete retains the object, while an
  explicit finalize endpoint clears the cleanup finalizer and removes it.
  Reconcile is idempotent, and `/v1/demo` proves `finalizer_cleared`.
- Concurrency: the in-memory controller protects create, reconcile, delete,
  finalize, and read paths with an RWMutex and is tested with `-race`.
- Scope: local in-memory operator proof only; production authorization,
  admission policy, and external cleanup retries remain deployment concerns.
