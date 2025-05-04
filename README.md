# Portfolio

I'm using this to house my portfolio projects. 


There should be 10 of them defined in design documents under `src/main/resources/`:
 - dd-build-system-expertise.md
 - dd-ci-cd-pipeline-with-flaky-test-detection.md
 - dd-cloud-native-microservices.md
 - dd-full-stack-sample-app.md
 - dd-developer-productivity-tooling.md
 - dd-cross-plat-nortarization-tool.md
 - dd-advanced-logging-and-tracing.md
 - dd-gitops-starter-kit.md
 - dd-open-source-plugin-contribution.md
 - dd-performance-benchmarking-n-tuning.md
 - dd-streaming-data-demo.md

## Cli Scaffolding Tool Design

### 1. Purpose

Create a Python-based CLI tool (e.g. projgen) to scaffold new projects with Bazel and/or Gradle, providing opinionated defaults for:
* Directory layout 
* CI integration 
* Test framework configuration 
* Linting/config 
* README/license/gitignore boilerplate 
* Multi-language build support 
* Observability instrumentation

### 2. Technology Stack

* Language: Python 3.10+ 
* CLI Framework: Click 
* Templating: Jinja2 
* Packaging: Poetry or setuptools

### 3. CLI UX & Commands

`projgen init <project-name>`

#### Options:

```
--build [bazel|gradle|both] (default: both)
--language <lang> (required) # [java, kotlin, groovy, cpp, c, python, rust, node, typescript]
--license <license-id> (default: MIT)
--ci <ci-provider> (default: github)
```

#### Flags:
```
--overwrite (replace existing files)
--no-telemetry (disable usage reporting)
```
### 4. Opinionated Defaults

#### Directory Layout

Gradle:
```
src/main/<lang>/
src/test/<lang>/
```
Bazel:
```
WORKSPACE
src/
BUILD.bazel
//… packages per feature
```
## CI Integration

GitHub Actions workflow (`.github/workflows/ci.yml`) running `bazel test` and/or `./gradlew check`

## Test Framework

Java/Kotlin/Groovy: JUnit 5, Kotest, Spock

Python: unittest & pytest

Node/TypeScript: Jest

Rust: cargo test

C/C++: Google Test, CppUnit

**Linting/Config**

Gradle: Spotless + Checkstyle configs

Bazel: buildifier + bazelisk wrapper script

Python: black + flake8

Node: eslint + prettier

#### Boilerplate Files

README.md with project name, description placeholder

LICENSE (chosen license template)

.gitignore tuned for Bazel, Gradle, language

## 5. Multi‑Language Support

Each language stub includes:

Sample “Hello World” code

Build file fragment with minimal dependencies

Test stub exercising the sample code

## 6. Observability Instrumentation

Include basic OpenTelemetry setup in stub code

Provide telemetry/ directory with per-language configuration and docs

Example: exported traces to a Zipkin endpoint or Prometheus metrics endpoint

## 7. Internal Architecture
```
projgen/
├── cli.py            # Click entrypoint
├── generators/       # Modules for each build system & language
├── templates/        # Jinja2 templates
├── config.py         # Defaults & mappings
└── telemetry.py      # Optional usage reporting
```
## 8. Example Usage
```
projgen init my-app \
--build both \
--language kotlin \
--license Apache-2.0 \
--ci github
```
#### Generates:
```
my-app/
├── README.md
├── LICENSE
├── .gitignore
├── WORKSPACE
├── BUILD.bazel
├── src/main/kotlin/com/example/App.kt
├── src/test/kotlin/com/example/AppTest.kt
├── build.gradle.kts
├── settings.gradle.kts
└── .github/workflows/ci.yml
```
## 9. Next Steps

- [ ] Validate UX flow & flags 
- [ ] Create Jinja2 templates for each combination 
- [ ] Implement generator modules & CLI wiring 
- [ ] Write unit tests & integration tests 
- [ ] Publish to PyPI and add project README

### to install and test
```shell
    pip install -e .
```
