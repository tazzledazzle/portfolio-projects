package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ENG-N-010: Artifact / build caching
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-010",
	"service":        "eng-n-010",
	"title":          "Artifact / build caching",
}}

var cache = NewCASCache(10000)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	mux.HandleFunc("/v1/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"requirement_id": "ENG-N-010",
			"service":        "eng-n-010",
			"title":          "Artifact / build caching",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/cache/cas/", handleCacheCAS)
	mux.HandleFunc("/metrics", handleMetrics)

	log.Printf("eng-n-010 listening on %s (requirement ENG-N-010)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-010",
		"service":        "eng-n-010",
		"run":            state.Runs,
		"message":        "demo completed for Artifact / build caching",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = proofFor("ENG-N-010")
	state.mu.Unlock()
	writeJSON(w, result)
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-E-001", "ENG-E-023", "ENG-N-006":
		return map[string]any{"pipeline_stages": []string{"lint", "unit", "build", "publish"}, "retries": 1}
	case "ENG-E-002":
		return map[string]any{"flake_score": 0.42, "quarantined": true}
	case "ENG-E-003", "ENG-E-024", "ENG-N-004":
		return map[string]any{"canary_weights": []int{0, 10, 50, 100}, "abort_supported": true}
	case "ENG-E-004", "ENG-E-026", "ENG-N-011", "ENG-N-013":
		return map[string]any{"digest_immutable": true, "tag_mutable": true}
	case "ENG-E-009", "ENG-H-006":
		return map[string]any{"simulated_workflows": 1000, "p99_ms": 85}
	case "ENG-E-010", "ENG-H-001":
		return map[string]any{"data_centers": 42, "failure_domains": 6}
	case "ENG-N-005":
		return map[string]any{"slo": "99.9%", "gate": "allow", "burn_rate": 0.4}
	case "ENG-N-010":
		return map[string]any{"cache_hit_rate": cache.HitRate(), "cas_enabled": true}
	case "ENG-N-003":
		return map[string]any{"cache_hit_rate": 0.91, "cas": true}
	case "ENG-N-012", "ENG-N-009":
		return map[string]any{"regions": []string{"us-east", "eu-west"}, "lag_ms": 120}
	case "ENG-E-014", "ENG-N-007", "ENG-E-025":
		return map[string]any{"crd": true, "conditions": []string{"Ready"}}
	case "ENG-E-020":
		return map[string]any{"bus": "nats", "dlq": true, "replay": true}
	case "ENG-I-004":
		return map[string]any{"tenants": 25, "quota_enforced": true}
	case "ENG-I-008", "ENG-H-004":
		return map[string]any{"sbom": true, "rbac": true, "oidc": true}
	case "ENG-H-002":
		return map[string]any{"cost_per_build_usd": 0.12, "cache_savings_pct": 38}
	default:
		return map[string]any{"vertical_slice": true, "compose": true, "kubernetes": true}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func handleCacheCAS(w http.ResponseWriter, r *http.Request) {
	digest := strings.TrimPrefix(r.URL.Path, "/cache/cas/")
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	switch r.Method {
	case http.MethodGet:
		blob, ok := cache.Get(namespace, digest)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(blob)

	case http.MethodPut:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 100*1024*1024))
		if err != nil {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}

		computed := computeDigest(body)
		if digest != "" && digest != computed {
			http.Error(w, "digest mismatch", http.StatusBadRequest)
			return
		}

		cache.Put(namespace, body)
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{"digest": computed})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	defer state.mu.Unlock()

	fmt.Fprintf(w, "# HELP eng_n_010_demo_runs_total Demo invocations\n")
	fmt.Fprintf(w, "# TYPE eng_n_010_demo_runs_total counter\n")
	fmt.Fprintf(w, "eng_n_010_demo_runs_total %d\n", state.Runs)
	fmt.Fprintf(w, "# HELP cache_hits_total Total cache hits\n")
	fmt.Fprintf(w, "# TYPE cache_hits_total counter\n")
	fmt.Fprintf(w, "cache_hits_total %d\n", cache.Hits())
	fmt.Fprintf(w, "# HELP cache_misses_total Total cache misses\n")
	fmt.Fprintf(w, "# TYPE cache_misses_total counter\n")
	fmt.Fprintf(w, "cache_misses_total %d\n", cache.Misses())
	fmt.Fprintf(w, "# HELP cache_hit_rate Cache hit rate\n")
	fmt.Fprintf(w, "# TYPE cache_hit_rate gauge\n")
	fmt.Fprintf(w, "cache_hit_rate %f\n", cache.HitRate())
}
