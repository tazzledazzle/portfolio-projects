# Devcontainer Feature Library

## Overview
A reusable library of composable devcontainer features that lets teams bootstrap a complete local platform stack from `.devcontainer/devcontainer.json` with minimal per-repo setup.

## Architecture
The project is organized around independent feature modules (`otel-collector`, `local-kafka`, `local-postgres-seed`, `local-keycloak`) and an index that catalogs feature metadata and compatibility. Example consumer devcontainers demonstrate composition patterns for Java and Node services.

## Use Cases
- Standardize local platform dependencies across microservice repositories.
- Reduce onboarding time by shipping tested devcontainer feature bundles.
- Enable consistent observability and auth dependencies in local environments.

## Usage
1. Select feature modules from `features/`.
2. Reference feature images/options in `.devcontainer/devcontainer.json`.
3. Use `scripts/validate-features.sh` to validate feature metadata and contracts.
4. Start the devcontainer and verify sidecars/services from the included examples.

## Control Flow
1. Developer selects required local capabilities.
2. Devcontainer runtime resolves feature declarations.
3. Each feature `install.sh` runs and provisions tools/services.
4. Feature index and templates keep configuration consistent across projects.

## Project Structure
```text
.
├── .devcontainer/
├── assets/
│   └── templates/
├── docs/
│   ├── architecture/
│   └── features/
├── examples/
│   ├── java-service/
│   │   └── .devcontainer/
│   └── node-service/
│       └── .devcontainer/
├── features/
│   ├── local-kafka/
│   ├── local-keycloak/
│   ├── local-postgres-seed/
│   └── otel-collector/
├── scripts/
├── src/
│   └── feature-index/
└── tests/
    ├── e2e/
    ├── integration/
    └── unit/
```
