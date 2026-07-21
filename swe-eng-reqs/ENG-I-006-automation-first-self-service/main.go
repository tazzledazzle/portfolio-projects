package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ENG-I-006: Automation-first / self-service
// Owns ticket-bottleneck removal + before/after metrics (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-I-006",
		"service":        "eng-i-006",
		"title":          "Automation-first / self-service",
	}}
	ssStore = NewSelfServiceStore(10)
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-i-006 listening on %s (requirement ENG-I-006)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(ssStore)))
}

func newMux(store *SelfServiceStore) http.Handler {
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
			"requirement_id": "ENG-I-006",
			"service":        "eng-i-006",
			"title":          "Automation-first / self-service",
			"owns":           []string{"ticket_removed", "automation_metrics"},
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/requests", func(w http.ResponseWriter, r *http.Request) {
		handleRequests(w, r, store)
	})
	mux.HandleFunc("/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.Metrics())
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, store)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_i_006_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_i_006_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_i_006_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleRequests(w http.ResponseWriter, r *http.Request, store *SelfServiceStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Kind    string `json:"kind"`
		Summary string `json:"summary"`
		// Intentionally ignore client ticket_removed (T-5-02)
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	created, err := store.Submit(req.Kind, req.Summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, created)
}

func handleDemo(w http.ResponseWriter, r *http.Request, store *SelfServiceStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	created, err := store.Submit("env-provision", "demo self-service")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m := store.Metrics()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-006",
		"service":        "eng-i-006",
		"run":            run,
		"message":        "live ticket-removal self-service demo",
		"acceptance": []string{
			"self-service submit removes ticket path",
			"server-computed before/after ticket metrics",
		},
		"proof": map[string]any{
			"ticket_removed":     created.TicketRemoved,
			"self_service":       true,
			"automation_metrics": m.AutomationMetrics,
		},
	})
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
