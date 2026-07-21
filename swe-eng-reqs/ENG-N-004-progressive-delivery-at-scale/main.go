package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ENG-N-004: Progressive delivery at scale.
// Multi-environment progressive delivery plans with automated promotion
// criteria (D-03). Promotion to the next environment happens only when
// server-computed criteria pass. Single-canary weight math belongs to
// ENG-E-024, not here.

type DemoState struct {
	mu       sync.Mutex
	Runs     int    `json:"runs"`
	LastDemo string `json:"last_demo"`
}

var state = &DemoState{}

var store = NewScaleStore()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-004 listening on %s (requirement ENG-N-004)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux() *http.ServeMux {
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
			"requirement_id": "ENG-N-004",
			"service":        "eng-n-004",
			"title":          "Progressive delivery at scale",
			"owns":           "multi-environment PD + automated promotion criteria",
			"not":            "single-canary weight internals (ENG-E-024)",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/plans", handlePlans)
	mux.HandleFunc("/v1/plans/", handlePlanRoutes)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-004_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-004_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-004_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handlePlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Envs     []string `json:"envs"`
		Criteria Criteria `json:"criteria"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	plan, err := store.CreatePlan(body.Envs, body.Criteria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, plan)
}

func handlePlanRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/plans/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	if len(parts) == 2 && parts[1] == "evaluate" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var m Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		plan, err := store.Evaluate(id, m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, plan)
		return
	}
	if len(parts) == 2 && parts[1] == "promote" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		plan, err := store.Promote(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, plan)
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		plan, err := store.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, plan)
		return
	}
	http.NotFound(w, r)
}

// handleDemo runs a live multi-env plan: a failing evaluation blocks promotion,
// then a passing evaluation auto-promotes to the next environment.
func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewScaleStore()
	plan, err := demo.CreatePlan([]string{"dev", "staging", "prod"}, Criteria{MaxErrorRate: 0.01, MinSuccess: 0.99})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Failing criteria must block promotion.
	if _, err := demo.Evaluate(plan.ID, Metrics{ErrorRate: 0.2, SuccessRate: 0.8}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, blockErr := demo.Promote(plan.ID)
	promoteBlocked := blockErr != nil

	// Passing criteria auto-promotes to the next environment.
	ev, err := demo.Evaluate(plan.ID, Metrics{ErrorRate: 0.001, SuccessRate: 0.999})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	promoted, err := demo.Promote(plan.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-004",
		"service":        "eng-n-004",
		"run":            state.Runs,
		"message":        "demo completed for Progressive delivery at scale",
		"acceptance": []string{
			"healthz returns 200",
			"failing criteria blocks promotion",
			"passing criteria auto-promotes to next environment",
		},
		"proof": map[string]any{
			"environments":                 len(plan.Envs),
			"criteria_passed":              ev.CriteriaPassed,
			"auto_promoted":                promoted.CurrentEnv == "staging",
			"promote_blocked_when_failing": promoteBlocked,
			"current_env":                  promoted.CurrentEnv,
			"env_status":                   promoted.EnvStatus,
		},
	}
	state.mu.Unlock()
	writeJSON(w, result)
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
