### Developer Environments (5)

**12. Devcontainer feature library** — A collection of composable devcontainer features (OTel collector sidecar, local Kafka, local Postgres with seed data, local Keycloak) that developers reference in `.devcontainer.json` to spin up a complete local stack.

**13. Environment drift detector** — Compares a developer's local tool versions (via a manifest: `jvm`, `node`, `terraform`, `kubectl`, etc.) against the team's canonical versions and generates a prescriptive fix script to close the gap.

**14. Local service mesh** — A Docker Compose + Caddy setup that routes traffic between locally running microservices using the same service names as Kubernetes, allowing developers to test service-to-service calls locally without a full cluster.

**15. Onboarding automation CLI** — A Kotlin CLI that automates new-hire environment setup: clones required repos, configures git remotes, sets up SSH keys, installs tool versions via `asdf`, and verifies the environment with a health check suite.

**16. Remote dev environment orchestrator** — A Temporal workflow that provisions ephemeral cloud dev environments (via Coder or Gitpod API), pre-populates them with repo state and seed data, enforces TTL-based teardown, and tracks cost per developer.