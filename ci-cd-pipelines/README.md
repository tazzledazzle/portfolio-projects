### CI/CD Pipelines (6)

**6. Pipeline telemetry exporter** — A GitHub Actions / GitLab CI plugin that emits span data (job start, step duration, failure reason) as OpenTelemetry traces, enabling end-to-end pipeline latency analysis alongside application traces.

**7. Flaky pipeline gate** — A service that tracks pass/fail history per job, computes a flakiness score (failure rate over a rolling window, excluding failures correlated with code changes), and blocks merges only when a test's failure is non-flaky.

**8. Release lead time calculator** — Pulls PR merge timestamps and deployment event timestamps from GitHub + your deployment platform, computes cycle time and lead time distributions per team/service, and exposes them as a DORA metrics API.

**9. Canary deployment controller** — A Kubernetes controller that progressively shifts traffic from a stable release to a canary using weighted Ingress rules, monitors error rate and p99 latency via Prometheus, and auto-rolls back when SLOs breach.

**10. Self-service pipeline template engine** — A Backstage scaffolder action that generates a complete CI/CD pipeline (GitHub Actions YAML) from a service template manifest — including lint, test, build, publish, and deploy stages — wired to the team's conventions.

**11. Pipeline cost analyzer** — Aggregates GitHub Actions billing data (minutes per runner type per workflow) and attributes cost to team/service/job, surfacing the highest-cost jobs and recommending optimizations (caching, parallelism, runner downsizing).