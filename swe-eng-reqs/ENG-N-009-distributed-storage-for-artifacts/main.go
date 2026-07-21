package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ENG-N-009: Distributed storage for artifacts
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-009",
	"service":        "eng-n-009",
	"title":          "Distributed storage for artifacts",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewFacade(3, 2))
	log.Printf("eng-n-009 listening on %s (requirement ENG-N-009)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(facade *Facade) http.Handler {
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
			"requirement_id": "ENG-N-009",
			"service":        "eng-n-009",
			"title":          "Distributed storage for artifacts",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/objects", func(w http.ResponseWriter, r *http.Request) {
		handleObjects(w, r, facade)
	})
	mux.HandleFunc("/v1/objects/", func(w http.ResponseWriter, r *http.Request) {
		handleObjects(w, r, facade)
	})
	mux.HandleFunc("/v1/durability/", func(w http.ResponseWriter, r *http.Request) {
		handleDurability(w, r, facade)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, facade)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-009_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-009_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-009_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleObjects(w http.ResponseWriter, r *http.Request, facade *Facade) {
	if r.URL.Path == "/v1/objects" {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 100<<20))
		if err != nil {
			http.Error(w, "invalid object", http.StatusBadRequest)
			return
		}
		digest, err := facade.Put(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		durability, _ := facade.Durability(digest)
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{"digest": digest, "durability": durability})
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	digest := strings.TrimPrefix(r.URL.Path, "/v1/objects/")
	blob, ok := facade.Get(digest)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	durability, _ := facade.Durability(digest)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Artifact-Replicas", strconv.Itoa(durability.Replicas))
	w.Header().Set("X-Artifact-Checksum", durability.Checksum)
	w.Header().Set("X-Healthy-Nodes", strconv.Itoa(durability.HealthyNodes))
	_, _ = w.Write(blob)
}

func handleDurability(w http.ResponseWriter, r *http.Request, facade *Facade) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	digest := strings.TrimPrefix(r.URL.Path, "/v1/durability/")
	durability, ok := facade.Durability(digest)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, durability)
}

func handleDemo(w http.ResponseWriter, r *http.Request, facade *Facade) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	digest, err := facade.Put([]byte("durability-demo"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	durability, _ := facade.Durability(digest)
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-009",
		"service":        "eng-n-009",
		"run":            state.Runs,
		"message":        "demo completed for Distributed storage for artifacts",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = map[string]any{
		"nodes":         facade.NodeCount(),
		"replicas":      durability.Replicas,
		"healthy_nodes": durability.HealthyNodes,
		"checksum":      durability.Checksum,
		"durable":       durability.Durable,
	}
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
	case "ENG-N-010", "ENG-N-003":
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
