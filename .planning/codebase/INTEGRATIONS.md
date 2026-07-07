# External Integrations

**Analysis Date:** 2026-06-22

## APIs & External Services

**LLM / AI:**
- OpenAI API — Optional test generation in `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/adapters/llm_adapter.py`
  - SDK/Client: `openai>=1.30.0` (optional extra in `ai-best-practices-examples/ai-code-assistant/pyproject.toml`)
  - Auth: `OPENAI_API_KEY` env var
- ComfyUI HTTP API — Image/video workflow execution in `ai-best-practices-examples/ai-image-video-generator/src/ai_image_video_generator/pipelines/comfy_client.py`
  - SDK/Client: `requests` via custom `ComfyClient`
  - Auth: None (local service)
  - Config: `AIVG_COMFYUI_BASE_URL` (default `http://127.0.0.1:8188` in `ai-best-practices-examples/ai-image-video-generator/src/ai_image_video_generator/config.py`)
- Ollama CLI — Local inference for domain expert demo in `ai-best-practices-examples/domain-expert-ai/src/domain_expert_ai/inference/serve_ollama.py`
  - SDK/Client: subprocess `ollama run`
  - Auth: None (local)
- Hugging Face — Embeddings for RAG in `ai-best-practices-examples/knowledge-qa-system/src/app/retrieval/embeddings.py`
  - SDK/Client: `langchain-huggingface`, `sentence-transformers`
  - Auth: HF token not required for default local models; network may be needed on first download

**GitHub:**
- GitHub CLI (`gh`) — PR metadata ingestion in `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/github_ingest.py`
  - Auth: Implicit `gh` session / `GITHUB_TOKEN` for CLI
- GitHub REST (conceptual) — CI/CD pipeline tools reference GitHub Actions billing, PR timestamps, deployment events (`ci-cd-pipelines/README.md`, `ci-cd-pipelines/release-lead-time-calculator/`, `ci-cd-pipelines/pipeline-cost-analyzer/`)
  - SDK/Client: Not implemented as live API clients; bootstrap Python scripts with fixture-oriented design
- GitHub Actions — Primary CI/CD platform (`.github/workflows/ci.yml`, per-project workflows under `c0de-quality-and-analysis/*/.github/workflows/`, `otel-demo-stack/.github/workflows/build.yml`, `rest-api-test-demo/.github/workflows/build.yml`)
- GitHub Pages — Static site deployment (`.github/workflows/static.yml`, `.github/workflows/pages.yml`)

**Static analysis / security tooling (inputs, not hosted SaaS):**
- Semgrep SARIF — Parsed in `c0de-quality-and-analysis/security-hotspot-annotator/src/input/semgrep_parser.py`
- OpenAPI specs — Diffed in `c0de-quality-and-analysis/api-breaking-change-detector/src/diff/openapi_diff.py`
- Gradle dependency trees — Parsed in `c0de-quality-and-analysis/license-compliance-scanner/src/ingest/gradle_tree_parser.py`
- Detekt — Custom rules in `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/` (local Gradle/detekt integration intended)

**Weather (demo stub):**
- Tool invocation in `ai-best-practices-examples/chat-ai/src/ai_app/tools/weather.py` — Local/demo tool, not a documented external API integration

## Data Storage

**Databases:**
- PostgreSQL 16 — Workflow job results store
  - Connection: `DATABASE_URL` env var (default `postgresql://postgres:postgres@db:5432/workflow` in `workflow-api-demo/worker/worker.py`)
  - Client: `asyncpg` (`workflow-api-demo/api/requirements.txt`, `workflow-api-demo/worker/requirements.txt`)
  - Orchestration: `workflow-api-demo/docker-compose.yml` (`postgres:16-alpine` service)
- SQLite/CSV (file-backed) — Online bookstore uses CSV files, not PostgreSQL despite README label
  - Files: `online-bookstore/src/books.csv`, `online-bookstore/src/orders.csv`
  - Client: Custom Python in `online-bookstore/src/operations.py`
- ChromaDB — Vector store for knowledge Q&A
  - Connection: Local path via settings/tests (`KQ_CHROMA_PATH` in `ai-best-practices-examples/knowledge-qa-system/tests/conftest.py`)
  - Client: `langchain-chroma`, `chromadb>=1.0.0` (`ai-best-practices-examples/knowledge-qa-system/pyproject.toml`)
- JSON file memory — Chat AI conversation memory
  - File: `chroma_data/memory.json` path in `ai-best-practices-examples/chat-ai/src/ai_app/services/conversation.py`
  - Client: `ai-best-practices-examples/chat-ai/src/ai_app/services/memory.py`

**Message queues:**
- Redis 7 — Job queue for workflow demo
  - Connection: `REDIS_URL` (default `redis://redis:6379` in `workflow-api-demo/worker/worker.py`)
  - Client: `redis.asyncio` / `redis>=5.0.0`
  - Orchestration: `workflow-api-demo/docker-compose.yml`
- RabbitMQ — AMQP messaging demo
  - Client: `com.rabbitmq:amqp-client:5.25.0` (`rabbit-mq/build.gradle`), Python `pika` in `rabbit-mq/app/src/main/python/`
  - Broker: External/local broker (host `localhost` in `rabbit-mq/app/src/main/python/channel.py`); no compose file in `rabbit-mq/` root

**File Storage:**
- Local filesystem — Generated assets, audit outputs, SBOMs, reports (`c0de-quality-and-analysis/license-compliance-scanner/src/output/sbom_writer.py`, `ai-best-practices-examples/ai-image-video-generator/` export paths)
- No cloud object storage (S3/GCS/Azure Blob) integration detected

**Caching:**
- Redis — Queue only in `workflow-api-demo/` (not used as cache layer)
- In-memory WebSocket connection state — `ws-chat-fast/app/ws_manager.py`
- None as dedicated cache service elsewhere

## Authentication & Identity

**Auth Provider:**
- Custom / demo — No centralized IdP
  - FastAPI OAuth2 password flow with fake user DB: `online-bookstore/src/security.py`
  - WebSocket password bearer for exclusive rooms: `ws-chat-fast/app/ws_password_bearer.py`, `ws-chat-fast/app/security.py`
  - Keycloak mentioned as devcontainer feature concept only (`dev-env/devcontainer-feature-library/README.md`); not wired in runnable compose stacks

## Monitoring & Observability

**Tracing & metrics:**
- OpenTelemetry SDK + OTLP HTTP exporter — `otel-demo-stack/api/main.py`, `otel-demo-stack/worker/worker.py`
  - Collector: `otel/opentelemetry-collector-contrib:0.96.0` (`otel-demo-stack/docker-compose.yml`)
  - Env: `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`
- OpenTelemetry API (partial) — `ai-best-practices-examples/knowledge-qa-system/pyproject.toml` (`opentelemetry-api>=1.27.0`)
- CI pipeline span emission (stdout bootstrap) — `ci-cd-pipelines/pipeline-telemetry-exporter/src/main.py`

**Metrics & dashboards:**
- Prometheus — `observability/docker-compose.yml`, config `observability/prometheus.yml`
- Grafana — `observability/docker-compose.yml`, provisioning in `observability/grafana/provisioning/`
- Loki — Log aggregation container in `observability/docker-compose.yml`

**Error Tracking:**
- None (no Sentry, Datadog APM, or similar SaaS integration)

**Logs:**
- stdout/stderr — Default for Python services and CI tools
- Structured OTel spans — `otel-demo-stack/`
- Design doc references Prometheus exporter + Redis history for flaky tests (`src/main/resources/dd-ci-cd-pipeline-with-flaky-test-detection.md`) — not implemented as runnable code in `ci-cd-pipelines/flaky-pipeline-gate/`

## CI/CD & Deployment

**Hosting:**
- GitHub Pages — Repo-wide static deploy (`.github/workflows/static.yml` uploads entire repo; `.github/workflows/pages.yml` builds MkDocs to `site/`)
- Local Docker — Primary runtime for multi-service demos

**CI Pipeline:**
- Root GitHub Actions — `.github/workflows/ci.yml`
  - Path-filtered jobs: Python (ruff, mypy, pytest, Codecov), Kotlin (gradle ktlint/detekt/test), Bazel (bazelisk build), docs (mkdocs + README table check)
- CodeQL — `.github/workflows/codeql.yml` (Python + Java, weekly schedule)
- README drift — `.github/workflows/readme-drift.yml`
- Release Please — `.github/workflows/release-please.yml` (disabled via `if: false`)
- Per-subproject CI — `c0de-quality-and-analysis/*/.github/workflows/ci.yml`, `otel-demo-stack/.github/workflows/build.yml`, `rest-api-test-demo/.github/workflows/build.yml`
- Codecov — Upload in root Python CI job (`.github/workflows/ci.yml`)

**Container orchestration:**
- Docker Compose — `workflow-api-demo/docker-compose.yml`, `otel-demo-stack/docker-compose.yml`, `rest-api-test-demo/docker-compose.yml`, `observability/docker-compose.yml`, `dev-env/local-service-mesh/docker-compose.yml`
- Bazel OCI rules — Pull distroless Java image in `bazel-multibuild/java/MODULE.bazel` (`gcr.io/distroless/java:17`)

**Dev environments:**
- Dev Containers — `.devcontainer/devcontainer.json` (Ubuntu, docker-in-docker, `make bootstrap`)
- Devcontainer features catalog — `dev-env/devcontainer-feature-library/` (otel-collector, local-kafka, local-postgres-seed, local-keycloak features documented)

## Environment Configuration

**Required env vars (by integration area):**

| Variable | Used by | Purpose |
|----------|---------|---------|
| `OPENAI_API_KEY` | `ai-code-assistant` | OpenAI test generation |
| `AIVG_COMFYUI_BASE_URL` | `ai-image-video-generator` | ComfyUI endpoint |
| `REDIS_URL` | `workflow-api-demo` | Job queue |
| `DATABASE_URL` | `workflow-api-demo` | PostgreSQL connection |
| `OTEL_SERVICE_NAME` | `otel-demo-stack` | Service identity for traces |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-demo-stack` | OTLP collector URL |
| `KQ_CHROMA_PATH` | `knowledge-qa-system` tests | Local Chroma persistence |
| `GITHUB_TOKEN` | `gh` CLI / Actions | GitHub API access (implicit in CI) |

**Secrets location:**
- Local environment / CI secrets — Not committed; `.env` patterns referenced in docs only
- Docker Compose inline dev credentials — `workflow-api-demo/docker-compose.yml` (postgres user/password for local demo only)
- Forbidden from docs: never commit `.env`, credentials files, or keys (see repo `.gitignore` patterns)

## Webhooks & Callbacks

**Incoming:**
- FastAPI HTTP endpoints — Job submission in `workflow-api-demo/api/`, health/OTel checks in `otel-demo-stack/api/main.py`, chat/WebSocket in `ws-chat-fast/app/`, bookstore REST in `online-bookstore/src/main.py`
- None as dedicated webhook receivers for GitHub/Stripe/etc.

**Outgoing:**
- OTLP HTTP export — Spans/metrics to OTel Collector (`otel-demo-stack/api/main.py`, `otel-demo-stack/worker/worker.py`)
- ComfyUI POST — `/api/generate` in `ai-best-practices-examples/ai-image-video-generator/src/ai_image_video_generator/pipelines/comfy_client.py`
- OpenAI chat completions — When `OPENAI_API_KEY` set (`ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/adapters/llm_adapter.py`)
- GitHub PR comments (planned pattern) — Payload builder in `c0de-quality-and-analysis/security-hotspot-annotator/src/publish/pr_comment_client.py` (formats comment body; no live GitHub API client in tree)
- CI span stdout — `ci-cd-pipelines/pipeline-telemetry-exporter/src/main.py` (bootstrap target for future collector)

## Integration Coverage by Portfolio Area

| Area | External systems | Maturity |
|------|------------------|----------|
| `workflow-api-demo/` | Redis, PostgreSQL, Docker | Runnable compose stack |
| `otel-demo-stack/` | OTel Collector, OTLP | Runnable compose stack |
| `observability/` | Prometheus, Grafana, Loki | Runnable compose stack |
| `rabbit-mq/` | RabbitMQ AMQP | Client code; broker external |
| `ai-best-practices-examples/` | OpenAI, ComfyUI, Ollama, Chroma, HuggingFace | Optional/local; env-dependent |
| `ci-cd-pipelines/` | GitHub Actions (conceptual), OTel (stdout) | Demo/stub integrations |
| `c0de-quality-and-analysis/` | Semgrep, OpenAPI, Gradle trees, GitHub PR comments (planned) | Offline/script-driven |
| `ws-chat-fast/` | None persistent (in-memory WS) | No Redis despite portfolio label |
| `online-bookstore/` | None (CSV files) | No PostgreSQL despite portfolio label |
| `forgex/`, `platform-audit-template/` | None runnable | Documentation/templates only |

## Not Detected

- Cloud providers (AWS, GCP, Azure SDKs)
- Stripe, Supabase, Firebase
- Kafka runtime (mentioned in devcontainer feature docs only)
- Kubernetes deployment manifests (Helm/K8s referenced in `forgex/README.md` design only)
- Datadog, Sentry, New Relic
- Slack/Discord/PagerDuty alerting webhooks
- Managed auth (Auth0, Clerk, Keycloak runtime)

---

*Integration audit: 2026-06-22*
