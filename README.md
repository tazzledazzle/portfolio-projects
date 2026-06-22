# Portfolio Projects

![CI](https://github.com/tazzledazzle/portfolio-projects/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/tazzledazzle/portfolio-projects/actions/workflows/release-please.yml/badge.svg)
![CodeQL](https://github.com/tazzledazzle/portfolio-projects/actions/workflows/codeql.yml/badge.svg)

## Portfolio Projects

<!-- PROJECTS_TABLE_START -->
| Name | Problem | Stack | Highlights | Status | Link |
|---|---|---|---|---|---|
| WebSocket Chat (FastAPI) | Low-latency chat demo | Python, FastAPI, Redis | Auto reconnection, Metrics | stable | [ws-chat-fast](./ws-chat-fast) |
| Project Generator | Scaffold new projects easily | Python, Jinja2, Typer | CLI UX, Template system | beta | [projgen](./projgen) |
| RabbitMQ Demo | Message queue patterns | Kotlin, RabbitMQ, Gradle | Pub/Sub, Work queues | stable | [rabbit-mq](./rabbit-mq) |
| Online Bookstore | E-commerce REST API demo | Python, FastAPI, PostgreSQL | REST API, OAuth2, WebSockets | stable | [online-bookstore](./online-bookstore) |
| Bazel Multi-Build | Polyglot build system | Bazel, Java, TypeScript | Build optimization, Cross-language | beta | [bazel-multibuild](./bazel-multibuild) |
| Modular JVM Build | Gradle multi-module Spring Boot | Kotlin, Spring Boot, Gradle | Layered modules, Health API | stable | [modular-jvm-build](./modular-jvm-build) |
| REST API Test Demo | Production-style FastAPI with tests | Python, FastAPI, Docker | pytest, Coverage, CI | stable | [rest-api-test-demo](./rest-api-test-demo) |
| Workflow API Demo | API to Redis queue to worker | Python, FastAPI, Redis, PostgreSQL | Async jobs, Docker Compose | stable | [workflow-api-demo](./workflow-api-demo) |
| OpenTelemetry Demo Stack | End-to-end distributed tracing | Python, OpenTelemetry, Docker | OTel Collector, Trace propagation | stable | [otel-demo-stack](./otel-demo-stack) |
| Observability Stack | Prometheus and Grafana locally | Docker, Prometheus, Grafana | Metrics, Dashboards | beta | [observability](./observability) |
| Platform Audit Template | SRE audit scripts and templates | Python, Bash, GCP | Billing analysis, Runbooks | beta | [platform-audit-template](./platform-audit-template) |
| AI Code Assistant | LLM-assisted test generation | Python, FastAPI, OpenAI | CLI, unit/integration/e2e tests | stable | [ai-best-practices-examples/ai-code-assistant](./ai-best-practices-examples/ai-code-assistant) |
| Knowledge QA System | RAG over documents | Python, LangChain, ChromaDB | Ingestion, Vector search | stable | [ai-best-practices-examples/knowledge-qa-system](./ai-best-practices-examples/knowledge-qa-system) |
| Chat AI | Conversational agent with tools | Python, FastAPI, LangChain | Memory, Tool calling | stable | [ai-best-practices-examples/chat-ai](./ai-best-practices-examples/chat-ai) |
| Domain Expert AI | Fine-tuning and eval pipeline | Python, Transformers, PEFT | Training, Guardrails | beta | [ai-best-practices-examples/domain-expert-ai](./ai-best-practices-examples/domain-expert-ai) |
| AI Image/Video Generator | ComfyUI-backed media generation | Python, Gradio, ComfyUI | Image gen, Live smoke tests | beta | [ai-best-practices-examples/ai-image-video-generator](./ai-best-practices-examples/ai-image-video-generator) |
| Canary Deployment Controller | Progressive traffic shifting with rollback | Python, Kubernetes, Prometheus | SLO gates, Auto rollback | beta | [ci-cd-pipelines/canary-deployment-controller](./ci-cd-pipelines/canary-deployment-controller) |
| Flaky Pipeline Gate | Flakiness-aware merge blocking | Python, CI analytics | Rolling window, Flake scoring | beta | [ci-cd-pipelines/flaky-pipeline-gate](./ci-cd-pipelines/flaky-pipeline-gate) |
| Release Lead Time Calculator | DORA lead time metrics API | Python, GitHub API | Cycle time, DORA | beta | [ci-cd-pipelines/release-lead-time-calculator](./ci-cd-pipelines/release-lead-time-calculator) |
| Pipeline Telemetry Exporter | CI spans as OpenTelemetry traces | Python, OpenTelemetry | Job spans, Step duration | beta | [ci-cd-pipelines/pipeline-telemetry-exporter](./ci-cd-pipelines/pipeline-telemetry-exporter) |
| Pipeline Template Engine | Generate CI YAML from manifests | Python, Jinja2, Backstage | Scaffolding, Conventions | beta | [ci-cd-pipelines/self-service-pipeline-template-engine](./ci-cd-pipelines/self-service-pipeline-template-engine) |
| Pipeline Cost Analyzer | Attribute CI runner spend to teams | Python, GitHub Actions | Cost attribution, Optimization | beta | [ci-cd-pipelines/pipeline-cost-analyzer](./ci-cd-pipelines/pipeline-cost-analyzer) |
| Kotlin Custom Detekt Rules | Team-specific static analysis rules | Kotlin, Detekt, Gradle | Custom rules, RuleSetProvider | beta | [c0de-quality-and-analysis/kotlin-custom-detekt-rules-library](./c0de-quality-and-analysis/kotlin-custom-detekt-rules-library) |
| License Compliance Scanner | SPDX license policy enforcement | Python, Gradle, SPDX | SBOM, Policy engine | beta | [c0de-quality-and-analysis/license-compliance-scanner](./c0de-quality-and-analysis/license-compliance-scanner) |
| Dead Code Surface Reporter | Prioritize unreachable stale code | Python, Git, Call graph | Blame recency, Scoring | beta | [c0de-quality-and-analysis/dead-code-surface-reporter](./c0de-quality-and-analysis/dead-code-surface-reporter) |
| API Breaking Change Detector | OpenAPI compatibility merge gate | Python, OpenAPI | Diff, Merge gate | beta | [c0de-quality-and-analysis/api-breaking-change-detector](./c0de-quality-and-analysis/api-breaking-change-detector) |
| Security Hotspot Annotator | Semgrep findings on PR comments | Python, Semgrep, GitHub | CWE mapping, PR annotations | beta | [c0de-quality-and-analysis/security-hotspot-annotator](./c0de-quality-and-analysis/security-hotspot-annotator) |
| Devcontainer Feature Library | Composable local dev stack features | Shell, Dev Containers, Docker | Kafka, Keycloak, Postgres | beta | [dev-env/devcontainer-feature-library](./dev-env/devcontainer-feature-library) |
| Environment Drift Detector | Detect local toolchain drift | Python, YAML | Version compare, Fix scripts | beta | [dev-env/environment-drift-detector](./dev-env/environment-drift-detector) |
| Local Service Mesh | K8s-like routing on Docker Compose | Docker, Caddy, Compose | Gateway, Service routing | beta | [dev-env/local-service-mesh](./dev-env/local-service-mesh) |
| Onboarding Automation CLI | Automate new-hire environment setup | Kotlin, Gradle, CLI | Repo clone, Health checks | beta | [dev-env/onboarding-automation-cli](./dev-env/onboarding-automation-cli) |
| Remote Dev Environment Orchestrator | Ephemeral cloud dev environments | Python, Temporal | TTL teardown, Provisioning | beta | [dev-env/remote-dev-environment-orchestrator](./dev-env/remote-dev-environment-orchestrator) |
| Developer Satisfaction Pulse | SPACE/NPS developer surveys | Python, FastAPI, PostgreSQL | Pulse surveys, Trends | beta | [dev-ex/developer-satisfaction-pulse-system](./dev-ex/developer-satisfaction-pulse-system) |
| Tooling Adoption Tracker | Measure tool rollout adoption | TypeScript, Node | Funnel metrics, Telemetry | beta | [dev-ex/tooling-adoption-tracker](./dev-ex/tooling-adoption-tracker) |
| Inner Loop Friction Scorer | Composite developer friction score | Python, CI telemetry | Friction taxonomy, Scoring | beta | [dev-ex/inner-loop-friction-scorer](./dev-ex/inner-loop-friction-scorer) |
| Platform Changelog Generator | API diff changelogs and migrations | Python, AST | Breaking changes, Rewriters | beta | [dev-ex/platform-changelog-migration-generator](./dev-ex/platform-changelog-migration-generator) |

<!-- PROJECTS_TABLE_END -->

Learn how each project aligns with key competencies in [PORTFOLIO.md](PORTFOLIO.md).

## Legacy Design Documents

Ten aspirational design documents live under `src/main/resources/`:

- dd-build-system-expertise.md
- dd-ci-cd-pipeline-with-flaky-test-detection.md
- dd-cloud-native-microservices.md
- dd-full-stack-sample-app.md
- dd-developer-productivity-tooling.md
- dd-cross-plat-nortarization-tool.md
- dd-gitops-starter-kit.md
- dd-open-source-plugin-contribution.md
- dd-performance-benchmarking-n-tuning.md
- dd-streaming-data-demo.md
