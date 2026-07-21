# Threat / failure notes — ENG-H-002

## Trust boundary
client → cost / builds API (untrusted build cost inputs)

## Mitigations (T-5-21 Spoofing / forged savings)
- `cost_per_build_usd` and `cache_savings_pct` are **server-computed** from recorded builds
- Client-supplied `cost_usd` / `savings_pct` / `cache_savings_pct` fields are ignored
- Accept only `duration_sec`, `cpu_cores`, `memory_gb`, `cache_hit`
- Reject non-positive duration/CPU/memory
- Body size limited via `http.MaxBytesReader` (1 MiB)

## Notes
- Authn/z: demo binds locally; production would gate cost APIs
- Does not own CAS digest store (N-010) or tenant quotas (I-004)
- Stdlib only (T-5-SC)
