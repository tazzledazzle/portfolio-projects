// Envoy/Istio proxy-wasm filter
//
// What this filter does:
//   1. Generates a UUID v4 trace ID and injects it as x-trace-id on every inbound request
//      (only if x-trace-id is not already present — respects upstream trace propagation)
//   2. Adds x-request-timestamp with the current Unix time in milliseconds
//
// Why proxy-wasm instead of Lua filters:
//   - Type safety and compile-time correctness (vs. untyped Lua runtime errors)
//   - Better performance: Rust compiles to efficient WASM bytecode
//   - Rich ecosystem: use any Rust crate that compiles to wasm32 (with caveats)
//   - Stronger sandboxing: WASM runs in a memory-safe sandbox inside the Envoy process
//
// Deployment (Istio WasmPlugin — configure in your GitOps manifests):
//
//   apiVersion: extensions.istio.io/v1alpha1
//   kind: WasmPlugin
//   metadata:
//     name: trace-id-filter
//     namespace: orders
//   spec:
//     selector:
//       matchLabels:
//         app: orders-service
//     url: oci://registry.example.com/wasm/envoy-trace-filter:0.1.0
//     phase: AUTHN    # Run before authentication — trace ID is available in auth logs
//     pluginConfig:
//       header_prefix: "x-"   # configurable via plugin_configuration (see on_configure)

use proxy_wasm::traits::*;
use proxy_wasm::types::*;
use uuid::Uuid;

// ─── Plugin root (one instance per Envoy worker thread) ───────────────────────

struct TraceFilterRoot {
    header_prefix: String,
}

impl Default for TraceFilterRoot {
    fn default() -> Self {
        Self {
            header_prefix: "x-".to_string(),
        }
    }
}

impl RootContext for TraceFilterRoot {
    /// Called when Istio delivers the WasmPlugin.pluginConfig to the filter.
    /// Parse JSON config here; fall back to defaults on parse failure.
    fn on_configure(&mut self, _plugin_configuration_size: usize) -> bool {
        if let Some(config_bytes) = self.get_plugin_configuration() {
            if let Ok(config_str) = std::str::from_utf8(&config_bytes) {
                // Minimal JSON parse: look for "header_prefix" key
                // In production, use serde_json (but check binary size impact)
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

// ─── HTTP filter context (one instance per HTTP request) ─────────────────────

struct TraceFilter {
    trace_id_header: String,
    timestamp_header: String,
}

impl Context for TraceFilter {}

impl HttpContext for TraceFilter {
    /// Called when Envoy receives the request headers from the downstream client.
    /// This is where we inject trace headers before forwarding to the upstream service.
    fn on_http_request_headers(&mut self, _num_headers: usize, _end_of_stream: bool) -> Action {
        // Only inject x-trace-id if not already present.
        // This preserves distributed trace context from upstream callers (e.g., the
        // gateway or a service mesh propagating an existing trace span).
        if self.get_http_request_header(&self.trace_id_header).is_none() {
            let trace_id = Uuid::new_v4().to_string();
            self.add_http_request_header(&self.trace_id_header, &trace_id);
        }

        // Always inject the gateway-observed timestamp.
        // Services can use this to compute total request latency including queue time.
        // Note: proxy-wasm does not provide wall-clock time directly; we use
        // get_current_time_nanos() which returns nanoseconds since the Unix epoch.
        let now_nanos = self.get_current_time_nanos();
        let now_ms = now_nanos / 1_000_000;
        self.add_http_request_header(&self.timestamp_header, &now_ms.to_string());

        Action::Continue
    }

    /// Called when Envoy receives the response headers from the upstream service.
    /// Optionally echo the trace ID back to the client for correlation in browser/mobile.
    fn on_http_response_headers(&mut self, _num_headers: usize, _end_of_stream: bool) -> Action {
        // Propagate the trace ID to the response so clients can correlate.
        // If the upstream service already added x-trace-id to the response, this is a no-op.
        if let Some(trace_id) = self.get_http_request_header(&self.trace_id_header) {
            self.set_http_response_header(&self.trace_id_header, Some(&trace_id));
        }
        Action::Continue
    }
}

// ─── WASM entry point ─────────────────────────────────────────────────────────

/// Called by Envoy when loading the WASM module.
/// Sets up the root context factory — Envoy calls this once per worker thread.
#[no_mangle]
pub fn _start() {
    proxy_wasm::set_log_level(LogLevel::Info);
    proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> {
        Box::new(TraceFilterRoot::default())
    });
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

/// Minimal JSON string extraction without pulling in serde_json.
/// Finds the string value of a JSON key in a flat object.
/// Returns None if key is not found or value is not a string.
fn extract_json_string(json: &str, key: &str) -> Option<String> {
    let search = format!("\"{}\"", key);
    let key_pos = json.find(&search)?;
    let after_key = &json[key_pos + search.len()..];
    let colon_pos = after_key.find(':')?;
    let after_colon = after_key[colon_pos + 1..].trim();
    if after_colon.starts_with('"') {
        let end = after_colon[1..].find('"')?;
        Some(after_colon[1..end + 1].to_string())
    } else {
        None
    }
}
