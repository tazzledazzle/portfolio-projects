- Reads and writes `build.gradle.kts` (Kotlin DSL) and `settings.gradle.kts`
- Configures dependencies: `implementation`, `api`, `testImplementation`, `runtimeOnly`
- Understands task graph; runs tasks with `./gradlew :module:taskName`
- Writes simple custom tasks; uses `doFirst`/`doLast`

#### Level 3 — Proficient
- Multi-project builds: `allprojects`, `subprojects`, `project()` references, composite builds
- Custom plugins: `Plugin<Project>`, registers tasks with `project.tasks.register`
- Convention plugins via `buildSrc` or included builds; shared build logic
- Dependency management: version catalogs (`libs.versions.toml`), BOM imports, capability conflicts
- Build caching: `@CacheableTask`, task input/output declarations for cache correctness
- Configuration avoidance: `register` vs. `create`, `configureEach` vs. `all`
- Incremental task API: `@InputFiles` with `@PathSensitive`, `@OutputDirectory`

#### Level 4 — Expert
- Gradle internals: configuration phase vs. execution phase; project model, dependency resolution engine
- Custom dependency resolution rules: `resolutionStrategy`, component metadata rules, artifact transforms
- Remote build cache (Gradle Enterprise / Develocity): setup, cache node configuration, hit rate analysis
- Bazel migration experience: identifying Bazel-equivalent concepts, `WORKSPACE`/`MODULE.bazel` setup
- Worker API for parallel task execution within a single Gradle invocation
- Toolchain API for multi-JDK support; cross-compilation configurations
- Performance profiling: `--profile`, `--scan`, identifying slow configurations and task execution
