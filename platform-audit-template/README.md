# Platform Audit Template

**Problem:** Teams need a repeatable way to assess service and platform maturity (deployments, observability, security, testing, docs) without ad-hoc checklists.

**Solution:** This template provides a structured audit framework, runbook examples, and SRE communications checklists you can adapt for your organization.

**What this demonstrates:** Reusable SRE/audit artifacts and runbook documentation—drawn from experience conducting data platform audits across many services and writing SRE-facing documentation.

---

## Contents

| Artifact | Description |
|----------|-------------|
| [Service/Platform Audit Template](docs/audit-template.md) | Checklist and structure for assessing service maturity (deployments, observability, security, testing, docs). |
| [Runbook: Verify OTel Pipeline](docs/runbook-verify-otel-pipeline.md) | Step-by-step runbook to verify OpenTelemetry traces and metrics are flowing. |
| [Runbook: Cost/Billing Diagnostic](docs/runbook-billing-diagnostic.md) | Pattern for safe, read-only cost/billing diagnostics (e.g. GCP). |
| [SRE Communications Checklist](docs/sre-communications-checklist.md) | Time zones, handoffs, incident comms, and leaving-desk protocols. |
| [Telemetry-Driven Rollout Playbook](docs/rollout-playbook.md) | Phases (pre-release, canary, staged), metrics to watch, rollback, and documentation. |
| [Diagnostic Script Pattern](scripts/README.md) | Safe, read-only diagnostic script pattern with an example (GCP billing summary). |

---

## How to use

1. **Audit:** Copy `docs/audit-template.md` (or the checklist sections) into your wiki or repo; fill per service or platform.
2. **Runbooks:** Adapt the runbooks to your stack (OTel endpoints, cloud provider, tooling).
3. **Comms:** Use the SRE communications checklist for onboarding or team norms.
4. **Rollouts:** Use the rollout playbook when planning canary or staged releases.
5. **Scripts:** Use `scripts/` as a pattern for read-only diagnostics; never commit credentials.

---

## License

MIT. Use and adapt freely.
