# Modular JVM Build

**Problem:** Demonstrate a modular JVM architecture and industry-standard build setup that supports cross-platform deployment (e.g. Apple Silicon and x86 JVM).

**Solution:** A Gradle-based multi-module Kotlin/Java project with clear boundaries (`core`, `api`, `app`), a Spring Boot service, and documentation of module layout and cross-platform choices.

**What this demonstrates:** Modular solution for JVM ecosystems, Gradle and Kotlin expertise, Spring Boot, and technical documentation—aligned with modular architecture for Apple Silicon and JVM, build systems, and API design.

---

## Module layout

| Module | Purpose |
|--------|---------|
| `core` | Shared domain types and utilities; no framework dependencies. |
| `api` | REST API surface (controllers, DTOs); depends on `core`. |
| `app` | Spring Boot application; wires `api` and runs the server. |

See [Architecture](docs/architecture.md) for design and cross-platform notes.

---

## How to run

**Build and test:**

If you have the Gradle wrapper (run `gradle wrapper` once if needed):

```bash
./gradlew build
```

Or with Gradle installed:

```bash
gradle build
```

**Run the application:**

```bash
./gradlew :app:bootRun
# or
gradle :app:bootRun
```

API: http://localhost:8080. Example: `GET http://localhost:8080/api/health`.

**Run tests only:**

```bash
./gradlew test
```

---

## Cross-platform note

The project is JVM-only; Gradle and the JVM run on Apple Silicon (aarch64) and x86. For future multi-arch native or Docker images, add architecture-specific tasks or use Gradle toolchains; see `docs/architecture.md`.

---

## License

MIT.
