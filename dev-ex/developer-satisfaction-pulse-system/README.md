# Developer Satisfaction Pulse System

## Overview
A lightweight DevEx survey system that captures weekly pulse responses and quarterly SPACE-aligned assessments, stores them in Postgres, and provides trend-friendly team metrics.

## Key Feature
Dual-cadence analytics that merges short weekly sentiment checks with quarterly deep surveys to produce one consistent rolling team health signal.

## Architecture
- FastAPI service for response ingestion and metrics access
- Domain models for survey responses and score normalization
- Scoring service for rolling NPS and friction trend indicators
- Postgres persistence layer (to be implemented behind repository interfaces)
- Dashboard/BI integration through metrics endpoints

## Use Cases
- Track team sentiment drift over release cycles
- Compare friction between platform-consuming teams
- Validate whether DevEx improvements improve satisfaction over time
- Provide quarterly executive reporting with weekly leading indicators

## Usage
```bash
make install
make run
```

Run tests:
```bash
make test
```

## Control Flow
1. Team members submit pulse or quarterly responses.
2. API validates and stores responses in Postgres.
3. Scoring pipeline computes rolling NPS and normalized friction metrics.
4. Aggregate metrics are exposed for dashboards and reporting APIs.
5. Teams review trend changes and correlate with platform initiatives.

## Project Structure
- `src/app/main.py`: FastAPI application bootstrap
- `src/app/api/`: HTTP routes and handlers
- `src/app/domain/`: Survey domain models
- `src/app/services/`: Scoring and analytics logic
- `tests/`: API and scoring test coverage
- `scripts/`: Local developer scripts
- `docs/`: Architecture and operational documentation
