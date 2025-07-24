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

- ✅ **Add interactive prompts** - Implemented comprehensive interactive setup with step-by-step configuration
- ✅ **Better template validation** - Added robust validation system with error reporting and template syntax checking
- ✅ **Plugin system** - Created extensible plugin architecture with built-in Docker and documentation plugins

## Recent Improvements (v2.0)

### Interactive Setup

- Full interactive mode with `--interactive` flag
- Smart prompts for missing arguments
- Configuration validation with detailed error reporting
- Preview and confirmation before generation

### Enhanced Validation

- Template syntax validation using Jinja2
- Project configuration validation
- Comprehensive error and warning reporting
- Manifest file validation for plugin templates

### Plugin Architecture

- Base `ProjectPlugin` class for extensibility
- Built-in plugins for Docker and MkDocs documentation
- Plugin manager for loading external plugins
- Post-generation hooks for advanced customization

### Additional Features

- Enhanced file templates with better defaults
- Comprehensive .gitignore generation
- Full license text generation (MIT, Apache-2.0, etc.)
- Better CLI UX with progress indicators and next steps
- Template validation command (`projgen validate`)
- Plugin management commands (`projgen plugins`, `projgen install-plugin`)

### Usage Examples

```bash
# Interactive setup
projgen init --interactive

# Quick setup with validation
projgen init my-project --language python --validate-only

# With plugins
projgen init my-app --language node --interactive  # Prompts for Docker, docs, etc.

# List available plugins
projgen plugins

# Validate templates
projgen validate
```
