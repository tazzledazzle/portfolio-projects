# Threat / failure notes — ENG-I-006

## Trust boundary
client → self-service request API (untrusted kind/summary)

## Mitigations (T-5-02)
- **Tampering:** `ticket_removed` and before/after ticket counts are **server-computed**; clients cannot override `ticket_removed` on submit.
- Local-only mutate via `demo-local` on `:18606`; production would require authn.
- Request body limited with `http.MaxBytesReader` (1 MiB).

## Out of scope
- Full IDP project/pipeline/environment catalog (ENG-E-005)
- Product SLAs / golden-path (ENG-I-002)
