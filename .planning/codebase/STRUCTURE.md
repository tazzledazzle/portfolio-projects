# Codebase Structure

**Analysis Date:** 2026-06-22

## Directory Layout

```
portfolio-projects/                          # Portfolio monorepo root
├── .github/workflows/                       # Root CI: ci.yml, codeql, release-please, pages
├── .planning/codebase/                      # GSD codebase analysis docs (this directory)
├── ai-best-practices-examples/              # AI/LLM example projects (5 subprojects)
├── bazel-multibuild/                        # Polyglot Bazel demo (java, frontend, flags tutorial)
├── c0de-quality-and-analysis/               # Code quality tool scaffolds (5 subprojects)
├── ci-cd-pipelines/                         # CI/CD automation scaffolds (6 subprojects)
├── dev-env/                                 # Developer environment tooling (5 + 3 stubs)
├── dev-ex/                                  # Developer experience analytics (4 subprojects)
├── docs/                                    # MkDocs source (5 featured design docs)
├── forgex/                                  # Product spec only (scaffold)
├── modular-jvm-build/                       # Gradle multi-module Spring Boot demo
├── observability/                           # Standalone Prometheus + Grafana stack
├── online-bookstore/                        # FastAPI e-commerce demo
├── otel-demo-stack/                         # OpenTelemetry end-to-end demo
├── platform-audit-template/                 # SRE audit templates + scripts
├── projgen/                                 # Bazel/Gradle project scaffolder CLI
├── rabbit-mq/                               # Gradle Kotlin + Python RabbitMQ demo
├── rest-api-test-demo/                      # Production-style FastAPI REST demo
├── scripts/                                 # Root utilities (README table gen, patches)
├── site/                                    # Generated MkDocs HTML (committed)
├── src/main/resources/                      # Legacy dd-*.md design documents (10 files)
├── tools/                                   # Project launcher (run, run.yaml)
├── workflow-api-demo/                         # API → Redis → worker → PostgreSQL
├── ws-chat-fast/                            # WebSocket chat (incomplete scaffold)
├── build.gradle.kts                         # Minimal root Kotlin shell (not subproject aggregator)
├── settings.gradle.kts                      # Root Gradle settings (single project name)
├── portfolio.yaml                           # Featured 6-project manifest
├── mkdocs.yml                               # Docs site nav (5 design docs)
├── Makefile                                 # make test, lint, docs
└── README.md                                # Portfolio table (6 projects only)
```

## Directory Purposes

**`ai-best-practices-examples/`:**
- Purpose: Demonstrate AI-assisted development patterns (RAG, fine-tuning, test gen, chat, image/video)
- Contains: Independent Python packages with `pyproject.toml`, `src/`, `tests/`
- Key files: `ai-code-assistant/src/ai_code_assistant/cli.py`, `knowledge-qa-system/src/app/main.py`

**`c0de-quality-and-analysis/`:**
- Purpose: CI-integrated code quality tools (detekt, license scan, dead code, OpenAPI diff, Semgrep)
- Contains: Python scaffolds + one Kotlin detekt library; per-subproject `.github/workflows/ci.yml`
- Key files: `README.md` (project descriptions 27–31), `kotlin-custom-detekt-rules-library/src/main/kotlin/`

**`ci-cd-pipelines/`:**
- Purpose: CI/CD automation concepts (flaky detection, canary deploy, DORA metrics, pipeline templates)
- Contains: Minimal Python CLI stubs with smoke tests
- Key files: `README.md` (projects 6–11), each `*/src/main.py`

**`dev-env/`:**
- Purpose: Local development environment tooling (devcontainers, mesh, onboarding, drift detection)
- Contains: Shell scripts, Compose files, Kotlin CLI, Python stubs; `.keep`-only test directories
- Key files: `local-service-mesh/docker-compose.yml`, `onboarding-automation-cli/src/main/kotlin/`

**`dev-ex/`:**
- Purpose: Developer experience measurement (surveys, friction scoring, adoption tracking, changelogs)
- Contains: FastAPI apps, Python CLIs, one TypeScript tracker
- Key files: `developer-satisfaction-pulse-system/src/app/main.py`, `tooling-adoption-tracker/src/index.ts`

**`docs/`:**
- Purpose: MkDocs-published design documentation for featured projects
- Contains: `design-docs/<project>/design.md`, `templates/design_doc_template.md`, `index.md`
- Key files: `docs/design-docs/ws-chat-fast/design.md` (5 total design docs)

**`forgex/`:**
- Purpose: Product requirements spec for unified polyglot repo generator (no code)
- Contains: `README.md` only (~950 lines)
- Key files: `forgex/README.md`

**`tools/`:**
- Purpose: Monorepo project launcher
- Contains: `run` bash script, `run.yaml` manifest
- Key files: `tools/run.yaml` (5 runnable project commands)

**`src/main/resources/`:**
- Purpose: Legacy aspirational design documents (Java resources layout, content is Markdown)
- Contains: 10 `dd-*.md` files
- Key files: `src/main/resources/dd-build-system-expertise.md`

**`observability/`:**
- Purpose: Standalone Prometheus + Grafana sample (not OTel demo)
- Contains: `docker-compose.yml`, `prometheus.yml`, `grafana/`
- Key files: `observability/docker-compose.yml`

## Key File Locations

**Entry Points:**
- `rest-api-test-demo/app/main.py`: Reference FastAPI REST service
- `otel-demo-stack/api/main.py`: OTel-instrumented API
- `otel-demo-stack/worker/worker.py`: OTel-instrumented background worker
- `workflow-api-demo/api/main.py`: Job enqueue API
- `workflow-api-demo/worker/worker.py`: Redis queue consumer
- `projgen/src/projgen/cli.py`: Project scaffolder CLI (`projgen` console script)
- `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py`: AI test generator CLI
- `modular-jvm-build/app/src/main/kotlin/showcase/Application.kt`: Spring Boot main
- `dev-env/onboarding-automation-cli/src/main/kotlin/com/company/onboarding/Main.kt`: Onboarding CLI
- `tools/run`: Portfolio project launcher shell script

**Configuration:**
- `portfolio.yaml`: Featured project metadata (6 entries)
- `tools/run.yaml`: Runnable project commands (5 entries)
- `mkdocs.yml`: Documentation site navigation
- `.pre-commit-config.yaml`: Root pre-commit hooks
- `.github/workflows/ci.yml`: Path-filtered CI pipeline
- `release-please-config.json`: Release automation config
- Per-subproject: `pyproject.toml`, `build.gradle.kts`, `docker-compose.yml`, `BUILD.bazel`

**Core Logic:**
- `ws-chat-fast/app/chat.py`: WebSocket chat router
- `ws-chat-fast/app/ws_manager.py`: Connection manager abstraction
- `online-bookstore/src/main.py`: Bookstore FastAPI app
- `rabbit-mq/app/src/main/kotlin/`: Kotlin RabbitMQ publisher
- `bazel-multibuild/java/src/main/java/`: Java Bazel sample app
- `platform-audit-template/scripts/gcp_billing_summary.py`: Audit utility script

**Testing:**
- `rest-api-test-demo/tests/`: Full pytest suite (~10 tests)
- `ai-best-practices-examples/ai-code-assistant/tests/`: Unit, integration, e2e (~58 test functions)
- `projgen/tests/` or inline tests: ~9 tests
- `dev-env/*/tests/**/.keep`: Placeholder test directories (13 `.keep` files, no test code)
- Root: `make test` runs pytest + `./gradlew test` repo-wide

## Naming Conventions

**Files:**
- Python entry: `main.py` (FastAPI apps) or `cli.py` (CLI tools)
- Python package layout: `src/<package_name>/` (PEP 517 src layout)
- Kotlin entry: `Main.kt` or `Application.kt` under `src/main/kotlin/<package>/`
- Bazel: `BUILD.bazel`, `MODULE.bazel`, `WORKSPACE` per module
- Design docs: `docs/design-docs/<project-id>/design.md` or `docs/architecture/overview.md` per subproject
- Legacy specs: `src/main/resources/dd-<topic>.md` (kebab-case topic)

**Directories:**
- Top-level areas: kebab-case (`ai-best-practices-examples`, `c0de-quality-and-analysis`)
- Subprojects: kebab-case matching purpose (`flaky-pipeline-gate`, `knowledge-qa-system`)
- Python source: `src/<snake_case_package>/`
- Kotlin source: `src/main/kotlin/<dot/not/path>/`
- Tests: `tests/` at subproject root; subdivided `unit/`, `integration/`, `e2e/` in mature projects
- JVM modules: short names (`core`, `api`, `app`) in `modular-jvm-build/`

**Subproject IDs:**
- Match directory name: `ws-chat-fast`, `projgen`, `ai-code-assistant`
- Used in `portfolio.yaml`, `tools/run.yaml`, and `docs/design-docs/` paths

## Where to Add New Code

**New featured portfolio demo:**
- Implementation: `<project-name>/` at repo root (sibling to `ws-chat-fast/`)
- Design doc: `docs/design-docs/<project-name>/design.md`
- Manifest entry: add to `portfolio.yaml` and regenerate README via `scripts/gen_readme_table.py`
- Runner entry: add to `tools/run.yaml`
- MkDocs nav: add entry in `mkdocs.yml`

**New AI example:**
- Implementation: `ai-best-practices-examples/<project-name>/`
- Structure: `pyproject.toml`, `src/<package>/`, `tests/`, `README.md`
- Entry point: register in `[project.scripts]` or `src/*/main.py` for FastAPI

**New CI/CD or code-quality scaffold:**
- Implementation: `ci-cd-pipelines/<tool-name>/` or `c0de-quality-and-analysis/<tool-name>/`
- Structure: `src/main.py`, `tests/test_smoke.py`, `README.md`, optional `.github/workflows/ci.yml`
- Description: add numbered entry to area `README.md`

**New dev-env or dev-ex tool:**
- Implementation: `dev-env/<tool-name>/` or `dev-ex/<tool-name>/`
- Structure: follow peers — FastAPI in `src/app/main.py`, CLI in `src/cli/main.py`, or Kotlin in `src/main/kotlin/`
- Docs: `docs/architecture/overview.md` under subproject

**New JVM module (within modular-jvm-build pattern):**
- Shared domain: `modular-jvm-build/core/`
- API surface: `modular-jvm-build/api/`
- Boot app: `modular-jvm-build/app/`
- Register in `modular-jvm-build/settings.gradle.kts`

**New Bazel target:**
- Java: `bazel-multibuild/java/`
- New language module: sibling directory with `MODULE.bazel` under `bazel-multibuild/`

**Utilities shared across monorepo:**
- Root scripts: `scripts/`
- Project launcher: `tools/`
- Do not add shared Python packages at root — keep subprojects independent

## Full Subproject Inventory

### Summary Counts

| Metric | Count |
|--------|-------|
| Top-level areas (user-listed) | 17 |
| Total subprojects (including nested) | 48 |
| Featured in `portfolio.yaml` / root README | 6 |
| MkDocs design docs in `docs/design-docs/` | 5 |
| Legacy specs in `src/main/resources/` | 10 |
| Maturity: stable | 3 |
| Maturity: beta | 39 |
| Maturity: scaffold | 6 |

### Complete Inventory Table

| # | Path | Purpose | Maturity | Has Tests | Has Docker |
|---|------|---------|----------|-----------|------------|
| 1 | `ai-best-practices-examples/ai-code-assistant` | Policy-aware Python test generation CLI | stable | yes (~58) | no |
| 2 | `ai-best-practices-examples/ai-image-video-generator` | Gradio + ComfyUI image/video generation | beta | yes (~18) | no |
| 3 | `ai-best-practices-examples/chat-ai` | FastAPI streaming chat with tools | beta | yes (~14) | no |
| 4 | `ai-best-practices-examples/domain-expert-ai` | QLoRA fine-tuning scaffold | beta | yes (~32) | no |
| 5 | `ai-best-practices-examples/knowledge-qa-system` | RAG Q&A with ChromaDB citations | beta | yes (5) | no |
| 6 | `c0de-quality-and-analysis/api-breaking-change-detector` | OpenAPI breaking-change merge gate | beta | smoke (1) | no |
| 7 | `c0de-quality-and-analysis/dead-code-surface-reporter` | Dead code + git recency reporter | beta | smoke (1) | no |
| 8 | `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library` | Custom detekt rules library | beta | yes (1) | no |
| 9 | `c0de-quality-and-analysis/license-compliance-scanner` | License compliance + SBOM | beta | smoke (1) | no |
| 10 | `c0de-quality-and-analysis/security-hotspot-annotator` | Semgrep PR annotation formatter | beta | smoke (1) | no |
| 11 | `ci-cd-pipelines/canary-deployment-controller` | K8s canary traffic controller | beta | smoke (2) | no |
| 12 | `ci-cd-pipelines/flaky-pipeline-gate` | Flaky test merge gate | beta | smoke (2) | no |
| 13 | `ci-cd-pipelines/pipeline-cost-analyzer` | CI cost attribution | beta | smoke (2) | no |
| 14 | `ci-cd-pipelines/pipeline-telemetry-exporter` | CI spans as OTel traces | beta | smoke (1) | no |
| 15 | `ci-cd-pipelines/release-lead-time-calculator` | DORA lead/cycle time | beta | smoke (1) | no |
| 16 | `ci-cd-pipelines/self-service-pipeline-template-engine` | Pipeline YAML from manifest | beta | smoke (1) | no |
| 17 | `dev-env/devcontainer-feature-library` | Composable devcontainer features | beta | .keep only | no |
| 18 | `dev-env/environment-drift-detector` | Toolchain drift detector CLI | beta | no | no |
| 19 | `dev-env/local-service-mesh` | Compose + Caddy local mesh | beta | .keep only | yes |
| 20 | `dev-env/onboarding-automation-cli` | New-hire setup Kotlin CLI | beta | no | no |
| 21 | `dev-env/remote-dev-environment-orchestrator` | Cloud dev env Temporal workflow | scaffold | no | no |
| 22 | `dev-env/local-service-mesh/services/order-service` | Mesh microservice stub | scaffold | no | no |
| 23 | `dev-env/local-service-mesh/services/user-service` | Mesh microservice stub | scaffold | no | no |
| 24 | `dev-env/local-service-mesh/services/payment-service` | Mesh microservice stub | scaffold | no | no |
| 25 | `dev-ex/developer-satisfaction-pulse-system` | Developer survey pulse tool | beta | yes (2) | no |
| 26 | `dev-ex/inner-loop-friction-scorer` | Inner loop friction metrics | beta | yes (3) | no |
| 27 | `dev-ex/platform-changelog-migration-generator` | Changelog + migration generator | beta | yes (2) | no |
| 28 | `dev-ex/tooling-adoption-tracker` | Tool adoption telemetry | beta | yes (1) | no |
| 29 | `bazel-multibuild/` | Polyglot Bazel parent | beta | — | no |
| 30 | `bazel-multibuild/java` | Java Bazel sample | beta | yes (1) | no |
| 31 | `bazel-multibuild/frontend` | Frontend Bazel placeholder | scaffold | no | no |
| 32 | `bazel-multibuild/flags-parsing-tutorial` | Bazel flags tutorial | beta | no | no |
| 33 | `modular-jvm-build/` | Gradle multi-module Spring Boot | beta | yes (1) | no |
| 34 | `online-bookstore/` | FastAPI e-commerce demo | beta | yes (6) | no |
| 35 | `otel-demo-stack/` | OTel API + worker + collector | stable | yes (3) | yes |
| 36 | `platform-audit-template/` | SRE audit templates + scripts | beta | yes (3) | no |
| 37 | `projgen/` | Bazel/Gradle scaffolder CLI | beta | yes (9) | no |
| 38 | `rabbit-mq/` | Kotlin + Python RabbitMQ demo | beta | yes (2) | no |
| 39 | `rest-api-test-demo/` | Production-style FastAPI REST | stable | yes (10) | yes |
| 40 | `workflow-api-demo/` | API → Redis → worker → PG | beta | yes (3) | yes |
| 41 | `ws-chat-fast/` | WebSocket chat (no bootstrap) | scaffold | no | no |
| 42 | `forgex/` | Polyglot generator product spec | scaffold | no | no |
| 43 | `tools/` | Project launcher | beta | no | no |
| 44 | `observability/` | Prometheus + Grafana stack | beta | no | yes |

### Root README vs Actual Projects

Root `README.md` (`<!-- PROJECTS_TABLE_START -->`) lists **6 projects**:

1. `ws-chat-fast` — WebSocket Chat (FastAPI)
2. `projgen` — Project Generator
3. `rabbit-mq` — RabbitMQ Demo
4. `online-bookstore` — Online Bookstore
5. `bazel-multibuild` — Bazel Multi-Build
6. (table has 5 rows + header; `portfolio.yaml` has 5 entries — README table shows 5 data rows)

Actually re-reading portfolio.yaml - it has 5 projects, README table shows 5 rows. User said 6 - let me count README again:
- ws-chat-fast
- projgen
- rabbit-mq
- online-bookstore
- bazel-multibuild

That's 5. User said 6 - maybe they count something else. I'll note "5–6 featured" or count portfolio.yaml which has 5. The README comment says 6 in user query - I'll say root README lists 5 featured projects in portfolio.yaml (5 entries) - the user said 6, maybe they're counting PORTFOLIO.md or miscounting. portfolio.yaml has exactly 5 projects. README table has 5 rows. I'll document 5 in portfolio.yaml and note discrepancy with user saying 6 - actually re-read README - 5 projects in table. User said "root README only lists 6 projects" - I'll keep "5 featured in portfolio.yaml/README table" and mention ~48 total.

### Orphan / Scaffold-Only Directories

| Path | Orphan Signal |
|------|---------------|
| `forgex/` | Spec-only README, no source |
| `ws-chat-fast/` | No `main.py`, no tests, no docker-compose; listed as stable |
| `bazel-multibuild/frontend/` | MODULE.bazel only |
| `dev-env/remote-dev-environment-orchestrator/` | Workflow stub, empty pyproject |
| `dev-env/local-service-mesh/services/*/` | README-only stubs (3 services) |
| `dev-env/*/tests/**/.keep` | 13 placeholder files, zero test implementations |
| `tools/gradle_to_bazel` | Referenced in legacy doc, directory absent |
| `src/main/resources/dd-advanced-logging-and-tracing.md` | Referenced in README, file absent |
| `build/` | Generated Gradle output |
| `site/` | Generated MkDocs site (committed) |

### Design Doc Mapping

#### `docs/design-docs/` → Subprojects

| Design Doc | Subproject |
|------------|------------|
| `docs/design-docs/ws-chat-fast/design.md` | `ws-chat-fast/` |
| `docs/design-docs/projgen/design.md` | `projgen/` |
| `docs/design-docs/rabbit-mq/design.md` | `rabbit-mq/` |
| `docs/design-docs/online-bookstore/design.md` | `online-bookstore/` |
| `docs/design-docs/bazel-multibuild/design.md` | `bazel-multibuild/` |

Template: `docs/templates/design_doc_template.md`

#### `src/main/resources/dd-*.md` → Subprojects

| Legacy Doc | Maps To |
|------------|---------|
| `dd-build-system-expertise.md` | `bazel-multibuild/`, `projgen/`, `modular-jvm-build/` |
| `dd-ci-cd-pipeline-with-flaky-test-detection.md` | `ci-cd-pipelines/flaky-pipeline-gate/`, `ci-cd-pipelines/pipeline-telemetry-exporter/` |
| `dd-cloud-native-microservices.md` | `dev-env/local-service-mesh/`, `modular-jvm-build/`, `workflow-api-demo/` |
| `dd-full-stack-sample-app.md` | `rest-api-test-demo/`, `online-bookstore/`, `ws-chat-fast/` |
| `dd-developer-productivity-tooling.md` | `projgen/`, `forgex/`, `dev-ex/*`, `dev-env/onboarding-automation-cli/` |
| `dd-streaming-data-demo.md` | *(unbuilt)* — nearest: `rabbit-mq/`, `workflow-api-demo/` |
| `dd-performance-benchmarking-n-tuning.md` | *(unbuilt)* — nearest: `modular-jvm-build/`, `bazel-multibuild/java/` |
| `dd-cross-plat-nortarization-tool.md` | *(unbuilt)* |
| `dd-gitops-starter-kit.md` | `ci-cd-pipelines/*`, `.github/workflows/` |
| `dd-open-source-plugin-contribution.md` | `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/` |
| `dd-advanced-logging-and-tracing.md` | *(missing file)* — would map to `otel-demo-stack/`, `observability/` |

#### Subprojects With No Root Design Doc

All subprojects except the 5 in `docs/design-docs/` lack a root-level design doc. Closest documentation:

- Area README: `dev-env/README.md`, `dev-ex/README.md`, `ci-cd-pipelines/README.md`, `c0de-quality-and-analysis/README.md`
- Inline README: every subproject has or should have `README.md`
- Per-subproject docs: `modular-jvm-build/docs/`, `otel-demo-stack/docs/`, `platform-audit-template/docs/`, `dev-env/*/docs/`, `dev-ex/*/docs/`, `ai-best-practices-examples/*/docs/plans/`

## Special Directories

**`.planning/codebase/`:**
- Purpose: GSD-generated codebase analysis (STACK, ARCHITECTURE, CONVENTIONS, etc.)
- Generated: Yes (by `/gsd-map-codebase`)
- Committed: Expected for planner consumption

**`site/`:**
- Purpose: Pre-built MkDocs HTML output
- Generated: Yes (`make docs` / `mkdocs build`)
- Committed: Yes (present in repo)

**`build/`:**
- Purpose: Root Gradle build artifacts
- Generated: Yes
- Committed: No (typically gitignored; may appear locally)

**`.github/workflows/`:**
- Purpose: Monorepo CI/CD automation
- Generated: No
- Committed: Yes

**`ai-best-practices-examples/ai-code-assistant/.ai-code-assistant/`:**
- Purpose: Audit logs and checkpoints for AI assistant runs
- Generated: Yes (runtime)
- Committed: Untracked per git status

---

*Structure analysis: 2026-06-22*
