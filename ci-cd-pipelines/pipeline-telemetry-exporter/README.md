# Pipeline Telemetry Exporter

## Overview
Pipeline Telemetry Exporter is a CI-focused emitter that converts pipeline lifecycle events into OpenTelemetry-like spans. It helps engineering teams correlate build failures and slow stages with service traces.

## Key Feature
The unique capability is step-level trace emission from CI jobs, including failure reason tags and duration metadata for every step.

## Architecture
- `src/main.py`: Parses event payloads and emits normalized spans.
- `tests/test_smoke.py`: Validates span generation behavior.
- `Makefile`: Local run and test commands.

## Use Cases
- Diagnose slow CI stages across repositories.
- Correlate deployment failures with upstream test failures.
- Build latency dashboards from pipeline traces.

## Usage
```bash
make run
make test
```

## Control Flow
1. Receive a pipeline event payload.
2. Normalize job/step metadata.
3. Build spans with start/end timestamps and attributes.
4. Emit spans to stdout (bootstrap target) for collector integration.

## Project Structure
```text
pipeline-telemetry-exporter/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
