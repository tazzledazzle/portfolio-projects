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

// ENG-E-024: Progressive rollout systems
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-024",
	"service":        "eng-e-024",
	"title":          "Progressive rollout systems",
}}

var canaries = NewCanaryStore()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-e-024 listening on %s (requirement ENG-E-024)", *addr)
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
			"requirement_id": "ENG-E-024",
			"service":        "eng-e-024",
			"title":          "Progressive rollout systems",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/canaries", handleCanaries)
	mux.HandleFunc("/v1/canaries/", handleCanaryRoutes)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-024_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-024_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-024_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleCanaries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Service string `json:"service"`
		Steps   []int  `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	c, err := canaries.Start(body.Service, body.Steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, c)
}

func handleCanaryRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/canaries/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		c, err := canaries.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, c)
		return
	}
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var (
		c   *Canary
		err error
	)
	switch parts[1] {
	case "step":
		c, err = canaries.Step(id)
	case "abort":
		c, err = canaries.Abort(id)
	case "promote":
		c, err = canaries.Promote(id)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, c)
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demoStore := NewCanaryStore()
	c, err := demoStore.Start("demo-svc", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := demoStore.Step(c.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	aborted, err := demoStore.Abort(c.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c2, err := demoStore.Start("demo-svc-promote", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	promoted, err := demoStore.Promote(c2.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-024",
		"service":        "eng-e-024",
		"run":            state.Runs,
		"message":        "demo completed for Progressive rollout systems",
		"acceptance": []string{
			"healthz returns 200",
			"canary steps 0→10→50→100",
			"abort and promote are terminal",
		},
		"proof": map[string]any{
			"canary_weights":  defaultWeights,
			"abort_supported": true,
			"aborted":         aborted.Status == "aborted" && aborted.Weight == 0,
			"promoted":        promoted.Status == "promoted" && promoted.Weight == 100,
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
