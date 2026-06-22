### Developer Experience (DevEx) (4)

**37. Developer satisfaction pulse survey system** — A lightweight survey tool (weekly 3-question pulse + quarterly SPACE framework survey) that stores results in Postgres, computes rolling NPS and friction scores per team, and renders trend dashboards.

**38. Tooling adoption tracker** — Instruments IDE plugin usage, CLI invocations, and portal page views via lightweight telemetry (opt-in), produces adoption funnels per tool launch, and surfaces which teams are not yet using a new platform capability.

**39. Inner loop friction scorer** — Defines a friction taxonomy (environment setup time, build time, test time, PR review latency, deploy time) and computes a composite friction score per team from CI telemetry, Git events, and self-reported data.

**40. Platform changelog and migration guide generator** — A CI workflow that, on each platform library release, diffs the public API, generates a structured changelog (breaking changes, deprecations, new features) and — for breaking changes — produces a Kotlin/Python migration script using AST transforms.