# ForgeX
----

## 1. Product Overview

**Name (Working Title):** ForgeX
**Tagline:** “Spin up production-ready, polyglot repos with one click.”
***Elevator Pitch:***
> ForgeX provides a unified UI + CLI to generate standardized, production-grade project skeletons across multiple languages and runtimes.
> Users select language, frameworks, build system (Gradle or Bazel), CI stack, Docker/K8s packaging, and optional modules (API layer, CLI tool, library, service).
> The system outputs a fully structured repository (monorepo or single-service) with opinionated best practices: code layout, dependency management, lint/format config, test harness, containerization, GitHub Actions (or GitLab CI) pipelines, K8s manifests/Helm chart, optional Observability bootstrap, plus `README` and Architecture doc template.

----

## 2. Core Objectives

1. Consistency: Eliminate hand-rolled divergent scaffolds; enforce conventions across teams.
2. Speed: Sub-minute generation from selection to downloadable artifact or remote Git push.
3. Flexibility: Support multiple stacks while preserving a unified high-level structure.
4. Extensibility: Plugin architecture for new languages, frameworks, build flavors.
5. Idempotency & Reproducibility: Same inputs produce identical tree (hashable manifest).
6. Developer Experience: Minimalistic “push button” web UI + CLI parity + REST API.
7. Governance: Embedded policy checks (license headers, security baseline, supply-chain).
8. Observability Ready: Optional OpenTelemetry + Prometheus wiring for service templates.
9. Build Choice: Bazel or Gradle for JVM / multi-lang; Bazel emphasizes large monorepo; Gradle for per-project simplicity.
10. Container & Deploy: Ready-to-build Dockerfiles + multi-stage builds + sample Helm chart + K8s deployment.

----

## 3. Target Users & Use Cases

| Persona|	Use Case|	Example|
|-------|-----------|----------|
|Platform Engineer	|Introduce standardized microservice skeleton (Go + Bazel + OTel)	|“Create service billing-service with REST + gRPC, Bazel, Helm chart.”|
|Backend Engineer	|Quick Java + Gradle service with Postgres & Testcontainers|	“Generate Spring Boot API + Flyway + GitHub Actions.”
|Data Engineer|	Python library with lint/format/test infra|	“Create internal ETL lib + docs skeleton.”|
|Systems Engineer|	C++ service with Bazel toolchain & sanitizers & Docker	|“Spin up high-performance gateway stub.”|
|Frontend Engineer|	TypeScript library or Node.js API with build & tests	|“Generate TS package w/ ESLint, Vitest, semantic-release.”|
|SRE / DevOps|	Demonstrate best-practice repo to teams|	“Provision polyglot monorepo with Kotlin + Go modules.”|

----

## 4. Functional Requirements

### 4.1 UI / Interaction

|ID|	Requirement|
|---|-------|
|FR-1|	Web UI wizard: Step 1 (Project meta), Step 2 (Language & Build), Step 3 (Modules & Features), Step 4 (Deployment options), Step 5 (Preview & Generate).|
|FR-2|	Instant preview tree & diff panel (virtual file system) before finalize.|
|FR-3|	JSON manifest export capturing all choices.|
|FR-4|	Download zipped archive OR push directly to a Git remote (OAuth tokens).|
|FR-5|	CLI: forgex init --config manifest.yaml to reproduce.|
|FR-6|	REST API: POST /generate with manifest returns archive stream + metadata IDs.|
|FR-7|	Feature toggles: “Observability”, “Security baseline”, “DB integration”, “CI provider”.|
|FR-8|	Form validation + guard rails (e.g. disallow uppercase package names, reserved words).|
|FR-9|	Dark/light theme minimalistic single-page UI with keyboard shortcuts.|

### 4.2 Template / Generation

|ID|	Requirement|
|---|-------|
FR-10	|Language templates: Kotlin (JVM app/library), Java, Python, Go, C/C++, JavaScript (Node), TypeScript (Node/Library).
FR-11|	Build system selection: Gradle (Kotlin DSL) OR Bazel (WORKSPACE, BUILD files).
FR-12|	Polyglot monorepo (Bazel) plus single-project mode (Gradle).
FR-13|	Dependency baseline: pinned versions & constraints; lockfile generation (Gradle versions catalog / Bazel external repos).
FR-14|	Standard directories per language: src/main/kotlin, src/main/java, src/main/python, pkg/ (Go), src/ (TS/JS), src/ + include/ (C/C++).
FR-15|	Generated tests scaffolds: JUnit5, Kotest, Go test, pytest, GoogleTest, Vitest/Jest.
FR-16|	Lint/format: ktlint or detekt, Spotless, golangci-lint, flake8/black/isort, clang-format, ESLint+Prettier.
FR-17|	Optional service features: HTTP REST skeleton, gRPC service (proto + codegen), CLI entrypoint.
FR-18|	Config management stub (e.g., HOCON/Typesafe for JVM, Viper for Go, pydantic for Python).
FR-19|	Security baseline: .editorconfig, license headers, supply chain policy file (OpenSSF scorecard config).
FR-20|	Observability (if selected): OpenTelemetry instrumentation, metrics endpoint, healthz readiness.
FR-21|	Dockerfile(s): multi-stage (build → runtime slim/alpine or distroless), with language-specific optimizations (layer caching).
FR-22|	K8s deployment: Deployment, Service, ConfigMap, HorizontalPodAutoscaler, optional Helm chart or raw Kustomize base.
FR-23|	CI pipeline: GitHub Actions workflows OR GitLab CI; includes build, test, lint, security scan, container build & push, SBOM generation (Syft).
FR-24|	LICENSE & CODEOWNERS & CONTRIBUTING.md scaffolds.
FR-25|	README template auto-filled with build/run instructions based on selected stack.
FR-26|	Architectural Overview doc template plus ADR (Architecture Decision Record) directory.
FR-27|	Generate SBOM script ./scripts/generate-sbom.sh (Syft / CycloneDX).
FR-28|	Option for semantic versioning & release pipeline (semantic-release / conventional commits).
FR-29|	Option for Database integration: Postgres (docker-compose dev service + test container config).
FR-30|	Parameterizable namespace / base package for code (e.g., com.example.billing).

### 4.3 Extensibility & Plugins

|ID|	Requirement|
|---|-------|
FR-31|	Plugin API: register new language template or feature using a defined interface + manifest schema extension.
FR-32|	Hot reload of templates (file watchers) in dev mode.
FR-33|	Versioned template packs; upgrades produce migration plan diff.
FR-34|	Capability introspection endpoint: GET /capabilities returns languages, versions, features.

### 4.4 Governance & Policy

|ID|	Requirement|
|---|-------|
FR-35|	Policy engine (simple rule DSL or OPA) validates manifest (e.g., “All production services must enable Observability”).
FR-36|	Compliance report embedded in output (/docs/compliance-report.md).
FR-37|	Hash of each file with origin template version – stored in .forgex/manifest.lock.

### 4.5 Non-functional

|ID|	Requirement|
|---|-------|
NFR-1|	Single generation < 10s for standard project (90th percentile).
NFR-2|	Concurrency: Handle 50 concurrent generation requests (scale horizontally).
NFR-3|	Deterministic output: same manifest + template version → identical file hashes.
NFR-4|	Latency UI interactions < 100 ms (local).
NFR-5|	Security: No secret leakage; sanitization of user input to prevent path traversal.
NFR-6|	Observability: Structured logs, OpenTelemetry spans for generation phases, metrics (generate_duration_seconds).
NFR-7|	Reliability: 99.9% uptime target (stateless + auto-scaling).
NFR-8|	Maintainability: >85% unit coverage on core generation engine.
NFR-9|	Accessibility: WCAG AA semantics (component alt text, focus states).
NFR-10|	Internationalization-ready (strings externalized).


----

## 5. System Architecture

### 5.1 High-Level Diagram (Text)

```bash
[Browser UI] ---- (HTTPS/JSON) ----> [API Gateway / Backend Service]
   |                                     |
   |  WebSocket / SSE (progress)         |---> Template Loader (Pluggable)
   |                                     |---> Generator Engine
   |                                     |---> Policy Validator
   |                                     |---> Manifest Serializer
   |                                     |---> SCM Integrator (GitHub/GitLab)
   |                                     |
   |                                     +---> Storage (Ephemeral / Object store for archives)
   |
 CLI (forgex) --------------------------> (Same REST API / Offline local mode)
```

### 5.2 Components

|Component	|Description|	Tech Choices|
|---------|-----------------|---------------|
|Web UI|	Single page minimal wizard + preview|	React + TypeScript + Tailwind (sleek minimal), or SvelteKit for lean build|
|API Service	|REST endpoints: manifest validate, generate, capabilities	|Go or Kotlin (fast, typed); choose Go for minimal footprint|
|Template Engine	|Renders file tree from templates + variable substitutions + conditional blocks|	Go text/template + Starlark or Jinja (embedded) + YAML descriptors|
|Plugin Manager|	Discovers language feature packs on startup	|File system + dynamic registry|
Policy Validator|	Manifest evaluation	|OPA (Rego) or custom rule DSL|
|SCM Integrator	|Push created repo to Git remote (GitHub, GitLab)	|OAuth2 flows; go-git for local commit|
|Archive Builder	|Tar/Gzip or Zip packaging	|Streaming writer (no full in-memory)|
|Cache|	Template pack cache keyed by version hash|	Local FS + optional Redis for multi-instance|
|CLI	|Wrap API OR run offline with embedded templates	|Go single static binary|
Metrics & Tracing|	Expose Prometheus metrics + OTEL exporters	|Prometheus client, OTEL SDK|
|Auth (optional)|	API key / OAuth bearer for enterprise mode|	Keycloak / internal|


----

## 6. Data & Manifest Model

### 6.1 Manifest Schema (YAML/JSON)

```yaml
apiVersion: forgeX/v1
kind: ProjectManifest
metadata:
  name: billing-service
  description: "Billing microservice"
  owner: "payments-team"
  templatePackVersion: "2025.07.1"
spec:
  repository:
    layout: single # or monorepo
    vcs:
      provider: github
      visibility: private
      org: acme
  build:
    system: bazel # or gradle
    language: kotlin
    languageVersion: "1.9.24"
    javaCompatibility: "21"
  modules:
    - name: billing-api
      type: service
      features:
        rest: true
        grpc: true
        cli: false
        db: postgres
        observability: true
        securityBaseline: true
        testing:
          unit: true
          integration: true
  deployment:
    docker:
      baseImage: eclipse-temurin:21-jre-alpine
      multiStage: true
      distroless: false
    kubernetes:
      enabled: true
      helm: true
      hpa:
        cpuTargetUtilization: 70
        minReplicas: 2
        maxReplicas: 10
  ci:
    provider: github-actions
    features:
      lint: true
      test: true
      coverage: true
      sbom: true
      release: semantic
  policies:
    allowUnexportedTemplates: false
  security:
    license: Apache-2.0
    codeowners:
      - "payments-team@example.com"
```

### 6.2 Internal Data Structures (Go)

```go
type ProjectManifest struct {
  APIVersion string            `json:"apiVersion"`
  Kind       string            `json:"kind"`
  Metadata   Metadata          `json:"metadata"`
  Spec       Spec              `json:"spec"`
}

type Spec struct {
  Repository RepositorySpec `json:"repository"`
  Build      BuildSpec      `json:"build"`
  Modules    []ModuleSpec   `json:"modules"`
  Deployment DeploymentSpec `json:"deployment"`
  CI         CISpec         `json:"ci"`
  Policies   PolicySpec     `json:"policies"`
  Security   SecuritySpec   `json:"security"`
}
```
(Expand similarly for each; ensure validation tags.)

----

## 7. Template Files Organization

```bash
/templates
  /packs
    /core-kotlin-bazel/...
      descriptor.yaml
      files/
        WORKSPACE.tmpl
        BUILD.bazel.tmpl
        MODULE.bazel.tmpl
        src/main/kotlin/__packagePath__/Application.kt.tmpl
        src/test/kotlin/__packagePath__/ApplicationTest.kt.tmpl
        Dockerfile.tmpl
        charts/app/templates/deployment.yaml.tmpl
        .github/workflows/ci.yml.tmpl
      partials/
        _licenseHeader.tmpl
        _otelInit.tmpl
    /core-go-bazel/...
    /core-python-gradle/ (N/A; python uses different build)
    /core-python-standalone/
    /core-cpp-bazel/
    /core-node-ts/
  /partials-common
  /fragments
    README.md.tmpl
    CONTRIBUTING.md.tmpl
```

**Descriptor Example (descriptor.yaml)**:

```yaml
name: core-kotlin-bazel
version: 2025.07.1
languages: [kotlin]
buildSystems: [bazel]
features:
  rest: templates/rest-controller.kt.tmpl
  grpc: templates/grpc-service.kt.tmpl
  observability: fragments/otel.kt.tmpl
variables:
  - name: package
    required: true
  - name: servicePort
    default: 8080
compatibility:
  kotlin: ">=1.9.0"
```

----

## 8. Generation Pipeline

1. Manifest Validate Phase:
    * JSON schema validation (fast).
    * Policy evaluation (OPA).
2. Template Resolution Phase:
    * Determine base pack by language + build system.
    * Merge feature fragments (dedupe).
3. Variable Resolution Phase:
    * Auto-infer fields (e.g., packagePath from metadata.name).
    * Validate no missing required variables.
4. File Rendering Phase:
    * Depth-first render; apply conditional blocks like:

    ```ruby
        {{ if .Features.Rest }} ... {{ end }}
    ```

    * Run license header injection.

5. Post-Processing Phase:
    * Run formatters (where offline formatting is possible) (e.g., goimports, ktfmt, prettier).
    * Generate dependency locks (Gradle: gradlew dependencies —write-locks; Bazel: bzlmod or fetch with bazel mod dependency).
        * For deterministic generation, lockfiles pre-baked or version pinned in templates – avoid network by vendor snapshot.
6. Hash & Metadata Phase:
    * Compute SHA256 of each file → store in .forgex/manifest.lock.json.
7. Packaging Phase:
    * Create Git commit (initial).
    * Optionally push remote (create repo via provider API).
8. Delivery Phase:
    * Stream archive to client; or respond with remote URL.
9. Audit Logging:
    * Log manifest hash, template pack versions, generation duration, file count.

----

## 9. Build System Details

### 9.1 Bazel Strategy

* Use bzlmod for external dependencies (prefetch pinned versions).
* Language-specific rules:
* Kotlin/Java: rules_kotlin, rules_jvm_external.
* Go: rules_go, gazelle invocation stub.
* Python: rules_python with requirements.txt pinned + lock.
* C/C++: standard cc_library, cc_binary with toolchain detection.
* Node/TS: (Option A) pure Bazel using rules_js; (Option B) keep NPM separate; choose A for uniformity.
* WORKSPACE / MODULE.bazel generated with dictionary of allowed rule versions.
* Provide tools/BUILD for custom lint rules or macros.
* Set up a top-level BUILD.bazel enumerating subpackages.

### 9.2 Gradle Strategy

* Kotlin DSL exclusively.
* Version Catalog gradle/libs.versions.toml centralizing versions.
* Build scans toggle optional.
* Module layout: root + :service, :common, etc.
* Plugin set: java, jacoco, spotless, shadow (if fat jar), com.github.johnrengelman.shadow, org.openapi.generator (optional stub).
* gradlew wrapper pinned.

----

## 10. Language-Specific Template Nuances

Language	Key Additions	Testing	Lint/Format	Docker Optimization
Kotlin/Java	Spring Boot or lightweight Ktor optional; health endpoint; OTel agent attach script	JUnit5 + Testcontainers	Spotless + ktlint	Multi-stage: build with Gradle → run on temurin jre slim
Go	main.go with HTTP mux (chi or stdlib), OTel instrumentation, config via env	go test + example integration test	golangci-lint config	Multi-stage: builder (golang:1.x) → distroless static (if CGO disabled)
Python	CLI or FastAPI skeleton; pyproject.toml (Poetry or Hatch)	pytest + coverage	black + isort + flake8	Build wheel in builder stage → slim python runtime + venv layering
C/C++	Basic service / library skeleton; optional gRPC stub (proto)	GoogleTest	clang-format, clang-tidy config	Multi-stage: build w/ clang container → run in distroless/base
JavaScript/TypeScript	Node service (Fastify/Express) or library; tsconfig.json; ESLint + Prettier	Vitest/Jest	ESLint + Prettier	Multi-stage: build dependencies & compile → copy dist only
Shared	README, CODEOWNERS, LICENSE, ADR folder, Observability	—	Editorconfig	—


----

11. Observability Template
* HTTP Service (any language):
* /healthz (liveness), /readyz (readiness).
* /metrics (Prometheus).
* OpenTelemetry exporter (OTLP endpoint variable).
* Logging: structured JSON (Zap for Go, Logback JSON for JVM, pino for Node, structlog for Python).
* Tracing Setup:
* Exporter config environment variables in deployment.yaml.
* Dashboards:
* Provide grafana/ folder with JSON dashboards (HTTP latency, error rate) optional.

----

12. Security & Compliance Scaffold

Item	Implementation
License Headers	Insert via template partial for code files (supported languages).
SBOM	Script using Syft (syft dir:./ -o cyclonedx-json => sbom.json).
SAST Hooks	GitHub Actions step (CodeQL optional toggle).
Dependency Scans	trivy fs or grype step in CI.
Supply Chain Policy	.github/workflows/policy.yml guard for approved dependency hosts.
Codeowners	Provided for auto-review gating.
Commit Conventions	Optional commitlint.config.js to enforce conventional commits.


----

13. Policy Engine Examples (Rego)

package policies

deny[msg] {
  input.spec.deployment.kubernetes.enabled == false
  input.spec.modules[_].type == "service"
  msg = "Service modules must enable Kubernetes deployment"
}

deny[msg] {
  some m
  m = input.spec.modules[_]
  m.features.observability != true
  m.type == "service"
  msg = sprintf("Service %s must enable observability", [m.name])
}

If deny non-empty → generation blocked (return 422 with cause list).

----

## 14. CLI Design

Command Surface:

```bash
forgex init           # interactive manifest wizard (terminal)
forgex validate -f manifest.yaml
forgex generate -f manifest.yaml -o ./out
forgex generate --remote-push --git-token $TOKEN
forgex list-packs
forgex upgrade -f manifest.yaml --to-version 2025.08.0
forgex diff -f manifest.yaml --against-lock
forgex template add custom-pack/

Flags: --json for machine readable results, --quiet, --no-format.
```

----

15. Internal Module Breakdown (Backend)

Module	Responsibility
cmd/server	Main entrypoint
internal/http	REST routes, handlers
internal/manifest	Parsing, validation, defaults
internal/policy	OPA wrapper
internal/templates	Load / index template packs
internal/generator	Orchestration of render pipeline
internal/formatter	Post-generation format hooks
internal/scm	Git init, commit, push
internal/archive	Zip/Tar streaming
internal/plugins	Plugin registry & lifecycle
internal/metrics	Prometheus & tracing instrumentation
internal/security	Input sanitization, license insertion
internal/hash	File hashing & lockfile writer
internal/versioning	Pack version compatibility
internal/upgrade	Migration diff engine
internal/log	Logging utilities


----

16. REST API Endpoints

Method	Path	Purpose	Auth
GET	/health	Liveness	None
GET	/ready	Readiness	None
GET	/capabilities	List languages, packs, versions	None
POST	/validate	Validate manifest; returns issues	Optional
POST	/generate	Accept manifest; streams archive; includes response metadata JSON frame	Optional
POST	/upgrade/plan	Compute migration diff manifest→new pack version	Auth
POST	/push	Generate & push repo remotely	Auth (SCM token)
GET	/history/{id}	Retrieve metadata (audit)	Auth
GET	/packs/{name}/{version}	Show descriptor	None

Streaming payload framing: first JSON header (jobId, hash), then binary archive chunked.

----

## 17. Error Handling

|Category	|HTTP|	Example|
|----|--|--|
|Validation Error|	422	|Missing required field spec.build.language
|Policy Violation|	403|	Service billing-api must enable observability|
|Internal|	500|	Template render panic|
|Unsupported Feature|	400|	gRPC not supported in selected language|
|SCM Push Failure	|502|	GitHub API error|

Return JSON with errorCode, message, optional details array.

----

## 18. Performance Considerations

* Template Rendering: Pre-parse templates into AST on startup; store compiled templates keyed by pack version.
* Concurrency: Use worker pool (bounded) for post-processing tasks (formatting, hashing).
* Caching: If manifest hash already generated (and ephemeral + no dynamic timestamps), allow “cache hit” reuse (optionally disabled if customizing).
* I/O: Stream compression on the fly with gzip.Writer + io.Pipe.

----

## 19. Determinism Strategies

* Avoid using current timestamp inside generated code except in metadata disclaimers wrapped with an option --include-timestamp.
* License header contains static year or manifest year (provided).
* Order file enumeration lexicographically.

----

## 20. Testing Strategy

Layer	Approach
Unit	Manifest parser, policy engine wrapper, template variable resolution, path sanitization.
Snapshot / Golden Tests	Sample manifests produce file tree whose hashes match golden set.
Integration	End-to-end API call with manifest, verify archive content & lockfile correctness.
Concurrency	Simultaneous 50 generate requests: ensure isolation and throughput metrics.
CLI Tests	Using internal fixture repos to assert correct exit codes & output.
Security	Fuzz relative paths, environment injection attempts.
Upgrade Tests	Manifest upgrade plan yields expected diff actions (add/remove/modify).
Performance	Benchmark rendering for large multi-module monorepo (e.g., 10 modules, 5 languages).
Policy	Policy violation test matrix (deny cases).


----

21. Metrics & Observability

Metric	Type	Labels	Description
generation_requests_total	Counter	status	Count of generation calls
generation_duration_seconds	Histogram	pack, buildSystem	Duration
generation_output_files	Histogram	pack	File count per generation
policy_violations_total	Counter	policy_id	Violations
template_cache_hits_total	Counter	pack	Cache usage
error_total	Counter	type	Errors by category

Tracing spans: ValidateManifest, ResolveTemplates, RenderFiles, FormatFiles, HashFiles, PackageArchive, SCMPush.

Logs: structured JSON with correlation requestId, jobId.

----

22. Security Hardening

Aspect	Control
Input Sanitization	Reject paths containing .., absolute root anchors, or invalid characters.
Rate Limiting	Per IP token bucket (gateway).
Auth (optional)	API Keys hashed at rest; SCM tokens ephemeral memory only.
Dependency Trust	Template pack signature (SHA256 + optional GPG).
Supply Chain	Build container pinned digest; verify base image digests in Dockerfile templates.
Code Injection	Template evaluation restricted to whitelisted functions (no arbitrary exec).
Secrets	No secret values inserted automatically; environment placeholders only.


----

23. Upgrade / Migration Mechanism
* forgex upgrade -f manifest.yaml --to-version X:
* Compute diff between current lock’s pack versions and target pack descriptor.
* Generate plan:

actions:
  - modify: file path=build.gradle.kts reason="Dependency version bump"
  - add: file path=.github/workflows/security.yml
  - remove: file path=legacy-script.sh


* Option --apply updates repo in place & writes new lock.

----

## 24. File Lock Format

`/.forgex/manifest.lock.json`:

```json
{
  "manifestHash": "sha256:...",
  "templatePackVersions": {
    "core-kotlin-bazel": "2025.07.1"
  },
  "files": [
    {"path":"README.md","sha256":"..."},
    {"path":"WORKSPACE","sha256":"..."}
  ],
  "generatedAt": "2025-07-18T12:00:00Z",
  "buildSystem": "bazel",
  "language": "kotlin"
}
```

----

## 25. ADR Template

`/docs/adr/0001-record-architecture-decisions.md`

```markdown

# 0001 - Record Architecture Decisions
Status: Accepted
Date: 2025-07-18
Context:
Decision:
Consequences:
```

----

26. Example Generated Go Service (Core Files – abbreviated)

```bash
.
├── README.md
├── go.mod
├── go.sum
├── cmd/billing-api/main.go
├── internal/config/config.go
├── internal/handler/health.go
├── internal/otel/init.go
├── internal/log/logger.go
├── Dockerfile
├── Makefile
├── k8s/deployment.yaml
├── k8s/service.yaml
├── .github/workflows/ci.yml
└── .forgex/manifest.lock.json
```
main.go skeleton:

```go
func main() {
  ctx := context.Background()
  cfg := config.Load()
  logger := log.New(cfg.LogLevel)
  tp := otel.Init(cfg) // sets up metrics + tracing if enabled
  defer func() { _ = tp.Shutdown(ctx) }()
  r := http.NewServeMux()
  r.Handle("/healthz", handler.Health())
  r.Handle("/readyz", handler.Ready())
  server := &http.Server{ Addr: fmt.Sprintf(":%d", cfg.Port), Handler: middleware.Instrument(r) }
  logger.Info("starting server", "port", cfg.Port)
  log.Fatal(server.ListenAndServe())
}
```

----

## 27. Task Breakdown & Estimates

Legend: (hrs = focused engineering hours). Priorities P0 (critical), P1 (high), P2 (normal).

Epic A: Foundation

ID	Task	Est	Owner	Priority
A1	Repo scaffold (backend + UI + templates folders)	4h	SR	P0
A2	Backend server skeleton (health, ready)	4h	JR1	P0
A3	Manifest schema + parser + validation	8h	JR1	P0
A4	Policy engine integration (OPA)	6h	JR2	P0
A5	Template pack loader & indexing	10h	JR2	P0
A6	Generator orchestration pipeline	12h	SR	P0
A7	Hash & lockfile writer	4h	JR1	P1
A8	Archive streaming endpoint	6h	JR2	P1
A9	Metrics/tracing instrumentation	6h	JR2	P1

Epic B: Templates (Baseline)

ID	Task	Est
B1	Kotlin Bazel pack	12h
B2	Kotlin Gradle pack	10h
B3	Go Bazel pack	10h
B4	Go Standalone (go mod) pack	6h
B5	Python pack (pyproject / Poetry)	8h
B6	C/C++ Bazel pack	12h
B7	JS/TS Node pack (ESLint+Prettier+Vitest)	10h
B8	Shared observability fragments	6h
B9	Security baseline fragments	6h
B10	Docker & K8s fragments	8h
B11	Helm chart template	6h
B12	CI workflows (GitHub)	6h

Epic C: Feature Extensions

|ID|	Task|	Est|
|----|----|------|
|C1|	REST service fragments (Go, Kotlin, TS, Python)	|12h|
|C2|	gRPC fragments (proto + build integration)	|10h|
|C3|	DB integration (Postgres + Testcontainers)	|10h|
|C4|	CLI module fragment (Cobra for Go, Picocli for JVM)|	8h|
|C5|	Observability instrumentation code injection	|6h|
|C6|	Release automation (semantic-release / tagging)	|6h|

Epic D: Web UI

|ID|	Task|	Est|
|---|----|----|
|D1|	React skeleton + routing + theme	6h|
|D2|	Manifest builder forms (multi-step)	12h|
|D3|	Preview file tree panel (virtual FS)	10h|
|D4|	Diff viewer (Monaco diff)	6h|
|D5|	Progress & streaming download (WebSocket)	6h|
|D6|	OAuth integration (SCM)	8h|
|D7|	Accessibility & keyboard nav	6h|
|D8|	Unit & e2e tests (Playwright/Cypress)	10h|

Epic E: CLI

|ID|	Task	Est|
|E1|	CLI skeleton (cobra)	6h|
|E2|	Offline embedded templates packaging	8h|
|E3|	Generate, validate, diff commands	10h|
|E4|	Upgrade command & diff viewer	8h|
|E5|	Auth integration (API token)	4h|

Epic F: Governance & Policy

|ID|	Task	Est|
|F1|	Default policy set	4h|
|F2|	Policy override config file support	4h|
|F3|	Policy enforcement tests	4h|

Epic G: Testing & Quality

|ID|	Task	|Est|
|---|----|---|
|G1|	Golden output snapshots (per pack)	|10h|
|G2|	Performance test harness	|8h|
|G3|	Security fuzz tests (path traversal)	8h|
|G4|	Coverage gating pipeline	4h|

Epic H: Documentation

|ID|	Task|	Est|
|---|----|---|
|H1|	User Guide (UI + CLI)	8h|
|H2|	Developer Guide (template authoring)	8h|
|H3|	API Reference (OpenAPI)	6h|
|H4|	Upgrade & versioning doc	4h|
|H5|	ADR for generator architecture	4h|

Epic I: Release & Ops

|ID|	Task|	Est|
|---|----|---|
|I1|	Containerization backend & UI	4h|
|I2|	Helm chart for ForgeX	6h|
|I3|	Horizontal scaling & readiness tests	6h|
|I4|	Observability dashboards (Grafana)	4h|
|I5|	On-call runbook	4h|

Approx Total: ~ (Aggregate ~ 350–380 hours) – ~8–9 weeks with 1 SR + 2–3 JR devs.

----

## 28. Acceptance Criteria (Selected)

Feature	Criterion
Deterministic Generation	Same manifest yields identical lockfile hashes across runs.
Multi-language Support	Each base language pack passes lint & test out-of-the-box (CI green).
Bazel Pack	bazel build //... succeeds; bazel test //... runs skeleton tests.
Gradle Pack	./gradlew build passes; version catalog resolves.
Observability Option	Service exposes /metrics and trace spans appear in collector (if OTEL env set).
Policy Violation	Manifest missing required observability for service returns 403 with message.
CLI Offline	forgex generate -f manifest.yaml works without network (using embedded templates).
Git Push	Generated repo appears in configured GitHub org with initial commit & workflows.
Performance	90% of single-module skeleton generations complete < 5s; 95% < 10s.
Upgrade Plan	Upgrading pack version produces actionable diff (≥ The changed files enumerated).


----

## 29. Risk Register

Risk	Impact	Probability	Mitigation
Template Sprawl	Maintenance burden	Medium	Versioned packs, deprecate old, docs for authorship.
Non-determinism (format tool variation)	Noisy diffs	Medium	Pin formatter versions, containerized formatting.
Bazel dependency resolution latency	Slower generation	Medium	Pre-baked external repos; offline module archive.
Security Misconfig (injected path)	File system escape	Low	Sanitize, enforce whitelisted relative roots.
Policy bypass (custom manifest edits)	Non-compliant repos	Low	Re-run policy in CI (provided workflow).
Scaling concurrency (compression CPU)	Latency spikes	Medium	Offload compression to worker pool; optional queue.
SCM API rate limits	Push failures	Medium	Backoff & token caching; manual fallback download.
Overgrown features in UI	Complexity	Low	Keep simple wizard + advanced YAML mode toggle.


----

## 30. Future Enhancements (Post v1)
* Multi-framework expansions (Spring Boot / Micronaut toggles).
* UI plugin marketplace for internal templates.
* Multi-repo orchestration (service mesh bootstrap).
* Code generation for OpenAPI / AsyncAPI specs.
* Internal dependency graph visualizer.
* Integration with secret management (Vault placeholders).
* Migrations CLI (apply delta when pack updates).

----

## 31. Developer Onboarding (Internal)
1.	Clone repo: git clone git@.../forgex.
2.	Install Go 1.x, Node 20+, Bazel, Docker.
3.	make bootstrap (installs linters, pre-commit hooks).
4.	Run backend: make run-backend.
5.	Run UI: make run-ui (Vite dev server).
6.	curl localhost:8080/capabilities to verify.
7.	Generate sample: curl -X POST /generate -d @examples/kotlin-bazel.json -o out.zip.
8.	Unzip & run bazel test //....
9.	Add new template pack (example) & run golden test update: make update-goldens.

----

## 32. Makefile / Scripts Outline
```makefile
bootstrap: install-go-tools install-node-deps
install-go-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59
    go install github.com/bufbuild/buf/cmd/buf@latest
run-backend:
    go run ./cmd/server
run-ui:
    npm --prefix ui run dev
test:
    go test ./... -cover
lint:
    golangci-lint run
golden:
    go test ./test/golden -update
build-cli:
    go build -o bin/forgex ./cmd/cli
docker-build:
    docker build -t forgex-backend:latest .
```

----

33. Sample GitHub Actions Workflow (Generated)

.github/workflows/ci.yml.tmpl:

```yml
name: CI
on:
  push: { branches: [main] }
  pull_request:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup
        uses: actions/setup-java@v4
        with: { distribution: temurin, java-version: 21 }
      - name: Cache Gradle
        uses: actions/cache@v4
        with:
          path: ~/.gradle/caches
          key: gradle-${{ hashFiles('**/*.gradle*','**/gradle/libs.versions.toml') }}
      - run: ./gradlew build --scan --no-daemon
      - name: Lint
        run: ./gradlew spotlessCheck
      - name: SBOM
        run: ./scripts/generate-sbom.sh
      - name: Docker Build
        run: docker build -t ghcr.io/${{ github.repository }}/app:${{ github.sha }} .
```

----

34. Extensibility (Plugin Interface Pseudocode)
```go
type TemplatePack interface {
  Name() string
  Version() string
  Supports(manifest Spec) bool
  Variables() []VariableDescriptor
  Render(ctx Context, manifest ProjectManifest) ([]GeneratedFile, error)
  Features() []FeatureDescriptor
}

Register via:

func init() {
  registry.Register(&KotlinBazelPack{})
}
```

----

## 35. File Naming & Variable Substitutions
* Replace __packagePath__ with strings.ReplaceAll(packageName, ".", "/").
* Placeholders inside filenames: Application__ServiceName__.kt.tmpl ⇒ after substitution ApplicationBilling.kt.
* Template snippet condition:

{{- if eq .Build.System "bazel" -}}
# Bazel-specific content
{{- end -}}


----

## 36. Content Hash Strategy

Compute stable hash: sha256(path + '\n' + fileContentNormalized). Normalize line endings to \n and remove trailing whitespace lines for determinism.

----

## 37. End-to-End Example (User Flow)
1.	User opens web UI: selects Go, Bazel, Observability ON, REST ON, DB OFF, gRPC ON.
2.	Wizard shows preview: ~24 files.
3.	User clicks Generate.
4.	Backend logs:
    * Start job jobId=abc.
    * Validate (30ms) → pass.
    * Resolve templates (15ms).
    * Render (150ms).
    * Format (goimports – 50ms).
    * Hash + lock (20ms).
    * Archive stream begins (~300ms total).
5.	User downloads billing-service.zip.
6.	Locally runs bazel test //..., passes.
7.	Pushes repo; CI runs green.

----

## 38. Quality Gates & Definition of Done (Project)

Area	Gate
Backend Coverage	≥ 85% (excluding main.go, generated code)
UI E2E Tests	Coverage of primary wizard path & error path
Template Golden Tests	Each pack’s canonical manifest has golden snapshot
Performance	P95 generation < 10s for largest baseline template (monorepo w/ 5 modules)
Security	SAST scan no high severity
Lint	Zero lint errors in core modules
Docs	User + Dev + API docs complete
Observability	Metrics & tracing present in production env


----

## 39. Final Deliverables (v1 Release)
1.	Backend service container image.
2.	UI static build.
3.	CLI binary (multi-platform).
4.	Six language template packs (Kotlin, Java, Go, Python, C/C++, JS/TS).
5.	Observability and security fragment packs.
6.	Policy set & example overrides.
7.	Documentation set (/docs).
8.	Helm chart to deploy ForgeX itself.
9.	Example manifests for each language.
10.	Golden test suite & CI pipelines.

----

## 40. Summary

ForgeX unifies scaffolding for a polyglot ecosystem through deterministic templating, policy enforcement, extensible packs, and seamless delivery across UI + CLI + API. This specification defines in detail every component required for a robust initial launch, with explicit tasks enabling parallelization and straightforward delegation to junior developers.

----
