# DESIGN DOCUMENT: Developer-Productivity Tooling Project

## Overview

Create a CLI tool (Python or Rust) that scaffolds new projects (Bazel or Gradle) with opinionated defaults, 
or build an IntelliJ plugin that detects and highlights stale dependencies.

## Goals and Objectives

• Reduce onboarding time for new team members by automating project setup
• Improve code health by proactively identifying unused or outdated dependencies

----

## Scope

• CLI mode: new-project --language kotlin --build bazel
• IDE plugin mode: scans build.gradle(.kts) or WORKSPACE for dependency usage
• Generate reports or quick-fix suggestions

## Architecture and Components

• CLI core module: argument parsing, template rendering
• Template library: commands, examples, README, basic CI files
• IntelliJ plugin module: Kotlin-based plugin using IntelliJ SDK

## Technology Stack

• Python 3.10 + Click or Rust + Clap for CLI
• JetBrains IntelliJ SDK for plugin
• Jinja2 for templating

## Data Flow and Interactions

User runs tool new-project … → tool copies and customizes templates

In IDE, plugin indexes build files, analyzes AST for references

Plugin underlines stale deps and offers quick-fix

## Non-Functional Requirements

• CLI execution < 2 s
• Plugin analysis latency < 500 ms per file

## Security Considerations

• Validate template inputs to avoid path traversal
• Restrict plugin file system access

## Deployment Strategy

• Publish CLI on PyPI or crates.io
• Distribute IntelliJ plugin via JetBrains Marketplace

## Testing Strategy
• Unit tests for CLI logic
• Plugin integration tests with IntelliJ Platform Test Framework

## Timeline and Milestones

Week 1: CLI MVP with one language/build combo
Week 2: Template library expansion
Week 3: Plugin MVP detecting one class of stale deps
Week 4: Quick-fix actions and docs

## Maintenance & Monitoring

• Collect user feedback via GitHub issues
• Quarterly updates for new build-system versions
