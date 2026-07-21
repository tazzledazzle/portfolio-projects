# Threat / failure notes — ENG-E-008

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| T-5-06 Information Disclosure via StructuredError | mitigate | Errors expose only `code` + `message`; no stack traces, secrets, or internal paths |
| Authn/z | accept (demo) | Demo binds locally; production would require OIDC/RBAC |
| Multi-tenant isolation | deferred | Quotas owned by other slices before shared exposure |
| Supply chain | deferred | Pin base images; signing covered in supply-chain slices |

## Failure modes
- Invalid `/v1/echo` bodies return structured JSON errors (400), not raw panics.
- Missing `X-Request-ID` generates a server-side ID so traces remain correlatable.
