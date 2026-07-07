# Remote Dev Environment Orchestrator

## Overview
A workflow-driven service that provisions ephemeral cloud development environments, hydrates them with repository state and seed data, enforces TTL teardown, and tracks per-developer cost.

## Architecture
Temporal workflows coordinate provisioning, repository bootstrap, data seeding, usage metering, and teardown activities. Adapter modules integrate with Coder/Gitpod APIs while domain models and workflow config enforce lifecycle policy.

## Use Cases
- Spin up per-developer remote environments for onboarding and short-lived feature work.
- Automatically tear down stale environments with TTL controls.
- Track cost consumption by user/team to guide platform budgeting.

## Usage
1. Configure workflow and provider settings in `config/workflow.yaml`.
2. Wire infrastructure dependencies from `infra/terraform`.
3. Run worker runtime via `scripts/run-worker.sh`.
4. Trigger workflow executions for environment requests.
5. Observe provisioning and teardown lifecycle events in workflow history.

## Control Flow
1. Workflow receives environment provisioning request.
2. Provisioning activity creates a remote workspace via provider adapter.
3. Repository and seed-data activities hydrate the workspace.
4. Cost tracker records lifecycle usage metrics.
5. TTL scheduler triggers teardown activity at expiry.
6. Final state is persisted and exposed for audit/reporting.

## Project Structure
```text
.
├── config/
├── docs/
│   ├── architecture/
│   └── operations/
├── infra/
│   └── terraform/
├── scripts/
├── src/
│   ├── activities/
│   │   ├── cost/
│   │   ├── provisioning/
│   │   ├── repository/
│   │   ├── seeding/
│   │   └── teardown/
│   ├── adapters/
│   │   ├── coder/
│   │   └── gitpod/
│   ├── domain/
│   └── workflows/
└── tests/
    ├── integration/
    ├── unit/
    └── workflow/
```
