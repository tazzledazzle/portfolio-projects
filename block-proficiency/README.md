# Engineering Proficiency Standards

> Reference guide for evaluating and demonstrating technical proficiency across the full stack.
> Technologies: Java · Kotlin · AWS · HTTP/JSON/gRPC/Protobuf · OkHttp/Jetty/JUnit/Guice · Hibernate/jOOQ/Aurora/MySQL/DynamoDB/Redis/Vitess · Kafka/Event-Driven/Microservices · DataDog · Buildkite/Gradle

---

## Proficiency Levels

| Level | Label | Description |
|-------|-------|-------------|
| 1 | **Aware** | Can read and understand existing code; needs guidance to contribute |
| 2 | **Practitioner** | Independently produces correct, working code; understands common patterns |
| 3 | **Proficient** | Makes architectural decisions; reviews others' work; handles edge cases |
| 4 | **Expert** | Deep internals knowledge; optimizes at scale; defines standards for the org |

---

## 1. Java

### Level 1 — Aware
- Understands class/interface/inheritance model
- Can read and trace Java code; understands checked vs. unchecked exceptions
- Familiar with `java.util` collections and `java.io` basics

### Level 2 — Practitioner
- Writes clean, idiomatic Java 11+ (var, records, text blocks, sealed classes in Java 17+)
- Uses generics correctly, including bounded wildcards (`? extends`, `? super`)
- Understands and applies `Stream` API, `Optional`, `CompletableFuture`
- Handles concurrency with `synchronized`, `ReentrantLock`, `ExecutorService`, `ThreadPoolExecutor`
- Manages memory: understands GC basics, heap vs. stack, reference types (weak/soft/phantom)
- Writes clean exception hierarchies; uses try-with-resources

### Level 3 — Proficient
- Designs clean layered architecture (domain, application, infrastructure)
- Applies SOLID principles and GoF design patterns appropriately
- Deep understanding of Java Memory Model (JMM): happens-before, volatile, atomic operations
- Tunes JVM: GC selection (G1, ZGC, Shenandoah), heap sizing, GC logging, thread dump analysis
- Writes thread-safe code using `java.util.concurrent` primitives correctly
- Understands classloading, reflection, and annotation processing
- Profiles with JFR, async-profiler, VisualVM

### Level 4 — Expert
- Deep bytecode understanding; can read/modify with ASM or Byte Buddy
- Implements custom class loaders, agents (`java.lang.instrument`)
- Identifies and eliminates false sharing, cache line contention, lock contention
- Benchmarks precisely with JMH; understands JIT warm-up and dead-code elimination pitfalls
- Designs APIs for binary and source compatibility across versions

---

## 2. Kotlin

### Level 1 — Aware
- Understands Kotlin's null safety (`?`, `!!`, `?.`, `?:`)
- Can read data classes, object declarations, companion objects
- Knows Kotlin compiles to JVM and is interoperable with Java

### Level 2 — Practitioner
- Idiomatic use of: extension functions, sealed classes, `when` expressions, destructuring
- Coroutines: launches with `launch`/`async`, understands `suspend`, uses `withContext`
- Uses `Flow` for reactive streams; handles `StateFlow`/`SharedFlow`
- Scope functions: `let`, `run`, `apply`, `also`, `with` — uses each appropriately
- Kotlin collections API: `map`, `filter`, `fold`, `groupBy`, `associate`, etc.
- Writes clean DSLs using lambda receivers and infix functions

### Level 3 — Proficient
- Structured concurrency: `CoroutineScope`, `SupervisorJob`, `CoroutineExceptionHandler`
- Cancellation propagation, cooperative cancellation with `isActive`/`ensureActive`
- Understands coroutine internals: continuation-passing style (CPS), state machines
- Designs type-safe builders and DSLs
- Uses `inline`/`reified` generics correctly for type erasure workarounds
- Applies `@JvmOverloads`, `@JvmStatic`, `@JvmField` for clean Java interop
- Delegates: `by lazy`, `by Delegates.observable`, property delegation contract

### Level 4 — Expert
- Compiler plugin authorship (K2 plugins, IR transforms)
- Deep coroutine dispatcher internals; writes custom dispatchers
- Kotlin Multiplatform (KMP): shared business logic, `expect`/`actual`, sourceset configuration
- Understands desugaring and optimization differences vs. Java equivalents
- Uses KSP/KAPT for annotation processing and code generation

---

## 3. AWS

### Level 1 — Aware
- Understands regions, availability zones, and the shared responsibility model
- Can navigate the AWS console; familiar with IAM users, roles, and policies conceptually
- Knows what EC2, S3, RDS, Lambda, VPC are at a high level

### Level 2 — Practitioner
- **Compute**: deploys and manages EC2 instances, Auto Scaling Groups, launch templates
- **Networking**: configures VPCs, subnets (public/private), security groups, NACLs, route tables
- **IAM**: writes least-privilege IAM policies; understands role assumption, instance profiles, STS
- **Storage**: uses S3 (lifecycle rules, versioning, bucket policies, presigned URLs), EBS, EFS
- **Managed Services**: stands up RDS, ElastiCache, SQS, SNS; configures DLQs
- Uses CloudFormation or CDK to define infrastructure as code
- Understands CloudWatch metrics, alarms, and log groups

### Level 3 — Proficient
- Designs multi-AZ, fault-tolerant architectures with RTO/RPO targets
- Load balancing: ALB/NLB/CLB differences; path-based routing, health checks, SSL termination
- EKS: cluster management, node groups, IRSA (IAM Roles for Service Accounts), add-ons
- Lambda: cold start mitigation, provisioned concurrency, layers, event source mappings
- Cost optimization: Reserved Instances, Savings Plans, Spot fleet strategies, right-sizing
- Implements cross-account access patterns with Organizations and SCPs
- Understands VPC peering, Transit Gateway, PrivateLink, Direct Connect

### Level 4 — Expert
- Architects globally distributed systems (Route 53 latency/failover routing, CloudFront, Global Accelerator)
- Deep networking: BGP, custom DHCP options, flow logs analysis, NAT gateway vs. instance tradeoffs
- Security: GuardDuty, Security Hub, Config rules, Macie, Detective — designs SIEM pipelines on AWS
- Service limits, quotas, and throttling patterns at scale
- Writes CDK constructs and CloudFormation macros; contributes to internal platform IaC libraries

---

## 4. HTTP, JSON, gRPC, Protocol Buffers

### Level 1 — Aware
- Knows HTTP verbs (GET, POST, PUT, DELETE, PATCH), status code families (2xx, 4xx, 5xx)
- Understands JSON syntax and data types
- Aware that gRPC and Protobuf exist for binary/typed RPC

### Level 2 — Practitioner
- **HTTP**: uses headers correctly (Content-Type, Authorization, Cache-Control, ETag/If-None-Match)
- Understands HTTP/1.1 vs. HTTP/2 (multiplexing, header compression, server push)
- Designs RESTful APIs: resource naming, idempotency, pagination (cursor vs. offset), versioning
- **JSON**: schema validation, serialization/deserialization with Jackson or kotlinx.serialization
- **Protobuf**: writes `.proto` files with messages, enums, oneof, repeated fields; understands field number rules
- **gRPC**: implements unary, server-streaming, client-streaming, and bidirectional RPC; uses metadata

### Level 3 — Proficient
- REST API design: HATEOAS, content negotiation, conditional requests, ETags at scale
- HTTP/2 and HTTP/3 (QUIC): understands head-of-line blocking elimination, connection coalescing
- gRPC: interceptors (auth, logging, tracing), deadlines/cancellations, error status codes, retry policies
- Protobuf: backward/forward compatibility rules (reserved fields, optional vs. required), `Any`, `Timestamp`, `Duration` well-known types
- gRPC-Web and transcoding (gRPC-Gateway); service reflection and grpcurl
- Implements rate limiting, circuit breaking, and retries at the HTTP layer

### Level 4 — Expert
- Designs API gateways and service meshes integrating gRPC + HTTP seamlessly
- Protobuf schema registry; managing breaking change policies across dozens of services
- Deep HTTP internals: TCP slow start, CWND, TLS handshake optimization, OCSP stapling
- Implements custom gRPC load balancing policies and name resolvers
- Benchmarks and profiles protobuf serialization vs. alternatives (FlatBuffers, Cap'n Proto, Avro)

---

## 5. OkHttp, Jetty, JUnit, Guice

### OkHttp

#### Level 2 — Practitioner
- Configures `OkHttpClient` with timeouts, connection pool, interceptors
- Makes synchronous and asynchronous calls; handles `Response` and `ResponseBody` correctly (closes body)
- Adds custom interceptors for logging, auth header injection, retry logic

#### Level 3 — Proficient
- Implements `Authenticator` for 401 token refresh; `CertificatePinner` for SSL pinning
- Tunes connection pool (`maxIdleConnections`, `keepAliveDuration`) for throughput vs. resource use
- Understands OkHttp's event lifecycle (`EventListener`); instruments for metrics
- Mocks with `MockWebServer` in tests; tests timeout and error scenarios

#### Level 4 — Expert
- Custom `Dns` resolver, `SocketFactory`, `SSLSocketFactory` for advanced networking
- Deep call/connection/stream lifecycle; diagnoses leaked connections
- Performance tuning: HTTP/2 framing, HPACK header compression, push promises

---

### Jetty

#### Level 2 — Practitioner
- Embeds Jetty (`Server`, `ServerConnector`, `ServletContextHandler`) programmatically
- Configures HTTPS with `SslContextFactory`; sets thread pool size
- Writes and registers Servlets and Filters; understands `HttpServletRequest`/`HttpServletResponse`

#### Level 3 — Proficient
- Configures `QueuedThreadPool` and `ScheduledExecutorScheduler` for production load
- Uses `AsyncContext` for non-blocking request handling; understands Servlet 3.1 async model
- WebSocket support with `WebSocketServlet`; session management
- `RequestLog` configuration for access logging; `StatisticsHandler` for metrics
- Configures Jetty with `jetty.xml`/`web.xml` for production deployments

#### Level 4 — Expert
- Custom `ConnectionFactory` and protocol negotiation (ALPN for HTTP/2)
- Jetty's internal `ByteBufferPool`, `Executor`, and I/O framework
- Diagnoses connection leaks, thread starvation, and head-of-line blocking in production

---

### JUnit

#### Level 2 — Practitioner
- Writes clean JUnit 5 tests: `@Test`, `@BeforeEach`, `@AfterEach`, `@BeforeAll`, `@AfterAll`
- Uses `@ParameterizedTest` with `@ValueSource`, `@CsvSource`, `@MethodSource`
- Assertions with `assertAll`, `assertThrows`, `assertTimeout`
- Mocking with Mockito: `@Mock`, `@InjectMocks`, `verify`, `when/thenReturn`, `ArgumentCaptor`

#### Level 3 — Proficient
- JUnit 5 extensions: `@ExtendWith`, implements `BeforeEachCallback`, `AfterEachCallback`, `TestInstancePostProcessor`
- Dynamic tests with `@TestFactory`; nested tests with `@Nested`
- Test lifecycle and parallelism: `@Execution(CONCURRENT)`, thread-safety considerations
- Integration testing with Testcontainers (real databases, Kafka, Redis in Docker)
- Property-based testing with jqwik

#### Level 4 — Expert
- Custom JUnit Platform `TestEngine`; implements `LauncherDiscoveryListener`
- Mutation testing (PIT); designing test architecture for large codebases
- Benchmark integration; test impact analysis and selective test execution at scale

---

### Guice

#### Level 2 — Practitioner
- Writes `AbstractModule`; uses `@Inject` (constructor, field, method injection)
- Understands bindings: `bind(Interface.class).to(Impl.class)`, `toInstance`, `toProvider`
- Scopes: `@Singleton`, `@RequestScoped`; understands lifecycle implications
- Uses `@Named` and `@Qualifier` annotations for multi-binding disambiguation

#### Level 3 — Proficient
- Multibindings: `Multibinder`, `MapBinder` for plugin-style architectures
- `PrivateModule` for encapsulation; `install()` composition patterns
- `TypeLiteral` for generic type injection; `Key` API
- `@Provides` methods; `ProviderMethods` module
- Interceptors with `AOP Alliance` (`MethodInterceptor`, `bindInterceptor`)
- Guice Stage: `DEVELOPMENT` vs. `PRODUCTION`; eager singleton loading

#### Level 4 — Expert
- `InjectionListener`, `TypeListener`, `BindingTargetVisitor` for framework-level hooks
- Bootstrapping Guice in complex multi-module applications with child injectors
- Performance analysis of injector creation; minimizing reflection overhead
- Migrating from Guice to or from Spring; hybrid configurations

---

## 6. Hibernate, jOOQ, Aurora, MySQL, DynamoDB, Redis, Vitess

### Hibernate (ORM)

#### Level 2 — Practitioner
- Entities with `@Entity`, `@Table`, `@Id`, `@GeneratedValue`, `@Column`
- Relationships: `@OneToMany`, `@ManyToOne`, `@ManyToMany`, `@OneToOne`; cascade types
- JPQL and Criteria API for queries
- Understands `EntityManager` lifecycle: persist, merge, remove, detach, refresh

#### Level 3 — Proficient
- N+1 query detection and resolution (`@BatchSize`, `JOIN FETCH`, EntityGraph)
- First-level (session) vs. second-level cache (`@Cache`, Ehcache/Redis integration)
- Optimistic locking with `@Version`; pessimistic locking strategies
- Inheritance strategies: `SINGLE_TABLE`, `JOINED`, `TABLE_PER_CLASS` — tradeoffs
- Custom types with `@Type`, user types; `@Formula` for computed columns
- Hibernate statistics, slow query logging, SQL output formatting for debugging

#### Level 4 — Expert
- Custom `ConnectionProvider`, `Dialect`, `PhysicalNamingStrategy`
- Schema migration integration (Flyway/Liquibase); multi-tenancy patterns
- Envers for auditing; custom event listeners (`PreInsertEventListener`, etc.)
- Deep understanding of flush modes, dirty checking, and session factory internals

---

### jOOQ

#### Level 2 — Practitioner
- Code generation from database schema; uses generated `DSL` classes
- Type-safe SELECT, INSERT, UPDATE, DELETE using DSL API
- Understands `Record`, `Result`, `Field`, `Table` abstractions
- Transactions with `DSLContext.transaction()`

#### Level 3 — Proficient
- Complex queries: CTEs (`with()`), window functions, lateral joins, `MULTISET`
- Dynamic query construction; conditional `WHERE` clause building
- Batch operations; `RETURNING` clause; upsert with `ON CONFLICT`/`ON DUPLICATE KEY`
- Custom converters and bindings for domain types (e.g., JSON columns, enums)
- `ResultQuery` streaming for large result sets; `fetchLazy()`

#### Level 4 — Expert
- Custom `ExecuteListener` for logging, metrics, query rewriting
- Multi-dialect support; managing generated code across environments
- Performance analysis of generated SQL vs. hand-crafted SQL

---

### Aurora / MySQL

#### Level 2 — Practitioner
- CRUD, JOINs, GROUP BY, ORDER BY, LIMIT/OFFSET
- Understands indexes: B-tree, composite index column ordering, covering indexes
- Basic `EXPLAIN` plan reading; identifies full table scans

#### Level 3 — Proficient
- Aurora-specific: reader/writer endpoint routing, Aurora Global Database, fast failover
- Index strategy: selectivity, index merges, invisible indexes, prefix indexes on text
- Query optimization: pushing predicates, avoiding function-on-column, index skips scans
- Transaction isolation levels (READ COMMITTED, REPEATABLE READ, SERIALIZABLE) and locking
- MySQL replication: binlog formats (ROW vs. STATEMENT vs. MIXED), GTID, lag monitoring
- Partitioning: RANGE, LIST, HASH — when each helps and its tradeoffs
- Connection pooling: ProxySQL, RDS Proxy — configuration and max_connections tuning

#### Level 4 — Expert
- Aurora storage architecture: shared distributed storage, redo log offloading
- InnoDB internals: buffer pool management, redo/undo logs, MVCC, deadlock detection
- Point-in-time recovery, clone instances, backtrack
- Designing schemas for high write throughput (hot row avoidance, UUID vs. auto-increment)

---

### DynamoDB

#### Level 2 — Practitioner
- Understands partition key, sort key, and their impact on data distribution
- CRUD via AWS SDK; `GetItem`, `PutItem`, `Query`, `Scan`
- `FilterExpression` vs. `KeyConditionExpression` — knows the difference
- DynamoDB Streams basics; TTL attribute configuration

#### Level 3 — Proficient
- Access pattern-first schema design: single-table design, GSI/LSI strategy
- Capacity modes: provisioned vs. on-demand; auto-scaling configuration
- Transactions: `TransactWriteItems`, `TransactGetItems` — 25-item limit, idempotency tokens
- Conditional writes: `ConditionExpression`, optimistic locking with version attributes
- Pagination: `LastEvaluatedKey`; parallel `Scan` for batch processing
- Hot partition identification; key sharding strategies (write sharding)

#### Level 4 — Expert
- Global Tables: multi-region active-active, conflict resolution, replication lag
- DynamoDB Accelerator (DAX): caching semantics, eventual consistency implications
- Advanced single-table patterns: adjacency lists, hierarchical data, inverted indexes
- Cost modeling at scale; reserved capacity planning

---

### Redis

#### Level 2 — Practitioner
- Core data structures: String, List, Set, Sorted Set, Hash
- Key expiration (`EXPIRE`, `EXPIREAT`, `TTL`); eviction policies
- Pub/Sub basics; simple caching patterns (cache-aside)
- Connects via Jedis or Lettuce from JVM

#### Level 3 — Proficient
- Lua scripting for atomic multi-key operations
- Pipelining and batching to reduce round-trips
- Redis Streams for event log / message queue use cases
- HyperLogLog for cardinality estimation; Bloom filter via RedisBloom
- Sentinel for HA; Cluster for horizontal scaling — understands slot distribution
- Redlock distributed lock algorithm (and its limitations)
- Persistence: RDB snapshots vs. AOF; hybrid persistence; durability tradeoffs

#### Level 4 — Expert
- Redis Cluster internals: hashslots, gossip protocol, failover election
- Memory optimization: `OBJECT ENCODING`, ziplist vs. listpack vs. hashtable transitions
- Keyspace notifications for event-driven patterns
- Performance: `DEBUG SLEEP`, `LATENCY HISTORY`, `SLOWLOG` — production diagnosis
- Designing eviction policies for different cache semantics (LRU vs. LFU)

---

### Vitess

#### Level 2 — Aware / Practitioner
- Understands Vitess as a MySQL sharding/proxy layer for horizontal scaling
- Knows VSchema, keyspaces, shards, tablets (primary, replica, rdonly)
- Basic VTGate query routing; understands why certain SQL constructs are unsupported

#### Level 3 — Proficient
- VSchema design: vindexes (lookup, hash, numeric), sharding keys, scatter avoidance
- MoveTables and Reshard workflows for online schema/resharding migrations
- Understands VStream for CDC (Change Data Capture) integration
- Query planning: `EXPLAIN FORMAT=vitess`; identifying scatter queries and optimizing
- Connection pooling in VTTablet; OLAP vs. OLTP workload routing

#### Level 4 — Expert
- Vitess operator (Kubernetes) deployment and upgrade lifecycle
- Designing vindexes for complex access patterns; custom vindex implementation
- VReplication internals; managing long-running migrations
- Diagnosing cross-shard transaction limitations; 2PC tradeoffs

---

## 7. Kafka, Event-Driven Architecture, Microservices

### Kafka

#### Level 2 — Practitioner
- Core model: topics, partitions, offsets, consumer groups, brokers
- Produces and consumes with Java/Kotlin client (`KafkaProducer`, `KafkaConsumer`)
- Understands `auto.offset.reset`, `enable.auto.commit`, and manual offset committing
- Basic topic configuration: `replication.factor`, `min.insync.replicas`, retention

#### Level 3 — Proficient
- Producer guarantees: `acks=all`, idempotent producer, transactional API
- Consumer group rebalancing: cooperative vs. eager rebalancing; `partition.assignment.strategy`
- Kafka Streams: `KStream`, `KTable`, stateful operations, windowing (tumbling, hopping, session)
- Schema registry with Avro/Protobuf: compatibility modes (BACKWARD, FORWARD, FULL)
- Performance tuning: `batch.size`, `linger.ms`, `compression.type`, `max.poll.records`
- Lag monitoring; consumer group lag alerting; Burrow, Cruise Control
- Exactly-once semantics (EOS): producer transactions + consumer `isolation.level`

#### Level 4 — Expert
- Broker internals: log segments, index files, page cache reliance, zero-copy (`sendfile`)
- Partition leadership rebalancing; controlled shutdown, unclean leader election tradeoffs
- Kafka Streams state store internals (RocksDB); standby replicas for failover
- Tiered storage; MirrorMaker 2 for geo-replication
- Capacity planning: throughput/partition math, broker sizing, disk throughput requirements

---

### Event-Driven Architecture

#### Level 2 — Practitioner
- Understands events vs. commands vs. queries (CQRS); pub/sub vs. point-to-point
- Designs events with stable schemas; includes event ID, timestamp, aggregate ID
- Implements idempotent consumers; uses deduplication keys

#### Level 3 — Proficient
- Event sourcing: event store, aggregate rebuilding, snapshots, projections
- Saga pattern: choreography (event-chain) vs. orchestration (workflow coordinator) — tradeoffs
- Outbox pattern for reliable event publishing from transactional writes
- Dead letter queue strategy; poison pill detection and handling
- Schema evolution: additive changes, field deprecation, consumer-driven contract testing

#### Level 4 — Expert
- Temporal coupling vs. spatial decoupling tradeoffs at organizational scale
- Event storming facilitation; bounded context and domain event mapping
- Long-running process managers; compensating transactions at scale
- Designs event mesh architectures; event catalog and discoverability

---

### Microservices

#### Level 2 — Practitioner
- Decomposes monoliths along domain boundaries; understands bounded contexts
- Implements REST or gRPC service-to-service communication
- Uses environment variables and config maps for configuration; 12-factor app principles
- Containerizes services with Docker; writes production-ready `Dockerfile`s

#### Level 3 — Proficient
- Service mesh: Istio/Linkerd — traffic management, mTLS, observability, circuit breaking
- API gateway patterns: rate limiting, auth offloading, request routing
- Health checks: liveness vs. readiness vs. startup probes; graceful shutdown
- Distributed tracing: OpenTelemetry instrumentation; trace context propagation (`W3C TraceContext`)
- Resilience patterns: circuit breaker (Resilience4j), bulkhead, retry with exponential backoff + jitter
- Service discovery: DNS-based, client-side (Ribbon), server-side (ELB/ALB)
- Database-per-service pattern; data consistency across service boundaries

#### Level 4 — Expert
- Migration strategies: strangler fig, anti-corruption layer, parallel run
- Designs for organizational scale: team API contracts, consumer-driven contract testing (Pact)
- Chaos engineering: Chaos Monkey, fault injection, game day planning
- Sizing and capacity modeling for microservice ecosystems

---

## 8. DataDog

### Level 1 — Aware
- Knows DataDog is an observability platform for metrics, logs, and traces
- Can navigate dashboards; read graphs and alert states

### Level 2 — Practitioner
- **Metrics**: instruments code with `DogStatsD` or `dd-trace`; understands gauge, counter, histogram, distribution metric types
- **APM**: enables distributed tracing; reads flame graphs and service maps
- **Logs**: configures log collection; writes log queries in DataDog Log Explorer
- **Monitors**: creates threshold, anomaly, and forecast monitors; sets up notification routing
- **Dashboards**: builds dashboards with time series, query value, heatmap, log stream widgets

### Level 3 — Proficient
- Custom metrics with tags for high-cardinality slicing; understands metric cardinality limits and cost impact
- **Trace Search**: uses span-level filtering, retention filters, and trace sampling strategies
- **SLOs**: defines and tracks error budget SLOs (metric-based and monitor-based)
- **RUM** (Real User Monitoring): session replay, core web vitals, error tracking
- **Infrastructure**: host maps, process monitoring, network performance monitoring (NPM)
- Writes DataDog synthetics (API and browser tests) for uptime and regression detection
- Log pipelines: parsers (Grok), remappers, enrichment lookups, index routing

### Level 4 — Expert
- Designs org-wide tagging taxonomy (`env`, `service`, `version`, `team`) for cross-team consistency
- Custom `dd-agent` integrations and checks for internal services
- Cost governance: metrics without limits, log retention policies, APM retention filter strategy
- Incident management workflows; runbook automation via DataDog Workflows
- Correlates metrics/logs/traces for root cause analysis in complex distributed systems

---

## 9. Buildkite & Gradle

### Buildkite

#### Level 2 — Practitioner
- Writes `pipeline.yml` with `command`, `wait`, `block`, and `trigger` steps
- Configures agents; understands agent queues and tags for routing
- Uses environment variables, `BUILDKITE_*` built-ins
- Uploads artifacts and uses `buildkite-agent artifact download`
- Implements basic parallelism with `parallelism` key

#### Level 3 — Proficient
- Dynamic pipelines: `buildkite-agent pipeline upload` from scripts; generates steps programmatically
- Matrix builds: fan-out across OS/language/environment combinations
- Plugin authorship: `plugin.yml` schema, Docker, test-collector, and custom plugins
- Secrets management: Elastic CI Stack SSM integration, environment hooks
- `pre-command`, `post-command`, `environment` hooks for cross-cutting concerns
- Test Analytics: uploads JUnit XML; tracks test suite health and flaky test detection
- Cluster and queue architecture for multi-environment isolation

#### Level 4 — Expert
- Self-hosted agent fleet operations: autoscaling (EC2 ASG, k8s, ECS), spot instance strategies
- Implements build observability: trace CI spans with OTel, dashboard build performance
- Designs CI architecture for monorepos: change detection, selective step execution, affected-module targeting
- Cache optimization: remote cache backends, cache invalidation strategies, artifact deduplication

---

### Gradle

#### Level 2 — Practitioner
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

---

## Cross-Cutting Standards

### Observability (all services)
- Every service emits structured logs (JSON), metrics (RED: Rate, Error, Duration), and distributed traces
- Services define and meet SLOs; alert on error budget burn rate, not just thresholds
- Runbooks are linked from every alert

### Testing (all languages)
- Unit tests: isolated, fast (<100ms each), no I/O
- Integration tests: use Testcontainers or local equivalents; real DB/cache/broker
- Contract tests: Pact or equivalent for service-to-service APIs
- Target: 80%+ line coverage on business logic; 100% coverage on critical paths

### Security (all services)
- Secrets never in code or logs; use SSM Parameter Store, Secrets Manager, or Vault
- mTLS or at minimum TLS 1.2+ for all service-to-service communication
- Least-privilege IAM; no wildcard actions in production policies
- Dependency vulnerability scanning in CI (Dependabot, Trivy, Snyk)

### Code Review Standards
- PRs scoped to single concern; description explains *why*, not just *what*
- No merging with failing CI or unresolved review threads
- Performance-sensitive code includes benchmark evidence

---

*Last updated: June 2026*