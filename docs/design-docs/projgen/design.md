# Project Generator

## Context

CLI tool for scaffolding new projects with templates.

## Problem & Goals

- Reduce project setup time
- Standardize project structure
- Provide template system
- Good CLI UX

## Constraints & Risks

- Template maintenance overhead
- Cross-platform compatibility

## Architecture & Alternatives

- Python with Typer for CLI
- Jinja2 for templating
- YAML configuration

## Trade-offs

- Flexibility vs simplicity
- Template variety vs maintenance

## Results & Metrics

- 90% reduction in setup time
- Consistent project structure
- Easy template addition

## What I'd change next time

- Add interactive prompts
- Better template validation
- Plugin system