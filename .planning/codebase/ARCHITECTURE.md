<!-- refreshed: 2026-06-22 -->
# Architecture

**Analysis Date:** 2026-06-22

## System Overview

This repository is a **portfolio monorepo**: a collection of loosely coupled, independently runnable subprojects grouped by competency domain. There is no single deployable application at the root. Instead, the root provides shared documentation, CI orchestration, and a project launcher while each subproject owns its own runtime, dependencies, and entry point.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Portfolio Root (orchestration layer)                     │
│  `README.md` · `portfolio.yaml` · `tools/run` · `.github/workflows/`        │
│  `docs/` (MkDocs) · `src/main/resources/dd-*.md` (legacy design docs)       │
└──────────┬──────────────────┬──────────────────┬──────────────────────────┘
           │                  │                  │
           ▼                  ▼                  ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────────────────────┐
│  Runnable Demos  │ │  Tooling / DX    │ │  Scaffold / Spec-Only Projects   │
│  (stable/beta)   │ │  (beta)          │ │  (scaffold)                      │
│                  │ │                  │ │                                  │
│ ws-chat-fast     │ │ projgen          │ │ forgex                           │
│ rest-api-test-   │ │ dev-ex/*         │ │ ws-chat-fast (no main.py)        │
│   demo           │ │ dev-env/*        │ │ bazel-multibuild/frontend        │
│ otel-demo-stack  │ │ ci-cd-pipelines/*│ │ remote-dev-env-orchestrator      │
│ workflow-api-    │ │ c0de-quality-*   │ │ local-service-mesh/service stubs │
│   demo           │ │ ai-best-practices│ │                                  │
│ online-bookstore │ │   -examples/*    │ │                                  │
│ rabbit-mq        │ │ platform-audit-  │ │                                  │
│ modular-jvm-build│ │   template       │ │                                  │
│ bazel-multibuild │ │                  │ │                                  │
└────────┬─────────┴────────┬─────────┴──────────┬───────────────────────────┘
         │                  │                     │
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│              External / Local Infrastructure (per subproject)                │
│  Docker Compose · Redis · PostgreSQL · RabbitMQ · OTel Collector · Grafana │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Portfolio Documentation Gap

The root `README.md` and `portfolio.yaml` catalog **6 projects** (via `<!-- PROJECTS_TABLE_START -->` marker). The repo actually contains **48 subprojects** across **17 top-level areas**. `docs/index.md`, `mkdocs.yml`, and `tools/run.yaml` also reference only the same 6 IDs. The remaining subprojects are documented in area-level READMEs (`dev-env/README.md`, `ci-cd-pipelines/README.md`, etc.) but not in the root portfolio table.

## Component Responsibilities

| Component | Responsibility | File / Path |
|-----------|----------------|-------------|
| Portfolio manifest | Canonical metadata for 6 featured projects (status, stack, tags) | `portfolio.yaml` |
| Project launcher | Runs docker-compose or CLI for featured projects | `tools/run`, `tools/run.yaml` |
| Root CI | Path-filtered lint/test across Python, Kotlin, Bazel, docs | `.github/workflows/ci.yml` |
| MkDocs site | Published design docs for 5 featured projects | `mkdocs.yml`, `docs/` |
| Legacy design docs | Aspirational project specs (10 files; README mentions 11 including missing `dd-advanced-logging-and-tracing.md`) | `src/main/resources/dd-*.md` |
| README table generator | Syncs README project table from YAML | `scripts/gen_readme_table.py` |
| Featured WebSocket demo | Real-time chat routers (incomplete — no app bootstrap) | `ws-chat-fast/app/` |
| REST API reference demo | Production-style FastAPI with Docker + CI | `rest-api-test-demo/app/main.py` |
| Observability demo | API + worker + OTel collector pipeline | `otel-demo-stack/` |
| Workflow pipeline demo | API → Redis queue → worker → PostgreSQL | `workflow-api-demo/` |
| Project scaffolder | Jinja2/Click CLI for Bazel/Gradle templates | `projgen/src/projgen/cli.py` |
| AI code assistant | Policy-aware test generation CLI | `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py` |
| JVM modular demo | Gradle multi-module Spring Boot (`core` → `api` → `app`) | `modular-jvm-build/` |
| Polyglot build demo | Bazel modules for Java, frontend placeholder, flags tutorial | `bazel-multibuild/` |
| Dev environment scaffolds | Devcontainer features, local mesh, onboarding CLI | `dev-env/` |
| DevEx scaffolds | Survey, friction scoring, adoption tracking, changelog gen | `dev-ex/` |
| CI/CD scaffolds | Flaky gate, canary controller, DORA metrics, pipeline templates | `ci-cd-pipelines/` |
| Code quality scaffolds | detekt rules, license scan, dead code, OpenAPI diff, Semgrep | `c0de-quality-and-analysis/` |
| ForgeX spec | Full product spec only — no implementation | `forgex/README.md` |
| Standalone observability stack | Prometheus + Grafana compose (separate from otel-demo-stack) | `observability/docker-compose.yml` |

## Pattern Overview

**Overall:** Polyglot portfolio monorepo with **domain-grouped subprojects**, each following its own architectural pattern. No shared application framework at root; conventions repeat by technology family.

**Key Characteristics:**
- **Independent deployability:** Each subproject has its own `pyproject.toml`, `build.gradle.kts`, or `BUILD.bazel`; root `build.gradle.kts` is a minimal Kotlin shell unrelated to subprojects.
- **Technology families:** Python FastAPI services (API + optional worker), Gradle/Kotlin JVM apps, Bazel polyglot modules, Markdown/spec scaffolds.
- **Scaffold-first expansion:** Many subprojects under `dev-env/`, `dev-ex/`, `ci-cd-pipelines/`, and `c0de-quality-and-analysis/` ship minimal `src/main.py` CLIs with smoke tests — architecture is documented in area READMEs, not fully implemented.
- **Documentation-driven portfolio:** Legacy `dd-*.md` specs describe aspirational systems; partial implementations exist across multiple subprojects.

## Layers

**Orchestration (root):**
- Purpose: CI, docs publishing, portfolio metadata, project runner
- Location: repo root, `.github/`, `docs/`, `tools/`, `scripts/`
- Contains: GitHub Actions, MkDocs config, YAML manifests, shell scripts
- Depends on: subproject paths for path-filter triggers
- Used by: contributors, GitHub Actions, MkDocs Pages

**Application (subproject):**
- Purpose: Runnable demos and tools
- Location: each subproject root (e.g. `rest-api-test-demo/`, `projgen/`)
- Contains: FastAPI apps, Kotlin mains, Bazel targets, CLI entry points
- Depends on: language-specific deps, Docker Compose where present
- Used by: `./tools/run <id>`, direct `docker compose up`, or CLI invocation

**Infrastructure (local/external):**
- Purpose: Backing services for demos
- Location: `docker-compose.yml` per subproject, `observability/`, `otel-demo-stack/collector/`
- Contains: Redis, PostgreSQL, RabbitMQ, OTel Collector, Prometheus, Grafana, Caddy
- Depends on: Docker
- Used by: API/worker subprojects at runtime

**Documentation:**
- Purpose: Design docs, runbooks, architecture overviews
- Location: `docs/design-docs/`, `src/main/resources/dd-*.md`, per-subproject `docs/`
- Contains: MkDocs-navigated design docs (5 projects), legacy specs (10), inline READMEs (all areas)
- Depends on: nothing at runtime
- Used by: planners, MkDocs site, GSD tooling

## Data Flow

### Primary Request Path — FastAPI Service Demo

Representative pattern used by `rest-api-test-demo`, `online-bookstore`, `workflow-api-demo`, `otel-demo-stack`:

1. HTTP client hits FastAPI app (`rest-api-test-demo/app/main.py`)
2. Route handler validates input via Pydantic models
3. Business logic reads/writes datastore or enqueues work
4. JSON response returned to client

### Distributed Pipeline — API + Queue + Worker

Used by `workflow-api-demo/`:

1. Client `POST /jobs` → `workflow-api-demo/api/main.py` enqueues job ID to Redis
2. Worker `workflow-api-demo/worker/worker.py` blocks on `BRPOP`, processes job
3. Worker writes result to PostgreSQL
4. Client `GET /jobs/{id}` reads status from PostgreSQL via API

### Observability Pipeline — OTel Demo

Used by `otel-demo-stack/`:

1. API request creates span in `otel-demo-stack/api/main.py` (OTel SDK)
2. W3C trace context propagated to worker via headers or shared trace ID
3. Worker emits spans/metrics in `otel-demo-stack/worker/worker.py`
4. OTLP export to collector (`otel-demo-stack/collector/otelcol.yaml`)
5. Metrics/traces visible in collector logs or configured backend (Grafana in full stack)

### WebSocket Chat (Partial)

Used by `ws-chat-fast/app/chat.py`:

1. Client connects via WebSocket to `/chatroom/{username}`
2. `ConnectionManager` in `ws-chat-fast/app/ws_manager.py` registers connection
3. Messages broadcast to all connected clients
4. **Gap:** No `main.py` or `FastAPI()` bootstrap — routers exist but app is not wired for standalone run

### CLI Tool Path

Used by `projgen`, `ai-code-assistant`, `dev-env/onboarding-automation-cli`:

1. User invokes CLI entry point (`projgen/src/projgen/cli.py`, `[project.scripts]` in `pyproject.toml`)
2. Argument parsing (Click or argparse)
3. Core logic (template render, repo scan, test generation)
4. File system output or stdout JSON

**State Management:**
- **WebSocket sessions:** In-memory `ConnectionManager` dict in `ws-chat-fast/app/ws_manager.py` (no Redis despite README claims)
- **Job queues:** Redis lists in `workflow-api-demo`
- **Persistent data:** PostgreSQL in `workflow-api-demo`; CSV/SQLAlchemy in `online-bookstore`
- **RAG vectors:** ChromaDB in `ai-best-practices-examples/knowledge-qa-system`
- **Audit logs:** JSONL in `ai-best-practices-examples/ai-code-assistant/.ai-code-assistant/audit.log.jsonl`

## Key Abstractions

**FastAPI Application Factory:**
- Purpose: Standard HTTP service entry
- Examples: `rest-api-test-demo/app/main.py`, `otel-demo-stack/api/main.py`, `workflow-api-demo/api/main.py`, `online-bookstore/src/main.py`
- Pattern: Module-level `app = FastAPI(...)`, uvicorn as ASGI server

**Click CLI Group:**
- Purpose: Multi-command Python CLI
- Examples: `projgen/src/projgen/cli.py`
- Pattern: `@click.group()` with subcommands, optional telemetry init

**Argparse Subcommand CLI:**
- Purpose: Policy-aware automation CLI
- Examples: `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py`
- Pattern: `build_parser()` → subparsers for `gen-tests`, `scan`, `automate`, etc.

**Gradle Multi-Module JVM:**
- Purpose: Layered JVM architecture
- Examples: `modular-jvm-build/` (`core` → `api` → `app`), `rabbit-mq/` (Kotlin app + Python channel)
- Pattern: `settings.gradle.kts` includes modules; Spring Boot in `app` module only

**Bazel Module:**
- Purpose: Polyglot build targets
- Examples: `bazel-multibuild/java/`, `bazel-multibuild/flags-parsing-tutorial/`
- Pattern: `MODULE.bazel` / `WORKSPACE`, `BUILD.bazel` per package

**Smoke-Test Scaffold CLI:**
- Purpose: Placeholder for future CI/CD/DevEx tools
- Examples: `ci-cd-pipelines/flaky-pipeline-gate/src/main.py`, all 6 `ci-cd-pipelines/*` projects
- Pattern: Minimal stdlib CLI + 1–2 pytest smoke tests; full behavior described in area README

## Entry Points

**Root orchestration:**
- Location: `Makefile` (`make test`, `make docs`, `make lint`)
- Triggers: Developer or CI
- Responsibilities: Aggregate pytest + gradlew across repo (best-effort, failures expected per Makefile comment)

**Featured project runner:**
- Location: `tools/run` + `tools/run.yaml`
- Triggers: `./tools/run <project-id>`
- Responsibilities: `docker compose up` or CLI for 5 featured IDs only

**Subproject entry points (representative):**

| Subproject | Entry Point | Invocation |
|------------|-------------|------------|
| `rest-api-test-demo` | `app/main.py` | `uvicorn app.main:app` or Docker Compose |
| `otel-demo-stack` | `api/main.py`, `worker/worker.py` | `docker compose up --build` |
| `workflow-api-demo` | `api/main.py`, `worker/worker.py` | `docker compose up --build` |
| `projgen` | `src/projgen/cli.py` | `python -m projgen` or `projgen` script |
| `ai-code-assistant` | `src/ai_code_assistant/cli.py` | `ai-code-assistant gen-tests` |
| `modular-jvm-build` | `app/.../Application.kt` | `./gradlew :app:bootRun` |
| `rabbit-mq` | `app/.../App.kt`, `app/.../channel.py` | Gradle + Python pika consumer |
| `bazel-multibuild/java` | `java/src/main/java/.../App.java` | `bazel run //java:app` |
| `onboarding-automation-cli` | `src/main/kotlin/.../Main.kt` | Gradle `run` task |
| `platform-audit-template` | `scripts/gcp_billing_summary.py` | Direct Python script execution |

## Portfolio Organization

### Top-Level Domain Areas

| Area | Subproject Count | Primary Pattern | Typical Maturity |
|------|------------------|-----------------|------------------|
| `ai-best-practices-examples/` | 5 | Python AI/LLM CLI + FastAPI | 1 stable, 4 beta |
| `c0de-quality-and-analysis/` | 5 | Python/Kotlin CI quality tools | beta (scaffold CLIs) |
| `ci-cd-pipelines/` | 6 | Python CI/CD automation CLIs | beta (scaffold CLIs) |
| `dev-env/` | 5 (+ 3 service stubs) | Devcontainer, Compose mesh, Kotlin CLI | beta / scaffold |
| `dev-ex/` | 4 | FastAPI + Python/TS analytics | beta |
| `bazel-multibuild/` | 3 modules | Bazel polyglot | beta / scaffold |
| Standalone demos | 9 | FastAPI, Gradle, Bazel, spec | mixed |
| `tools/`, `observability/`, `docs/`, `src/` | 4 | Root support | beta / docs-only |

### Full Subproject Catalog (Purpose + Maturity)

#### ai-best-practices-examples/

| Path | Purpose | Maturity |
|------|---------|----------|
| `ai-best-practices-examples/ai-code-assistant` | Policy-aware Python test generation CLI with audit, risk scoring, LLM adapter | **stable** |
| `ai-best-practices-examples/ai-image-video-generator` | Gradio UI + ComfyUI pipeline for image/video generation | **beta** |
| `ai-best-practices-examples/chat-ai` | FastAPI streaming chat with memory and tool calling | **beta** |
| `ai-best-practices-examples/domain-expert-ai` | QLoRA fine-tuning scaffold for domain expert assistant | **beta** |
| `ai-best-practices-examples/knowledge-qa-system` | RAG over PDF/Notion/web with ChromaDB citations | **beta** |

#### c0de-quality-and-analysis/

| Path | Purpose | Maturity |
|------|---------|----------|
| `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library` | Custom detekt rule set for team Kotlin conventions | **beta** |
| `c0de-quality-and-analysis/license-compliance-scanner` | Transitive dependency license scan + SBOM | **beta** |
| `c0de-quality-and-analysis/dead-code-surface-reporter` | Call graph + git blame dead code backlog | **beta** |
| `c0de-quality-and-analysis/api-breaking-change-detector` | OpenAPI diff merge gate | **beta** |
| `c0de-quality-and-analysis/security-hotspot-annotator` | Semgrep → PR annotation formatter | **beta** |

#### ci-cd-pipelines/

| Path | Purpose | Maturity |
|------|---------|----------|
| `ci-cd-pipelines/pipeline-telemetry-exporter` | Emit CI job spans as OTel traces | **beta** |
| `ci-cd-pipelines/flaky-pipeline-gate` | Flakiness score for merge gating | **beta** |
| `ci-cd-pipelines/release-lead-time-calculator` | DORA lead/cycle time API | **beta** |
| `ci-cd-pipelines/canary-deployment-controller` | K8s canary traffic shifting + rollback | **beta** |
| `ci-cd-pipelines/self-service-pipeline-template-engine` | Backstage-style pipeline YAML generator | **beta** |
| `ci-cd-pipelines/pipeline-cost-analyzer` | GitHub Actions cost attribution | **beta** |

#### dev-env/

| Path | Purpose | Maturity |
|------|---------|----------|
| `dev-env/devcontainer-feature-library` | Composable devcontainer features (Kafka, Postgres, Keycloak, OTel) | **beta** (`.keep`-only test dirs) |
| `dev-env/environment-drift-detector` | Local toolchain vs canonical manifest diff + fix script | **beta** (no tests) |
| `dev-env/local-service-mesh` | Docker Compose + Caddy K8s-like local routing | **beta** (`.keep`-only test dirs) |
| `dev-env/onboarding-automation-cli` | Kotlin CLI for new-hire repo/tool setup | **beta** (no tests) |
| `dev-env/remote-dev-environment-orchestrator` | Temporal workflow for ephemeral cloud dev envs | **scaffold** |
| `dev-env/local-service-mesh/services/user-service` | Mesh routing placeholder stub | **scaffold** |
| `dev-env/local-service-mesh/services/order-service` | Mesh routing placeholder stub | **scaffold** |
| `dev-env/local-service-mesh/services/payment-service` | Mesh routing placeholder stub | **scaffold** |

#### dev-ex/

| Path | Purpose | Maturity |
|------|---------|----------|
| `dev-ex/developer-satisfaction-pulse-system` | Weekly pulse + quarterly SPACE surveys | **beta** |
| `dev-ex/tooling-adoption-tracker` | Opt-in IDE/CLI/portal adoption telemetry | **beta** |
| `dev-ex/inner-loop-friction-scorer` | Composite friction score from CI/Git signals | **beta** |
| `dev-ex/platform-changelog-migration-generator` | API diff → changelog + AST migration scripts | **beta** |

#### Standalone top-level subprojects

| Path | Purpose | Maturity |
|------|---------|----------|
| `ws-chat-fast` | FastAPI WebSocket chat (routers only; no app bootstrap) | **scaffold** (listed as **stable** in `portfolio.yaml` — metadata drift) |
| `projgen` | Bazel/Gradle project scaffolder CLI | **beta** |
| `rabbit-mq` | Gradle Kotlin + Python RabbitMQ pub/sub demo | **beta** |
| `online-bookstore` | FastAPI e-commerce with CSV-backed catalog | **beta** (listed as **stable** in `portfolio.yaml`) |
| `bazel-multibuild` | Polyglot Bazel monorepo parent | **beta** |
| `bazel-multibuild/java` | Java Bazel sample app + JUnit | **beta** |
| `bazel-multibuild/frontend` | Frontend Bazel MODULE placeholder | **scaffold** |
| `bazel-multibuild/flags-parsing-tutorial` | Bazel flags / Starlark tutorial | **beta** |
| `modular-jvm-build` | Gradle multi-module Spring Boot (`core`/`api`/`app`) | **beta** |
| `otel-demo-stack` | OTel-in-a-box: API + worker + collector | **stable** |
| `platform-audit-template` | SRE platform maturity audit templates + scripts | **beta** |
| `rest-api-test-demo` | Production-style FastAPI REST API with Docker + CI | **stable** |
| `workflow-api-demo` | FastAPI → Redis → worker → PostgreSQL pipeline | **beta** |
| `forgex` | Unified polyglot repo generator product spec | **scaffold** (README only, ~950 lines) |
| `tools` | `./tools/run` project launcher from YAML manifest | **beta** |
| `observability` | Standalone Prometheus + Grafana compose stack | **beta** |

### Orphan / Scaffold-Only Directories

| Path | Signal | Notes |
|------|--------|-------|
| `forgex/` | README spec only, no `src/` | Full product requirements doc; zero implementation |
| `ws-chat-fast/` | Routers in `app/` but no `main.py`, no tests, no docker-compose | Metadata says stable; code is incomplete |
| `bazel-multibuild/frontend/` | `MODULE.bazel` only | No application source |
| `dev-env/remote-dev-environment-orchestrator/` | Single workflow file, empty `pyproject.toml` | Temporal workflow stub |
| `dev-env/local-service-mesh/services/*/` | Empty README stubs | Referenced by mesh compose but not implemented |
| `dev-env/*/tests/*/.keep` | 13 `.keep` files across 5 dev-env projects | Test directory placeholders with no test files |
| `tools/gradle_to_bazel` | Referenced in `dd-build-system-expertise.md` | **Not present** in repo |
| `src/main/resources/dd-advanced-logging-and-tracing.md` | Listed in root `README.md` | **Not present** (10 of 11 legacy docs exist) |
| Root `build/` | Gradle build output | Generated, not a subproject |
| Root `site/` | MkDocs built HTML | Generated from `mkdocs build` |

### Design Doc → Subproject Mapping

#### MkDocs design docs (`docs/design-docs/`)

| Design Doc | Subproject | In `portfolio.yaml`? |
|------------|------------|---------------------|
| `docs/design-docs/ws-chat-fast/design.md` | `ws-chat-fast/` | Yes |
| `docs/design-docs/projgen/design.md` | `projgen/` | Yes |
| `docs/design-docs/rabbit-mq/design.md` | `rabbit-mq/` | Yes |
| `docs/design-docs/online-bookstore/design.md` | `online-bookstore/` | Yes |
| `docs/design-docs/bazel-multibuild/design.md` | `bazel-multibuild/` | Yes |

Navigated via `mkdocs.yml` — **5 projects only**. No MkDocs entries for the other 43 subprojects.

#### Per-subproject docs (not in root `docs/design-docs/`)

| Doc Path | Subproject |
|----------|------------|
| `modular-jvm-build/docs/architecture.md` | `modular-jvm-build/` |
| `otel-demo-stack/docs/opentelemetry-verification-guide.md` | `otel-demo-stack/` |
| `rest-api-test-demo/docs/test-plan.md` | `rest-api-test-demo/` |
| `platform-audit-template/docs/audit-template.md` | `platform-audit-template/` |
| `dev-env/*/docs/architecture/overview.md` | Respective `dev-env/*` subproject |
| `dev-ex/*/docs/*.md` | Respective `dev-ex/*` subproject |
| `ai-best-practices-examples/*/docs/plans/*.md` | Respective AI subproject |

#### Legacy aspirational specs (`src/main/resources/dd-*.md`)

| Legacy Doc | Primary Subproject Mapping | Coverage |
|------------|---------------------------|----------|
| `dd-build-system-expertise.md` | `bazel-multibuild/`, `projgen/`, `modular-jvm-build/` | Partial — no C++ module, no sanitizer rules, `tools/gradle_to_bazel` missing |
| `dd-ci-cd-pipeline-with-flaky-test-detection.md` | `ci-cd-pipelines/flaky-pipeline-gate/`, `ci-cd-pipelines/pipeline-telemetry-exporter/`, `observability/` | Partial — scaffold CLIs, no Gradle plugin or Grafana dashboard |
| `dd-cloud-native-microservices.md` | `dev-env/local-service-mesh/`, `modular-jvm-build/`, `workflow-api-demo/`, `otel-demo-stack/` | Partial — no Kafka, Terraform, or AWS deployment |
| `dd-full-stack-sample-app.md` | `rest-api-test-demo/`, `online-bookstore/`, `ws-chat-fast/` | Partial — no React frontend, no Playwright/Cypress |
| `dd-developer-productivity-tooling.md` | `projgen/`, `forgex/` (spec), `dev-ex/*`, `dev-env/onboarding-automation-cli/`, `dev-env/environment-drift-detector/` | Partial — no IntelliJ plugin |
| `dd-streaming-data-demo.md` | **Unbuilt** — closest: `rabbit-mq/`, `workflow-api-demo/` | No Kafka producer/consumer pipeline |
| `dd-performance-benchmarking-n-tuning.md` | **Unbuilt** — candidates: `modular-jvm-build/`, `online-bookstore/`, `bazel-multibuild/java/` | No JMH or pytest-benchmark module |
| `dd-cross-plat-nortarization-tool.md` | **Unbuilt** — no matching subproject | Spec only |
| `dd-gitops-starter-kit.md` | `ci-cd-pipelines/*`, `platform-audit-template/`, `.github/workflows/` | Partial — no Terraform, Argo CD, or Flux |
| `dd-open-source-plugin-contribution.md` | `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/`, `rabbit-mq/buildSrc/` | Partial — no external OSS PR |
| `dd-advanced-logging-and-tracing.md` (referenced, missing) | Would map to `otel-demo-stack/`, `observability/` | File not present |

## Architectural Constraints

- **Threading:** Python FastAPI services use asyncio event loop; workers use blocking Redis `BRPOP` in separate processes. No shared thread pool across subprojects.
- **Global state:** WebSocket `ConnectionManager` singleton per process in `ws-chat-fast/app/ws_manager.py`. No cross-service shared state except via Redis/PostgreSQL in pipeline demos.
- **Circular imports:** Not detected at portfolio level; subprojects are isolated. `projgen` uses relative imports from `generators.scaffold` (sibling package under `src/`).
- **Build isolation:** Root `settings.gradle.kts` names project `portfolio-project` but does not include subproject modules. Each JVM subproject has its own Gradle settings.
- **Metadata drift:** `portfolio.yaml` status labels (`stable` for `ws-chat-fast`, `online-bookstore`) do not always match code completeness assessed in this analysis.
- **CI scope:** Root `.github/workflows/ci.yml` uses path filters — changes outside Python/Kotlin/Bazel/docs paths may not trigger relevant jobs. Individual subprojects (`rest-api-test-demo`, `otel-demo-stack`, `c0de-quality-and-analysis/*`) have their own `.github/workflows/`.

## Anti-Patterns

### Listing Incomplete Projects as Stable

**What happens:** `portfolio.yaml` marks `ws-chat-fast` as `stable` but the subproject lacks `main.py`, tests, and docker-compose.
**Why it's wrong:** Portfolio consumers expect runnable, tested demos; `./tools/run ws-chat-fast` references docker-compose that does not exist.
**Do this instead:** Align `portfolio.yaml` status with actual maturity; add `ws-chat-fast/app/main.py` wiring routers before marking stable.

### Scaffold CLIs Presented as Full Systems

**What happens:** `ci-cd-pipelines/*` and `c0de-quality-and-analysis/*` ship minimal `src/main.py` with smoke tests while area READMEs describe full production behavior (K8s controllers, Semgrep integration).
**Why it's wrong:** Executors may implement duplicate logic not knowing a stub exists.
**Do this instead:** Treat area README as target architecture; extend existing `src/main.py` rather than creating parallel modules.

### Documentation Split Across Three Locations

**What happens:** Design docs live in `docs/design-docs/` (5 projects), `src/main/resources/dd-*.md` (10 legacy specs), and per-subproject `docs/` folders — with no index linking all 48 subprojects.
**Why it's wrong:** Planners cannot find the canonical spec for a subproject.
**Do this instead:** Add subproject to nearest design doc mapping; prefer `docs/design-docs/<name>/design.md` for featured projects; reference legacy `dd-*.md` in subproject README when aspirational.

## Error Handling

**Strategy:** Per-subproject; no shared error library. Python services use FastAPI `HTTPException`; CLI tools use exit codes and stderr.

**Patterns:**
- FastAPI: Pydantic validation errors → 422; explicit `HTTPException(status_code=...)` in route handlers
- CLI: `argparse`/`click` validation; `ai-code-assistant` returns JSON error payloads in headless mode
- Gradle/Kotlin: standard exception propagation; Spring Boot error handlers in `modular-jvm-build`

## Cross-Cutting Concerns

**Logging:** Python services use stdlib `logging` or uvicorn logger (`ws-chat-fast/app/chat.py`). No shared logging framework across monorepo.

**Validation:** Pydantic models in FastAPI subprojects; Click/argparse for CLIs; `projgen/src/projgen/validation.py` for template validation.

**Authentication:** Basic auth stub in `ws-chat-fast/app/security.py` for exclusive chatrooms; JWT mentioned in legacy full-stack doc but not implemented portfolio-wide.

**CI/CD:** Root `.github/workflows/ci.yml` (path-filtered Python/Kotlin/Bazel/docs); `release-please.yml`, `codeql.yml`, `readme-drift.yml`, `pages.yml` at root. Subproject-specific workflows in `rest-api-test-demo/.github/workflows/build.yml`, `otel-demo-stack/.github/workflows/build.yml`, and each `c0de-quality-and-analysis/*/.github/workflows/ci.yml`.

**Observability:** `otel-demo-stack/` is the reference implementation. `observability/` provides standalone Prometheus/Grafana. Other subprojects do not share OTel instrumentation.

---

*Architecture analysis: 2026-06-22*
