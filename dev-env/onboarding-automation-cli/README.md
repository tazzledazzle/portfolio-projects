# Onboarding Automation CLI

## Overview
A Kotlin CLI that automates new engineer setup by cloning required repositories, configuring git/SSH, installing toolchain versions, and validating setup through health checks.

## Architecture
The CLI uses command modules for user flows, installer adapters for dependency managers (`asdf`), and subsystem handlers for repository bootstrap, SSH setup, and environment verification. Shared config/templates drive repeatable onboarding profiles.

## Use Cases
- Bring new hires to a working local setup quickly.
- Eliminate manual setup drift between teams.
- Standardize setup verification before first development tasks.

## Usage
1. Define org defaults in `config/defaults.yaml`.
2. Customize repository template data in `src/main/resources/templates/repos.yaml`.
3. Build and run CLI commands (for example: `setup`).
4. Run health checks to verify repository, tooling, and credential readiness.

## Control Flow
1. CLI command parses onboarding options.
2. Repo bootstrapper clones/configures required repositories.
3. Installer adapter ensures required toolchain versions exist.
4. SSH and git modules configure authentication and remotes.
5. Health suite executes validations and reports pass/fail status.

## Project Structure
```text
.
├── config/
├── docs/
│   ├── architecture/
│   └── runbooks/
├── scripts/
├── src/
│   └── main/
│       ├── kotlin/com/company/onboarding/
│       │   ├── commands/
│       │   ├── git/
│       │   ├── health/
│       │   ├── installers/
│       │   └── ssh/
│       └── resources/templates/
└── tests/
    ├── fixtures/
    ├── integration/
    └── unit/
```
