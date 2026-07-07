# Self-Service Pipeline Template Engine

## Overview
Self-Service Pipeline Template Engine generates CI pipelines from service template manifests so teams can bootstrap standardized delivery flows quickly.

## Key Feature
Its unique capability is convention-aware pipeline generation from a lightweight service manifest.

## Architecture
- `src/main.py`: Manifest parsing and workflow rendering.
- `tests/test_smoke.py`: Verifies output contains required stages.
- `Makefile`: Run/test targets.

## Use Cases
- Onboard new services with consistent CI/CD defaults.
- Enforce organization-wide delivery standards.
- Reduce pipeline authoring overhead.

## Usage
```bash
make run
make test
```

## Control Flow
1. Read service template manifest.
2. Resolve stages (lint/test/build/deploy).
3. Render workflow YAML string.
4. Write generated pipeline output.

## Project Structure
```text
self-service-pipeline-template-engine/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
