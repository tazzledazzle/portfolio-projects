# Bazel Multi-Build

Polyglot Bazel demo with Java and TypeScript modules.

## Quick start

```bash
cd java
bazel test //:myproject_test
```

The `frontend/` and `flags-parsing-tutorial/` subdirectories contain additional Bazel examples.

Container image targets in `java/BUILD` are commented out until `rules_oci` is wired for your Bazel version.
