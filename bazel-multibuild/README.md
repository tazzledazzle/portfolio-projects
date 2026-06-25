# Bazel Multi-Build

Polyglot Bazel demo with Java and TypeScript modules.

## Quick start

Java (all targets):

```bash
cd java
bazel test //...
```

TypeScript frontend:

```bash
cd frontend
bazel test //...
```

The `flags-parsing-tutorial/` subdirectory contains an additional Bazel example.

Container image targets in `java/BUILD` are commented out until `rules_oci` is wired for your Bazel version.
