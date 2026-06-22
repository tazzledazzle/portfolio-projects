# Environment Drift Detector

## Overview
A manifest-driven tool that compares a developer machine's installed toolchain versions against a canonical team baseline and generates actionable remediation scripts.

## Architecture
The detector pipeline includes: manifest parsing, canonical policy loading, semantic version comparison, drift classification, and fix-script generation. Reporting modules provide human-readable summaries for CLI and CI execution.

## Use Cases
- Detect drift in `node`, `jvm`, `terraform`, `kubectl`, and other required tooling.
- Generate prescriptive scripts to close version gaps quickly.
- Enforce environment consistency in preflight checks for local and CI workflows.

## Usage
1. Define baseline versions in `config/canonical-toolchain.yaml`.
2. Capture local versions into a manifest input.
3. Run `scripts/run-detector.sh`.
4. Review report output and execute generated fix script safely.

## Control Flow
1. CLI entrypoint receives a manifest path.
2. Manifest module validates and normalizes tool versions.
3. Detector compares local state with canonical policy.
4. Drift results are categorized (missing, outdated, unsupported).
5. Fix-script module renders shell commands and reporting outputs.

## Project Structure
```text
.
├── config/
├── docs/
│   ├── architecture/
│   └── rules/
├── scripts/
├── src/
│   ├── cli/
│   ├── detector/
│   ├── fix_script/
│   ├── manifest/
│   └── reporting/
└── tests/
    ├── fixtures/
    │   ├── manifests/
    │   └── tool-versions/
    ├── integration/
    └── unit/
```
