# Threat / failure notes — ENG-I-001

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| Copyright / content leakage | mitigate | Leadership artifacts are original portfolio paraphrases; no copyrighted book text |
| Authn/z | accept (demo) | Demo binds locally; production would require OIDC/RBAC |
| Soft-skill kit confusion | mitigate | Folder proves code+leadership together; Phase 7 owns mentoring kits |

## Failure modes
- Missing `artifacts/leadership/*.md` → `hybrid_ic` false (ADR-only / leadership-only is insufficient).
- Service without leadership files fails the hybrid proof even if `/healthz` is green.
