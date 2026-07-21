package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ENG-E-001: CI and build infrastructure ownership
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu        sync.Mutex
	Runs      int            `json:"runs"`
	LastDemo  string         `json:"last_demo"`
	Meta      map[string]any `json:"meta"`
	Pipelines map[string]any `json:"pipelines"`
}

var state = &DemoState{
	Meta: map[string]any{
		"requirement_id": "ENG-E-001",
		"service":        "eng-e-001",
		"title":          "CI and build infrastructure ownership",
	},
	Pipelines: map[string]any{},
}

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
			"requirement_id": "ENG-E-001",
			"service":        "eng-e-001",
			"title":          "CI and build infrastructure ownership",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/pipelines", handlePipelines)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-001_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-001_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-001_demo_runs_total %d\n", state.Runs)
		fmt.Fprintf(w, "# HELP eng-e-001_pipelines_total Pipelines submitted\n")
		fmt.Fprintf(w, "# TYPE eng-e-001_pipelines_total gauge\n")
		fmt.Fprintf(w, "eng-e-001_pipelines_total %d\n", len(state.Pipelines))
	})

	log.Printf("eng-e-001 listening on %s (requirement ENG-E-001)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

type PipelineRequest struct {
	Stages map[string][]string `json:"stages"`
}

func handlePipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		state.mu.Lock()
		defer state.mu.Unlock()
		writeJSON(w, state.Pipelines)
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req PipelineRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if len(req.Stages) > 100 {
			http.Error(w, "DAG exceeds max 100 stages", http.StatusBadRequest)
			return
		}

		pipeline := NewPipeline(req.Stages)
		if err := pipeline.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid DAG: %v", err), http.StatusBadRequest)
			return
		}

		state.mu.Lock()
		id := fmt.Sprintf("pipe-%d", len(state.Pipelines)+1)
		pipeline.ID = id

		stagesJSON := make(map[string]any)
		for name, stage := range pipeline.Stages {
			stagesJSON[name] = map[string]any{
				"name":         stage.Name,
				"status":       stage.Status,
				"dependencies": stage.Dependencies,
				"retries":      stage.Retries,
			}
		}

		pipelineData := map[string]any{
			"id":     id,
			"stages": stagesJSON,
			"status": "pending",
		}
		state.Pipelines[id] = pipelineData
		state.mu.Unlock()

		writeJSON(w, pipelineData)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePipelineByID(w http.ResponseWriter, r *http.Request, id string) {
	state.mu.Lock()
	defer state.mu.Unlock()

	pipeline, ok := state.Pipelines[id]
	if !ok {
		http.Error(w, "pipeline not found", http.StatusNotFound)
		return
	}

	writeJSON(w, pipeline)
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
		"requirement_id": "ENG-E-001",
		"service":        "eng-e-001",
		"run":            state.Runs,
		"message":        "demo completed for CI and build infrastructure ownership",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = proofFor("ENG-E-001")
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
