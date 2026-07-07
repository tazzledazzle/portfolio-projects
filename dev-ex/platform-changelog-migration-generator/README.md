# Platform Changelog Migration Generator

## Overview
A release automation utility that diffs old and new platform APIs, produces structured changelog output, and generates starter migration scripts for breaking changes in Kotlin and Python codebases.

## Key Feature
Automated migration guidance that links each detected breaking API change directly to language-specific AST rewrite scaffolds.

## Architecture
- Python CLI orchestration layer
- API diff engine for public signature comparison
- Change classifier for breaking/deprecated/additive categorization
- Migration transform module with language-specific generators
- Artifact emission for changelog and migration script outputs

## Use Cases
- Generate release notes from API changes on every library release
- Reduce migration effort for downstream service teams
- Enforce consistency in deprecation communication
- Accelerate platform adoption by lowering upgrade friction

## Usage
```bash
make install
make run
```

Run tests:
```bash
make test
```

Or run directly:
```bash
python3 -m changelog_generator.cli --old api_v1.json --new api_v2.json
```

## Control Flow
1. CI trigger runs generator on old/new API artifacts.
2. Diff engine detects signature-level changes.
3. Classifier groups changes by compatibility impact.
4. Changelog builder emits structured release entries.
5. Migration generators output Kotlin/Python rewrite stubs for breaking changes.

## Project Structure
- `src/changelog_generator/cli.py`: CLI entrypoint
- `src/changelog_generator/diff_engine.py`: API diff logic
- `src/changelog_generator/transforms/`: Kotlin/Python migration generators
- `tests/`: Diff and transform tests
- `scripts/`: Local run scripts
- `docs/`: Migration strategy and release process docs
