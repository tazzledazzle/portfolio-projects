# Platform / Service Audit Template

Use this template to assess **service maturity** across deployments, observability, security, testing, and documentation. Score each area (e.g. 1–5 or Not applicable / Partial / Full), and capture notes and remediation items.

**Audit metadata**

| Field | Value |
|-------|--------|
| Service / component name | |
| Repo(s) | |
| Owner / team | |
| Date | |
| Auditor | |

---

## 1. Deployments & release

| Criterion | Score (1–5 or N/A / Partial / Full) | Notes |
|-----------|-------------------------------------|-------|
| Build is reproducible (e.g. pinned deps, versioned artifacts) | | |
| CI runs on every PR (build + test) | | |
| CD or release process is documented | | |
| Deployment is automated (e.g. Helm, K8s, Terraform) | | |
| Rollback procedure exists and is documented | | |
| Environment parity (dev/staging/prod) or documented drift | | |

**Remediation (deployments):**

- 

---

## 2. Observability

| Criterion | Score | Notes |
|-----------|-------|-------|
| Health/readiness endpoints exist and are used | | |
| Metrics exposed (e.g. Prometheus, OTel) | | |
| Logs structured and routable (e.g. JSON, level, correlation ID) | | |
| Traces available (e.g. OTel, Jaeger) for key flows | | |
| Dashboards or runbooks reference these signals | | |
| Alerts defined for critical failure modes | | |

**Remediation (observability):**

- 

---

## 3. Security

| Criterion | Score | Notes |
|-----------|-------|-------|
| Secrets not in repo (vault, env, or secret manager) | | |
| Least-privilege IAM / service accounts | | |
| Dependencies scanned (e.g. Snyk, Dependabot) | | |
| Network: egress/ingress constrained where applicable | | |
| Pod/container security (e.g. non-root, read-only fs where possible) | | |

**Remediation (security):**

- 

---

## 4. Testing

| Criterion | Score | Notes |
|-----------|-------|-------|
| Unit tests exist and run in CI | | |
| Integration tests for critical paths | | |
| E2E or smoke tests (optional but noted) | | |
| Test coverage or critical-path coverage documented | | |
| Flaky test policy or quarantine process | | |

**Remediation (testing):**

- 

---

## 5. Documentation

| Criterion | Score | Notes |
|-----------|-------|-------|
| README: how to run locally | | |
| README or ADR: architecture / design | | |
| Runbooks or playbooks for common operations | | |
| API contract (OpenAPI, GraphQL schema) or equivalent | | |
| On-call or escalation path documented | | |

**Remediation (documentation):**

- 

---

## Summary

**Overall maturity (1–5 or narrative):**

**Top 3 priorities:**

1. 
2. 
3. 

**Next review date:** 
