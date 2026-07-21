# Threat / failure notes — ENG-E-007

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; stdlib only (no new modules) (T-5-SC).
- Failure modes: missing ADR dir fails startup; demo should not invent decisions.
- Boundary: ADR + thin skeleton only — no Phase 7 soft-skill mentoring kit content;
  no copyrighted book excerpts (paraphrase catalogs only).
