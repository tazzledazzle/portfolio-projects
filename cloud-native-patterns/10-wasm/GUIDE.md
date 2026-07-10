# WebAssembly in Cloud-Native: Extending Envoy with Proxy-WASM Filters

WebAssembly (WASM) started as a browser technology — a compact binary format that JavaScript engines could execute at near-native speed. But the same properties that make it useful in browsers — portable, sandboxed, fast — make it valuable on the server side. In cloud-native systems, the most practical use of WASM today is extending Envoy proxy with custom filter logic: injecting headers, enforcing custom auth schemes, manipulating requests and responses before they reach your service.

This guide covers the Proxy-WASM filter in this codebase — a Rust-to-WASM implementation that injects distributed tracing headers into every request passing through an Istio/Envoy sidecar or waypoint.

---

## WebAssembly Beyond the Browser

The core WASM spec defines a portable binary format and an execution model. What's missing is any concept of host environment — no filesystem, no network, no system calls. That's intentional: the sandbox is the feature.

**WASI** (WebAssembly System Interface) adds a standardized system call layer, enabling WASM modules to run on server-side runtimes like `wasmtime`, `wazero`, and `WasmEdge`. This is how WASM microservices work — a WASM module compiled from Rust or Go runs in a wasmtime host, with access to a controlled set of system capabilities.

In cloud-native networking, a different interface standard applies: **Proxy-WASM**. This is the ABI (Application Binary Interface) that Envoy exposes to WASM filter plugins. A Proxy-WASM filter runs inside Envoy's process, with access to the current request's headers, body, and metadata — but sandboxed from the host OS. It can read and modify request and response headers, call external services via the Envoy HTTP filter API, and record metrics. It cannot open network connections directly or access the filesystem.

---

## Why Proxy-WASM Instead of Lua?

Envoy has supported Lua scripting for years. Proxy-WASM is the successor:

- **Type safety**: Rust's type system catches errors at compile time. Lua errors surface at runtime in production.
- **Performance**: Rust compiles to native-speed WASM. Lua is interpreted; the same logic runs faster as WASM.
- **Ecosystem**: Any Rust crate that compiles to `wasm32` is available — JSON parsers, UUID generators, cryptographic primitives. Lua's library ecosystem is narrow.
- **Sandbox isolation**: WASM runs in a separate memory space within the Envoy process. A bug in the filter cannot corrupt Envoy's own memory.
- **Distribution**: WASM binaries are pulled from OCI registries via `WasmPlugin` resources, versioned and deployed like container images.

The tradeoff: Lua filters are simpler to write and iterate on. For simple header manipulation, Lua might be faster to ship. For anything with business logic, type safety, or external library dependencies, Proxy-WASM is the right tool.

---

## The Filter: Tracing Header Injection

The filter in this codebase does two things:

1. Injects `x-trace-id` on every inbound request — a UUID v4 — if the header isn't already present
2. Injects `x-request-timestamp` with the current Unix time in milliseconds
3. Echoes `x-trace-id` back on the response so clients can correlate

This is exactly the kind of cross-cutting concern that belongs in the proxy layer, not in application code. Every service benefits from having a trace ID without each service implementing UUID generation, header injection, and response propagation independently.

---

## The Proxy-WASM Model: Roots and Contexts

```rust
struct TraceFilterRoot {
    header_prefix: String,
}

impl RootContext for TraceFilterRoot {
    fn on_configure(&mut self, _plugin_configuration_size: usize) -> bool {
        if let Some(config_bytes) = self.get_plugin_configuration() {
            if let Ok(config_str) = std::str::from_utf8(&config_bytes) {
                if let Some(prefix) = extract_json_string(config_str, "header_prefix") {
                    self.header_prefix = prefix;
                }
            }
        }
        true
    }

    fn create_http_context(&self, _context_id: u32) -> Option<Box<dyn HttpContext>> {
        Some(Box::new(TraceFilter {
            trace_id_header: format!("{}trace-id", self.header_prefix),
            timestamp_header: format!("{}request-timestamp", self.header_prefix),
        }))
    }

    fn get_type(&self) -> Option<ContextType> {
        Some(ContextType::HttpContext)
    }
}
```

Proxy-WASM has two types of context:

**RootContext**: One instance per Envoy worker thread. Handles plugin lifecycle — initialization, configuration, background tasks, shared state. `on_configure` receives the `pluginConfig` from the `WasmPlugin` Kubernetes resource and parses it.

**HttpContext**: One instance per HTTP request. Created by `create_http_context`. Handles request and response header/body processing for a single request/response cycle.

The root creates HTTP contexts — this is the factory pattern. The header prefix is configured once in the root and passed to each HTTP context on creation.

---

## Request Header Processing

```rust
struct TraceFilter {
    trace_id_header: String,
    timestamp_header: String,
}

impl HttpContext for TraceFilter {
    fn on_http_request_headers(&mut self, _num_headers: usize, _end_of_stream: bool) -> Action {
        // Only inject x-trace-id if not already present.
        // Preserves trace context from upstream callers.
        if self.get_http_request_header(&self.trace_id_header).is_none() {
            let trace_id = Uuid::new_v4().to_string();
            self.add_http_request_header(&self.trace_id_header, &trace_id);
        }

        let now_nanos = self.get_current_time_nanos();
        let now_ms = now_nanos / 1_000_000;
        self.add_http_request_header(&self.timestamp_header, &now_ms.to_string());

        Action::Continue
    }
```

`on_http_request_headers` is called when Envoy receives the request headers from the downstream client. The filter runs before Envoy forwards the request to the upstream service.

The conditional injection (`if ... is_none()`) is important: if the request already carries `x-trace-id` (set by the API gateway or an upstream proxy), this filter preserves it. Overwriting existing trace IDs would break distributed tracing — each hop would generate a new ID, making it impossible to correlate the full request path.

`Action::Continue` tells Envoy to proceed with request forwarding. `Action::Pause` would hold the request — used when waiting for an async result like an external auth check.

`get_current_time_nanos()` is a Proxy-WASM host function that returns nanoseconds since the Unix epoch. Wall-clock time is not directly available to WASM modules — it comes from the host via ABI functions.

---

## Response Header Processing

```rust
fn on_http_response_headers(&mut self, _num_headers: usize, _end_of_stream: bool) -> Action {
    if let Some(trace_id) = self.get_http_request_header(&self.trace_id_header) {
        self.set_http_response_header(&self.trace_id_header, Some(&trace_id));
    }
    Action::Continue
}
```

Reading `x-trace-id` from the *request* headers and writing it to the *response* headers propagates the trace ID back to the client. This lets browsers, mobile apps, and API clients correlate a request they sent with the trace ID they'd use to look up the distributed trace in their observability platform.

---

## The Rust Build Configuration

```toml
[package]
name = "envoy-trace-filter"
version = "0.1.0"

[[lib]]
crate-type = ["cdylib"]   # C-compatible dynamic library — required for WASM

[dependencies]
proxy-wasm = "0.2"
uuid = { version = "1", features = ["v4"] }

[profile.release]
opt-level = "s"    # Optimize for size — WASM binary size matters
lto = true
codegen-units = 1
strip = true
```

`crate-type = ["cdylib"]` compiles to a C-compatible dynamic library — the format Proxy-WASM requires.

The build target: `wasm32-wasip1` (WASI Preview 1). Install it with:
```
rustup target add wasm32-wasip1
cargo build --target wasm32-wasip1 --release
```

The output is `target/wasm32-wasip1/release/envoy_trace_filter.wasm`.

The release profile optimizes for binary size (`opt-level = "s"`) rather than speed. WASM binary size matters because Envoy loads the filter binary into memory for each worker thread, and it's distributed over the network from an OCI registry. LTO and single codegen unit produce smaller binaries at the cost of longer compile times.

---

## Deploying via WasmPlugin

The filter is deployed using Istio's `WasmPlugin` resource, referenced in the source comments:

```yaml
apiVersion: extensions.istio.io/v1alpha1
kind: WasmPlugin
metadata:
  name: trace-id-filter
  namespace: orders
spec:
  selector:
    matchLabels:
      app: orders-service
  url: oci://registry.example.com/wasm/envoy-trace-filter:0.1.0
  phase: AUTHN
  pluginConfig:
    header_prefix: "x-"
```

`url` references an OCI image containing the `.wasm` binary. Istio pulls it, distributes it to the relevant Envoy proxies, and loads it.

`phase: AUTHN` places the filter in the authentication phase of Envoy's filter chain — it runs before authorization filters but after TLS. The trace ID is available in authorization decision logs.

`pluginConfig` is delivered to `on_configure` as JSON. The `header_prefix` key is what the filter reads to configure `x-trace-id` vs. any other prefix.

`selector.matchLabels` means only the `orders-service` pods get this filter — not every service in the namespace. Filters can be scoped to the namespace (no selector), to a service (label selector), or to the mesh (via `MeshConfig`).

---

## What Proxy-WASM Can and Cannot Do

**Can do:**
- Read and modify request/response headers, body, trailers
- Send local responses (reject requests, return synthetic responses)
- Call external HTTP services via Envoy's async dispatch API
- Record metrics (increment counters, set gauges)
- Access shared data across filter contexts within a worker

**Cannot do:**
- Open TCP/UDP connections directly — all network access goes through Envoy's APIs
- Access the filesystem
- Make blocking system calls
- Share state across worker threads without explicit shared data APIs

These restrictions are the sandbox. A filter that hits an out-of-bounds memory access aborts the current request and logs the error — it does not crash Envoy or affect other requests.

---

## Performance Characteristics

WASM execution is fast but not free. Rule of thumb: a minimal Proxy-WASM filter adds ~1–5 microseconds per request, depending on what it does. UUID generation and header manipulation are cheap. Complex JSON parsing of the request body is more expensive.

The Proxy-WASM ABI has overhead at each host function call (`get_http_request_header`, `add_http_request_header`). Minimize host calls in hot paths. This filter makes four host calls per request — well within acceptable range for a latency-sensitive proxy.

Compare to Lua: a simple Lua filter runs at similar latency. A filter using external Lua libraries or doing heavy string processing is slower. Rust WASM has an advantage when logic is computationally intensive.

---

## Key Takeaways

- Proxy-WASM extends Envoy/Istio with custom filter logic compiled to WASM — type-safe, sandboxed, and distributed via OCI registries
- WASM in the proxy layer is the right place for cross-cutting request manipulation: trace injection, custom auth, header normalization
- `RootContext` handles plugin lifecycle (one per worker thread); `HttpContext` handles individual requests (one per request)
- `Action::Continue` forwards the request; `Action::Pause` holds it for async processing (external auth calls)
- Always check if a header already exists before injecting — overwriting trace IDs from upstream proxies breaks distributed tracing
- Compile with `crate-type = ["cdylib"]` and target `wasm32-wasip1`; use `opt-level = "s"` to minimize binary size
- Proxy-WASM cannot directly open connections or access the filesystem — all host interaction goes through the Envoy ABI
- Deploy via `WasmPlugin` resources, scoped by label selector; `phase: AUTHN` places the filter before authorization
- WASM filters add ~1–5 microseconds per request for simple header manipulation — acceptable for most latency budgets
- Proxy-WASM beats Lua when you need type safety, external Rust crates, or complex logic; Lua wins for quick one-off header transforms
