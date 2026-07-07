# Technology Stack

**Analysis Date:** 2026-06-22

## Languages

**Primary:**
- Python 3.10–3.12 — Dominant language across demos, AI examples, CI/CD tooling, DevEx utilities, and most FastAPI services (`ws-chat-fast/`, `workflow-api-demo/`, `otel-demo-stack/`, `online-bookstore/`, `ai-best-practices-examples/`, `ci-cd-pipelines/`, `dev-ex/`, `c0de-quality-and-analysis/`)
- Kotlin 1.9.x–2.3.0 — JVM services and Gradle modules (`modular-jvm-build/`, `rabbit-mq/`, `dev-env/onboarding-automation-cli/`, root `build.gradle.kts`)
- Java — Bazel Java targets (`bazel-multibuild/java/`), legacy bookstore design docs (implementation is Python)
- TypeScript 5.6 — Node tooling (`dev-ex/tooling-adoption-tracker/package.json`)

**Secondary:**
- JavaScript/TypeScript (React) — Bazel frontend module (`bazel-multibuild/frontend/`)
- HTML/Jinja2 templates — Server-rendered UIs (`ws-chat-fast/templates/`, `site/`)
- Shell — Run scripts in `c0de-quality-and-analysis/*/scripts/`, `platform-audit-template/scripts/`
- YAML/TOML — CI configs, policy files, `portfolio.yaml`, per-project `pyproject.toml`

## Runtime

**Environment:**
- Python 3.12 — Root CI target (`.github/workflows/ci.yml`)
- Python 3.10–3.11 — Per-project minimums in `pyproject.toml` files
- JVM 17–23 — Toolchains vary by Gradle module (root uses JVM 23; `modular-jvm-build/` uses 17)
- Node 20.13.1 — Bazel frontend toolchain (`bazel-multibuild/frontend/MODULE.bazel`)
- Docker — Local orchestration for multi-service demos (`docker-compose.yml` in `workflow-api-demo/`, `otel-demo-stack/`, `observability/`, etc.)

**Package Manager:**
- pip — Primary for Python; root `Makefile` bootstrap references `requirements.txt` (file not present at repo root; subprojects use local `requirements.txt` or `pyproject.toml`)
- setuptools — Standard build backend in all `pyproject.toml` projects
- Gradle Wrapper — Root (`gradle/wrapper/gradle-wrapper.properties` → 8.10.2), `rabbit-mq/gradlew` (8.14.2), `modular-jvm-build/gradle/wrapper/` (8.6)
- npm — `dev-ex/tooling-adoption-tracker/package.json` only (no lockfile detected)
- Bazelisk — Root CI Bazel job (`.github/workflows/ci.yml`)
- Lockfile: **Mixed** — Some Python projects pin versions in `requirements.txt` (`otel-demo-stack/api/requirements.txt`); others unpinned (`ws-chat-fast/requirements.txt`); no root `poetry.lock`/`uv.lock`; Bazel frontend references `pnpm-lock.yaml` in `MODULE.bazel` but file not present in repo

## Frameworks

**Core:**
- FastAPI — HTTP/WebSocket APIs (`ws-chat-fast/app/`, `online-bookstore/src/main.py`, `workflow-api-demo/api/`, `otel-demo-stack/api/`, `dev-ex/developer-satisfaction-pulse-system/`, AI examples under `ai-best-practices-examples/`)
- Spring Boot 3.2.2 — JVM modular demo (`modular-jvm-build/app/build.gradle.kts`)
- Gradio 4.x — AI image/video UI (`ai-best-practices-examples/ai-image-video-generator/pyproject.toml`)
- LangChain 0.3.x — RAG pipeline (`ai-best-practices-examples/knowledge-qa-system/pyproject.toml`)
- Transformers/PEFT/bitsandbytes — Fine-tuning stack (`ai-best-practices-examples/domain-expert-ai/pyproject.toml`)
- Jinja2 + Click/Typer — Project scaffolding (`projgen/pyproject.toml`, `projgen/src/projgen/`)
- MkDocs — Documentation site (`mkdocs.yml`, `.github/workflows/pages.yml`)

**Testing:**
- pytest 8.x+ — Standard across Python subprojects; root CI runs `pytest -q --cov=.`
- JUnit Platform — Kotlin/Java Gradle tests (`build.gradle.kts`, `modular-jvm-build/app/build.gradle.kts`)
- Node test runner — `dev-ex/tooling-adoption-tracker/package.json` (`node --test`)

**Build/Dev:**
- Gradle 8.6–8.14.2 — JVM builds (version varies by subproject; see inventory below)
- Bazel (bzlmod) — `bazel-multibuild/java/MODULE.bazel`, `bazel-multibuild/frontend/MODULE.bazel`; legacy WORKSPACE in `bazel-multibuild/flags-parsing-tutorial/`
- Make — Root orchestration (`Makefile`) and per-project run/test shortcuts
- Ruff + mypy — Root CI Python lint/typecheck (`.github/workflows/ci.yml`)
- ktlint + detekt — Referenced in root CI Kotlin job; plugins not configured in root `build.gradle.kts`
- pre-commit — `black`, `ruff`, `ktlint` hooks (`.pre-commit-config.yaml`)

## Key Dependencies

**Critical:**
- `fastapi`, `uvicorn`, `pydantic` — Web stack across most Python services
- `redis`, `asyncpg` — Workflow queue demo (`workflow-api-demo/api/requirements.txt`, `workflow-api-demo/worker/requirements.txt`)
- `opentelemetry-*` — Observability demo (`otel-demo-stack/api/requirements.txt`, `otel-demo-stack/worker/requirements.txt`)
- `com.rabbitmq:amqp-client:5.25.0` — RabbitMQ client (`rabbit-mq/build.gradle`, `rabbit-mq/app/build.gradle`)
- `openai>=1.30.0` (optional) — LLM test generation (`ai-best-practices-examples/ai-code-assistant/pyproject.toml`)
- `langchain-chroma`, `chromadb`, `sentence-transformers` — Vector RAG (`ai-best-practices-examples/knowledge-qa-system/pyproject.toml`)
- Aspect Bazel rules — Polyglot build (`bazel-multibuild/frontend/MODULE.bazel`: `aspect_rules_js`, `aspect_rules_ts`, etc.)

**Infrastructure:**
- Docker Compose — Local stacks (`workflow-api-demo/docker-compose.yml`, `otel-demo-stack/docker-compose.yml`, `observability/docker-compose.yml`, `rest-api-test-demo/docker-compose.yml`, `dev-env/local-service-mesh/docker-compose.yml`)
- Redis 7, PostgreSQL 16 — Workflow demo containers (`workflow-api-demo/docker-compose.yml`)
- OTel Collector 0.96.0 — Trace/metric pipeline (`otel-demo-stack/docker-compose.yml`)
- Prometheus, Grafana, Loki — Observability stack (`observability/docker-compose.yml`)
- Caddy 2 — Local gateway (`dev-env/local-service-mesh/docker-compose.yml`)

## Configuration

**Environment:**
- Per-project env vars — e.g. `OPENAI_API_KEY` (`ai-best-practices-examples/ai-code-assistant/`), `AIVG_COMFYUI_BASE_URL` (`ai-best-practices-examples/ai-image-video-generator/src/ai_image_video_generator/config.py`), `REDIS_URL`/`DATABASE_URL` (`workflow-api-demo/worker/worker.py`), `OTEL_*` (`otel-demo-stack/api/main.py`), `KQ_CHROMA_PATH` (tests in `ai-best-practices-examples/knowledge-qa-system/tests/conftest.py`)
- `.env` files — Not committed; use env vars or compose `environment:` blocks
- Portfolio manifest — `portfolio.yaml` drives README project table via `scripts/gen_readme_table.py`

**Build:**
- Root: `Makefile`, `build.gradle.kts`, `settings.gradle.kts`, `gradle/wrapper/gradle-wrapper.properties`, `mkdocs.yml`, `portfolio.yaml`, `.pre-commit-config.yaml`
- Python: `pyproject.toml` (11 projects) or bare `requirements.txt` (9 files)
- JVM: `build.gradle.kts` / `build.gradle` + optional `gradlew`
- Bazel: `MODULE.bazel` + `BUILD` + legacy `WORKSPACE`
- Node: `package.json` + `tsconfig.json` (`dev-ex/tooling-adoption-tracker/`)

## Platform Requirements

**Development:**
- Python 3.10+ (3.12 recommended for CI parity)
- JDK 17–23 depending on Gradle subproject
- Docker + Docker Compose for multi-service demos
- Optional: Bazelisk, Gradle wrapper (`./gradlew`), GitHub CLI (`gh`) for `ai-code-assistant` PR ingestion, Ollama for `domain-expert-ai`, ComfyUI for `ai-image-video-generator`
- Dev container: `.devcontainer/devcontainer.json` (Ubuntu base, docker-in-docker, `make bootstrap`)

**Production:**
- GitHub Pages — Static site + MkDocs (`site/`, `.github/workflows/pages.yml`, `.github/workflows/static.yml`)
- Containerized services — Individual Dockerfiles in `otel-demo-stack/`, `workflow-api-demo/`, `rest-api-test-demo/`
- No unified cloud deployment manifest at monorepo root

---

## Per-Subproject Build Tooling Inventory

Legend: ✅ = present, ❌ = absent, ⚠️ = present but incomplete/empty

### Documented in root README (`portfolio.yaml` → `README.md`)

| Path | Makefile | pyproject.toml | requirements.txt | Gradle | Bazel | Docker | Notes |
|------|----------|----------------|------------------|--------|-------|--------|-------|
| `ws-chat-fast/` | ❌ | ❌ | ✅ (unpinned) | ❌ | ❌ | ❌ | README claims Redis; no Redis in code |
| `projgen/` | ❌ | ✅ | ✅ (`src/`) | ❌ (generates Gradle templates) | ❌ (generates Bazel templates) | ❌ | Dual `setup.py` + `pyproject.toml` |
| `rabbit-mq/` | ❌ | ❌ | ✅ (Python submodule) | ✅ + `gradlew` (8.14.2) | ❌ | ❌ | Kotlin/Java Gradle multi-module |
| `online-bookstore/` | ❌ | ⚠️ empty | ✅ (`src/`, unpinned) | ❌ | ❌ | ❌ | README says Java/Spring; code is FastAPI + CSV |
| `bazel-multibuild/` | ❌ | ❌ | ❌ | ❌ | ✅ (3 submodules) | ❌ | See Bazel breakdown below |

### `ai-best-practices-examples/` (not in README table)

| Path | Makefile | pyproject.toml | Gradle | Bazel | Docker |
|------|----------|----------------|--------|-------|--------|
| `ai-code-assistant/` | ❌ | ✅ | ❌ | ❌ | ❌ |
| `ai-image-video-generator/` | ❌ | ✅ | ❌ | ❌ | ❌ |
| `chat-ai/` | ❌ | ✅ | ❌ | ❌ | ❌ |
| `domain-expert-ai/` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `knowledge-qa-system/` | ❌ | ✅ | ❌ | ❌ | ❌ |

### `ci-cd-pipelines/` (not in README table)

| Path | Makefile | pyproject.toml | Gradle | Bazel |
|------|----------|----------------|--------|-------|
| `canary-deployment-controller/` | ✅ | ❌ | ❌ | ❌ |
| `flaky-pipeline-gate/` | ✅ | ❌ | ❌ | ❌ |
| `pipeline-cost-analyzer/` | ✅ | ❌ | ❌ | ❌ |
| `pipeline-telemetry-exporter/` | ✅ | ❌ | ❌ | ❌ |
| `release-lead-time-calculator/` | ✅ | ❌ | ❌ | ❌ |
| `self-service-pipeline-template-engine/` | ✅ | ❌ | ❌ | ❌ |

Pattern: Python `src/main.py` + `tests/` + Makefile (`run`, `test`); no packaging manifest.

### `c0de-quality-and-analysis/` (not in README table)

| Path | Makefile | pyproject.toml | Gradle | Bazel | Shell CI |
|------|----------|----------------|--------|-------|----------|
| `api-breaking-change-detector/` | ❌ | ❌ | ❌ | ❌ | ✅ `.github/workflows/ci.yml` |
| `dead-code-surface-reporter/` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `kotlin-custom-detekt-rules-library/` | ❌ | ❌ | ⚠️ CI expects `./gradlew` but no Gradle files in tree | ❌ | ✅ |
| `license-compliance-scanner/` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `security-hotspot-annotator/` | ❌ | ❌ | ❌ | ❌ | ✅ |

Pattern: Python/Kotlin source + `scripts/*.sh` runners; per-project GitHub Actions workflows.

### `dev-env/` (not in README table)

| Path | Makefile | pyproject.toml | Gradle | Docker Compose |
|------|----------|----------------|--------|----------------|
| `devcontainer-feature-library/` | ❌ | ❌ | ❌ | ❌ (example `.devcontainer/` only) |
| `environment-drift-detector/` | ❌ | ⚠️ empty | ❌ | ❌ |
| `local-service-mesh/` | ❌ | ❌ | ❌ | ✅ |
| `onboarding-automation-cli/` | ❌ | ❌ | ✅ `build.gradle.kts` (Kotlin 1.9.24) | ❌ |
| `remote-dev-environment-orchestrator/` | ❌ | ⚠️ empty | ❌ | ❌ |

### `dev-ex/` (not in README table)

| Path | Makefile | pyproject.toml | package.json | Gradle |
|------|----------|----------------|--------------|--------|
| `developer-satisfaction-pulse-system/` | ✅ | ✅ | ❌ | ❌ |
| `inner-loop-friction-scorer/` | ✅ | ✅ | ❌ | ❌ |
| `platform-changelog-migration-generator/` | ✅ | ✅ | ❌ | ❌ |
| `tooling-adoption-tracker/` | ✅ | ❌ | ✅ | ❌ |

### Other top-level projects (not in README table)

| Path | Makefile | pyproject.toml | requirements.txt | Gradle | Bazel | Docker |
|------|----------|----------------|------------------|--------|-------|--------|
| `modular-jvm-build/` | ❌ | ❌ | ❌ | ✅ (8.6, no `gradlew`) | ❌ | ❌ |
| `otel-demo-stack/` | ❌ | ❌ | ✅ (api/worker) | ❌ | ❌ | ✅ compose + Dockerfiles |
| `workflow-api-demo/` | ❌ | ❌ | ✅ (api/worker) | ❌ | ❌ | ✅ compose + Dockerfiles |
| `rest-api-test-demo/` | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ |
| `observability/` | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ compose only |
| `forgex/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Design doc only (`forgex/README.md`) |
| `platform-audit-template/` | ❌ | ❌ | ⚠️ `scripts/requirements-test.txt` | ❌ | ❌ | ❌ | Docs + scripts only |
| `tools/gradle_to_bazel/` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Utility scripts |

### Bazel submodules (`bazel-multibuild/`)

| Path | WORKSPACE | MODULE.bazel | BUILD | CI trigger note |
|------|-----------|--------------|-------|-----------------|
| `bazel-multibuild/java/` | ✅ | ✅ (rules_java 7.11.1, rules_oci 1.4.0) | ✅ | Root CI filters on `BUILD.bazel` — these use `BUILD` (no extension) |
| `bazel-multibuild/frontend/` | ✅ | ✅ (Node 20.13.1, aspect_rules_js/ts) | ✅ | Same CI path-filter gap |
| `bazel-multibuild/flags-parsing-tutorial/` | ✅ | ❌ | ✅ | Legacy WORKSPACE-only module |

### Root monorepo tooling

| Artifact | Path | Version / detail |
|----------|------|------------------|
| Gradle Wrapper | `gradle/wrapper/gradle-wrapper.properties` | 8.10.2 |
| Kotlin (root) | `build.gradle.kts` | 2.3.0, JVM toolchain 23 |
| Makefile | `Makefile` | `bootstrap`, `lint`, `test`, `docs` |
| Gradlew | `gradlew` | Present at root and `rabbit-mq/gradlew` |

---

## README Table Coverage

**Source of truth:** `portfolio.yaml` → injected into `README.md` by `scripts/gen_readme_table.py`.

**In README table (5 projects):**
- `ws-chat-fast`, `projgen`, `rabbit-mq`, `online-bookstore`, `bazel-multibuild`

**Present in repo but missing from README table (~35+ leaf projects):**
- All of `ai-best-practices-examples/*` (5)
- All of `ci-cd-pipelines/*` (6)
- All of `c0de-quality-and-analysis/*` (5)
- All of `dev-env/*` (5)
- All of `dev-ex/*` (4)
- `modular-jvm-build/`, `otel-demo-stack/`, `workflow-api-demo/`, `rest-api-test-demo/`, `observability/`, `forgex/`, `platform-audit-template/`, `tools/`

**Drift enforcement:** `.github/workflows/readme-drift.yml` and docs CI check `scripts/gen_readme_table.py --check` — table stays synced with `portfolio.yaml`, not with all repo folders.

---

## Dependency Health Signals

| Signal | Location | Impact |
|--------|----------|--------|
| Unpinned Python deps | `ws-chat-fast/requirements.txt`, `online-bookstore/src/requirements.txt` | Non-reproducible installs |
| Empty `pyproject.toml` | `online-bookstore/pyproject.toml`, `dev-env/environment-drift-detector/pyproject.toml`, `dev-env/remote-dev-environment-orchestrator/pyproject.toml` | Packaging/build metadata missing |
| Stack label vs code mismatch | `portfolio.yaml` (Redis, Java/Spring for bookstore) vs actual code | Misleading portfolio metadata |
| Gradle version fragmentation | 8.6 / 8.10.2 / 8.14.2 across wrappers | Inconsistent build behavior |
| Kotlin version fragmentation | 2.3.0 (root), 1.9.22 (`modular-jvm-build/`), 1.9.24 (`onboarding-automation-cli/`) | Cross-module compatibility risk |
| Root CI detekt/ktlint without plugins | `.github/workflows/ci.yml` vs `build.gradle.kts` | Kotlin CI job likely fails when triggered |
| Detekt project CI without Gradle | `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/.github/workflows/ci.yml` | `./gradlew test detekt` has no wrapper |
| Bazel CI path filter mismatch | CI watches `BUILD.bazel`; repo uses `BUILD` | Bazel job may never run on changes |
| Deprecated Maven repo | `bazel-multibuild/java/MODULE.bazel` lists `jcenter.bintray.com` | Resolution failures |
| Missing lockfiles | Bazel frontend `pnpm-lock.yaml` referenced but absent; npm project has no `package-lock.json` | Non-reproducible JS builds |
| `:latest` container tags | `observability/docker-compose.yml` | Unpinned infra images |
| Release automation disabled | `.github/workflows/release-please.yml` (`if: false`) | No automated semver releases |
| Root bootstrap missing file | `Makefile` → `pip install -r requirements.txt` | No root `requirements.txt` present |

---

*Stack analysis: 2026-06-22*
