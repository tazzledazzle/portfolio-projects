# Architecture: Modular JVM Build

## Module layout

```
modular-jvm-build/
├── core/       # Domain types, utilities; no Spring/framework
├── api/        # REST controllers, DTOs; depends on core
└── app/        # Spring Boot entrypoint; depends on api (and thus core)
```

- **core:** Keeps business primitives and pure logic framework-agnostic so they can be reused in other modules (e.g. workers, CLI) or tested without Spring.
- **api:** HTTP layer only; delegates to core or app-specific services. Keeps controllers thin.
- **app:** Application assembly: `@SpringBootApplication`, configuration, and wiring. Only this module has `spring-boot` and runs the server.

## Dependency direction

- `app` → `api` → `core`. No reverse dependencies.
- No cycles; core has zero dependency on api or app.

## Cross-platform and deployment

- **JVM:** Build and run on any JVM (Java 17+). Gradle and the JVM support both Apple Silicon (aarch64) and x86, so the same build runs on both.
- **Docker:** Build a single image from the `app` module (e.g. `./gradlew :app:bootJar` then `COPY app/build/libs/app-*.jar`). For multi-arch images, use Docker buildx or CI matrix for `linux/amd64` and `linux/arm64`.
- **Future native / multi-arch:** To add GraalVM native or explicit arch splits, introduce separate Gradle tasks or profiles and document in this file.

## Build and test pipeline

- `./gradlew build` runs compile, test, and packaging for all modules.
- CI (GitHub Actions) runs `./gradlew check` (or `build`) on push/PR.
- No integration tests in this minimal demo; add a separate `integration` or `app`-scoped test source set if needed.
