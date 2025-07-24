

# Build-System Expertise Project: Design Document

## 1. Overview

Purpose:
Build a demonstration monorepo featuring Kotlin, C++, and Python modules, integrating custom Bazel rules, and a Gradle-to-Bazel migration tool, to showcase mastery over complex build system design, language interoperability, and build automation.

⸻

## 2. Goals and Objectives
*	Custom Build Rules: Write Starlark rules for compiling, testing, and packaging targets in Kotlin, C++, and Python—supporting mixed-language dependencies.
*	Migration Tool: Automate migration of Gradle projects to Bazel, including test suite conversion and artifact preservation.
*	Sanitizer Integration: Integrate AddressSanitizer (ASan), ThreadSanitizer (TSan), and UBSan into C++/Kotlin build/test flows; generate comparative safety reports.
*	Reporting: Provide detailed, CI-generated sanitizer logs.
*	Developer Experience: Streamline build/test workflow and ensure CI/CD best practices.

⸻

## 3. Scope
	•	Monorepo Structure: Unified repo with clear modular boundaries.
	•	Custom Rules: Separate .bzl libraries for each language, supporting advanced build/test features (e.g., sanitizer toggling).
	•	Migration CLI: Python-based tool to analyze, translate, and verify Gradle-to-Bazel conversion.
	•	Sanitizer Workflows: Bazel test targets for running builds/tests with sanitizers, and collecting logs.
	•	CI/CD Integration: Example GitHub Actions (or similar) pipelines.

⸻

## 4. Architecture & Components

4.1 Repo Layout
```
/
├── WORKSPACE
├── BUILD.bazel (root-level aggregates)
├── language/
│   ├── kotlin/
│   │   ├── src/   (app and lib code)
│   │   ├── test/
│   │   ├── BUILD.bazel
│   ├── cpp/
│   │   ├── src/
│   │   ├── test/
│   │   ├── BUILD.bazel
│   ├── python/
│   │   ├── src/
│   │   ├── test/
│   │   ├── BUILD.bazel
├── starlark/
│   ├── kotlin_rules.bzl
│   ├── cpp_rules.bzl
│   ├── python_rules.bzl
├── migration-tool/
│   ├── cli.py
│   ├── gradle_parser.py
│   ├── BUILD.bazel
├── reporting/
│   ├── sanitizer_report.py
│   ├── logs/
│   ├── BUILD.bazel
├── ci/
│   ├── github_actions.yml
│   ├── scripts/
└── README.md
```
4.2 Key Components
*	Custom Rules:
*	Starlark .bzl files for each language.
*	Macros for multi-language targets (e.g., Kotlin-JVM calling C++ via JNI).
*	Built-in options for sanitizer enablement (using Bazel config_setting/select()).
*	Migration Tool:
*	Python CLI: analyze, translate, verify commands.
*	Parses build.gradle using regex or via Kotlin DSL AST.
*	Emits Bazel BUILD stubs with preserved source/test/resources structure.
*	Sanitizer Reporting:
*	Bazel test targets with sanitizer flags.
*	Scripts to parse sanitizer output into structured JSON.
*	Summarized HTML/Markdown/JSON reports for CI artifacts.
*	CI Pipeline:
*	Runs builds and tests for all modules with and without sanitizers.
*	Publishes migration reports and sanitizer logs as build artifacts.

⸻

## 5. Technology Stack
	•	Build: Bazel 6.x, Starlark
	•	Migration Tool: Python 3.10+
	•	Gradle Source: Gradle 8.x (JVM), C++17 toolchain, Kotlin JVM plugin
	•	Languages: Kotlin (JVM), C++17, Python 3.x

⸻

## 6. Data Flow & Interactions

Developer Workflow:
	1.	Developer runs bazel build //language/...
	2.	Bazel uses custom Starlark rules to compile sources, run tests, and package artifacts.
	3.	Migration tool runs as migration-tool/cli.py analyze --source=gradle_project/
	•	Inspects build.gradle files, outputs analysis JSON.
	4.	cli.py translate emits corresponding BUILD.bazel files, preserving tests/resources.
	5.	CI runs sanitizer-enabled test targets (bazel test --config=asan //... etc.)
	6.	Logs from sanitizer runs collected in reporting/logs/.
	7.	sanitizer_report.py aggregates and summarizes logs for developer review.

⸻

## 7. Non-Functional Requirements
	•	Build Cache: ≥80% cache hits across repeated builds.
	•	Migration Tool Runtime: <30s for medium-sized Gradle projects (≤50 modules).
	•	Error Reporting: All tools must emit actionable, clear error messages for missing/unsupported dependencies.

⸻

## 8. Security Considerations
	•	Review Starlark rules for unsafe use of ctx or external commands.
	•	Ensure Bazel runs (and sanitizer jobs) are sandboxed to prevent leaking or overwriting files.

⸻

9. Deployment
	•	Monorepo: Hosted publicly on GitHub.
	•	Migration Tool:
	•	Published to PyPI, versioned and documented.
	•	Dockerfile for containerized usage.
	•	Documentation:
	•	Full README with quickstart, setup, migration, and troubleshooting sections.

⸻

## 10. Testing Strategy
	•	Starlark Unit Tests: Use Bazel’s built-in Skylark test facilities to test custom rule logic and macros.
	•	E2E Migration Tests:
	•	Example Gradle projects in migration-tool/tests/data/
	•	Verify translated BUILD.bazel files produce correct artifacts and run all tests.
	•	Sanitizer Regression:
	•	Deliberate “bad code” test cases to ensure sanitizer failures are detected and reported.
	•	CI:
	•	All above tests are mandatory on PRs.

⸻

## 11. Timeline & Milestones

Week	Deliverable
1	Monorepo skeleton, basic Bazel/Gradle configs, sample code in each language, initial Starlark rules
2	Migration tool MVP, can parse simple build.gradle and emit correct BUILD.bazel
3	Sanitizer integration in Bazel rules, sample reporting scripts, E2E tests
4	CI/CD setup, all tests automated, documentation and usage guide finalized


⸻

## 12. Maintenance & Monitoring
	•	Version Tracking:
	•	Track Bazel/Gradle/Kotlin/C++/Python updates quarterly.
	•	Community/Issues:
	•	Use GitHub issue templates for rule/migration bugs or improvement requests.
	•	Continuous Improvement:
	•	Automated CI badge for build/cache/sanitizer status.
	•	Quarterly review for compatibility and best practice updates.

⸻

## 13. (Optional) Architecture Diagram

Include a simple diagram if presenting:

Developer
    |
    v
Migration Tool (Python CLI) <-- parses build.gradle
    |
    v
Bazel BUILD files (generated)
    |
    v
Bazel Build (Custom Starlark Rules)
    |
    +-> [Kotlin] ------> Compile/Test/Package
    +-> [C++] ---------> Compile/Test/Sanitize
    +-> [Python] ------> Test/Lint
    |
    v
Sanitizer Reporting & CI Pipeline
    |
    v
Artifacts & Logs (HTML, JSON, badge)


⸻

## 14. Risks & Mitigations
	•	Complex Gradle projects (multi-level, plugins):
Mitigation: Document unsupported cases in tool output; provide extensibility hooks.
	•	Sanitizer false negatives:
Mitigation: Maintain sample “bad code” regression cases, and compare sanitizer runs to reference logs.

⸻

## 15. References
	•	Bazel Starlark Reference
	•	Gradle Build Model Reference
	•	Sanitizers in Bazel
	•	PyPI Packaging Guide
	•	Bazel CI Example

⸻

## 16. Appendix
	•	Example migration tool invocation and expected output.
	•	Sample custom Starlark rule for C++ target with sanitizer config.
	•	Typical sanitizer report artifact structure.
