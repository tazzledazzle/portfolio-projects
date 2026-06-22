# Tooling Adoption Tracker

## Overview
A telemetry-driven service that tracks IDE plugin usage, CLI invocations, and internal portal interactions to measure how quickly teams adopt new developer platform capabilities.

## Key Feature
Per-team adoption funnel analysis (view -> install -> run) that highlights exactly where rollout friction occurs for each tool launch.

## Architecture
- Node.js + TypeScript ingestion service
- Event model for normalized telemetry across IDE, CLI, and portal sources
- Funnel analytics module for stage conversion computation
- Storage abstraction for in-memory and database-backed event stores
- Reporting output for dashboards and launch scorecards

## Use Cases
- Measure adoption velocity after launching a new CLI capability
- Identify teams that viewed documentation but never installed the tool
- Compare adoption performance across tool categories
- Trigger enablement campaigns based on funnel drop-off points

## Usage
```bash
make install
make dev
```

Build for production:
```bash
make build
```

## Control Flow
1. Telemetry events are ingested from opt-in sources.
2. Events are normalized into a shared schema.
3. Pipeline computes funnel stage counts by tool and team.
4. Aggregates are stored and exported to reporting surfaces.
5. Platform teams use funnel gaps to plan onboarding interventions.

## Project Structure
- `src/index.ts`: Application entrypoint
- `src/ingest/`: Telemetry ingestion and normalization
- `src/analytics/`: Funnel and adoption metric computation
- `src/storage/`: Repository abstractions and adapters
- `tests/`: Analytics behavior tests
- `scripts/`: Local run scripts
- `docs/`: Telemetry schema and reporting docs
