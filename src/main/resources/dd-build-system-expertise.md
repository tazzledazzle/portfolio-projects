# DESIGN DOCUMENT: Build-System Expertise Project
## Overview:
Implement a multi-language monorepo (Kotlin, C++, Python) with custom Bazel rules and a Gradle-to-Bazel migration tool to illustrate end-to-end build-system mastery.

## Goals and Objectives:
• Demonstrate writing Starlark rules that compile, test, and package mixed-language targets
• Automate migration of existing Gradle projects to Bazel, preserving test suites and artifact outputs
• Integrate Address, Thread, and Undefined Behavior Sanitizers and publish comparative reports

## Scope:
• Monorepo layout with three modules
• Custom rule definitions for each language
• Migration tool CLI with “analyze → translate → verify” workflow
• CI integration to run sanitizers

## Architecture and Components:
• Monorepo root with WORKSPACE and BUILD files
• language/ kotlin/, cpp/, python/ subdirectories
• starlark/ folder containing .bzl rule libraries
• migration-tool/ Python CLI using libpython to parse Gradle metadata
• reporting/ scripts to gather sanitizer logs

## Technology Stack:
• Bazel 6.x, Starlark
• Python 3.10 for migration tool
• Gradle 8.x (source)
• C++17 toolchain, Kotlin JVM plugin

## Data Flow and Interactions:

* Developer runs `bazel build //…`

* Bazel invokes custom rules to compile sources

* Migration tool inspects build.gradle files, emits BUILD stubs

* Sanitizer runs under Bazel test targets, logs to JSON

## Non-Functional Requirements:
• Build cache effectiveness ≥ 80% cache hits
• Migration tool runtime < 30s on medium-sized repo
• Clear error reporting for missing dependencies

## Security Considerations:
• Audit custom Starlark code for injection vectors
• Run sanitizers in sandboxed Bazel sandbox

## Deployment Strategy:
• Host monorepo on GitHub
• Publish migration tool as PyPI package
• Provide detailed README with setup steps

## Testing Strategy:
• Unit tests for Starlark rule behavior (using Bazel’s skylark unit test framework)
• End-to-end example repo with Gradle projects to verify migration
• Automated sanitizer regression tests

## Timeline and Milestones:
Week 1: Monorepo skeleton, basic rules
Week 2: Migration tool MVP
Week 3: Sanitizer integration and reports
Week 4: CI setup and documentation

## Maintenance & Monitoring:
• Track Bazel and Gradle version compatibility quarterly
• Issue template for rule improvements and bug reports