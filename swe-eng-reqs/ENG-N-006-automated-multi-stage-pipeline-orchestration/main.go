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

// ENG-N-006: Automated multi-stage pipeline orchestration.
// Advances release/environment stages ONLY when a gate returns allow (D-09).
// This is NOT a CI DAG: no lint/unit/build/publish semantics, no HTTP call to
// ENG-N-005 — the GateEvaluator is an in-folder stub (D-05, D-11).

type DemoState struct {
	mu       sync.Mutex
	Runs     int    `json:"runs"`
	LastDemo string `json:"last_demo"`
}

var state = &DemoState{}

var store = NewOrchestrator(NewStubGate("allow"))

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-006 listening on %s (requirement ENG-N-006)", *addr)
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
			"requirement_id": "ENG-N-006",
			"service":        "eng-n-006",
			"title":          "Automated multi-stage pipeline orchestration",
			"orchestrates":   "release/environment stages",
			"not":            "CI DAG (lint/unit/build/publish)",
			"gate":           "embedded GateEvaluator stub (no HTTP to ENG-N-005)",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/orchestrations", handleOrchestrations)
	mux.HandleFunc("/v1/orchestrations/", handleOrchestrationRoutes)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-006_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-006_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-006_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleOrchestrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Stages []string `json:"stages"`
		SLOID  string   `json:"slo_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	orc, err := store.Create(body.Stages, body.SLOID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, orc)
}

func handleOrchestrationRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/orchestrations/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 2 && parts[1] == "tick" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		orc, decision, err := store.Tick(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		orc.LastDecision = decision.Decision
		writeJSON(w, orc)
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		orc, err := store.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, orc)
		return
	}
	http.NotFound(w, r)
}

// handleDemo runs a live orchestration: deny first (blocked_on_deny), then flip
// the gate to allow and advance to the terminal stage (stages_advanced). Every
// Tick consults the gate (gate_required). Proof vocabulary is release-stage
// only — no CI job names (D-09).
func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gate := NewStubGate("deny")
	orch := NewOrchestrator(gate)
	orc, err := orch.Create(nil, "checkout-availability")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	afterDeny, denyDecision, err := orch.Tick(orc.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	blockedOnDeny := afterDeny.Blocked && afterDeny.StagesAdvanced == 0 && denyDecision.Decision == "deny"

	gate.SetDecision("allow")
	if _, _, err := orch.Tick(orc.ID); err != nil { // dev -> staging
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	afterAllow, _, err := orch.Tick(orc.ID) // staging -> prod
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-006",
		"service":        "eng-n-006",
		"run":            state.Runs,
		"message":        "demo completed for Automated multi-stage pipeline orchestration",
		"acceptance": []string{
			"healthz returns 200",
			"deny blocks stage advance",
			"allow advances release/env stage after gate check",
		},
		"proof": map[string]any{
			"stages_advanced": afterAllow.StagesAdvanced,
			"blocked_on_deny": blockedOnDeny,
			"gate_required":   gate.Calls() >= 3,
			"stages":          afterAllow.Stages,
			"current_stage":   afterAllow.CurrentStage,
			"terminal":        afterAllow.CurrentStage == "prod",
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
