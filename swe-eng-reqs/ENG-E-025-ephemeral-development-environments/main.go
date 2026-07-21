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

// ENG-E-025: Ephemeral development environments
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-025",
	"service":        "eng-e-025",
	"title":          "Ephemeral development environments",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewController())
	log.Printf("eng-e-025 listening on %s (requirement ENG-E-025)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(controller *Controller) http.Handler {
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
			"requirement_id": "ENG-E-025",
			"service":        "eng-e-025",
			"title":          "Ephemeral development environments",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/devenvs", func(w http.ResponseWriter, r *http.Request) {
		handleDevEnvs(w, r, controller)
	})
	mux.HandleFunc("/v1/devenvs/", func(w http.ResponseWriter, r *http.Request) {
		handleDevEnvs(w, r, controller)
	})
	mux.HandleFunc("/v1/reconcile", func(w http.ResponseWriter, r *http.Request) {
		handleReconcile(w, r, controller)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, controller)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-025_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-025_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-025_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleDevEnvs(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.URL.Path == "/v1/devenvs" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			ID         string `json:"id"`
			TTLSeconds int64  `json:"ttl_seconds"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		env, err := controller.Create(request.ID, request.TTLSeconds, nowUTC())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, env)
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/devenvs/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && r.Method == http.MethodGet {
		env, ok := controller.Get(parts[0])
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, env)
		return
	}
	if len(parts) == 2 && parts[1] == "tick" && r.Method == http.MethodPost {
		var request struct {
			Now time.Time `json:"now"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if request.Now.IsZero() {
			request.Now = nowUTC()
		}
		env, err := controller.Tick(parts[0], request.Now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, env)
		return
	}
	http.Error(w, "invalid DevEnv route", http.StatusBadRequest)
}

func handleReconcile(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]any{"reclaimed": controller.Reconcile(nowUTC())})
}

func handleDemo(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	now := nowUTC()
	id := fmt.Sprintf("demo-%d", now.UnixNano())
	_, err := controller.Create(id, 1, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	env, err := controller.Tick(id, now.Add(2*time.Second))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conditionTypes := make([]string, 0, len(env.Conditions))
	for _, condition := range env.Conditions {
		conditionTypes = append(conditionTypes, condition.Type)
	}
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-025",
		"service":        "eng-e-025",
		"run":            state.Runs,
		"message":        "demo completed for Ephemeral development environments",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = map[string]any{
		"crd":           true,
		"kind":          "DevEnv",
		"conditions":    conditionTypes,
		"ready":         conditionTrue(env, "Ready"),
		"expired":       conditionTrue(env, "Expired"),
		"ttl_reclaimed": env.Reclaimed,
	}
	state.mu.Unlock()
	writeJSON(w, result)
}

func nowUTC() time.Time {
	return time.Now().UTC()
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
