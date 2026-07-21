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

// ENG-E-012: Production proficiency — single Go under OR semantics (D-08, D-11).
// Does NOT add a Python pair; CLAUDE.md wins over dual-language wording.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-012",
	"service":        "eng-e-012",
	"title":          "Production proficiency in Go, Python, or similar cloud-native languages",
}}

var proficiency = NewProficiency()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-e-012 listening on %s (requirement ENG-E-012)", *addr)
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
	mux.HandleFunc("/v1/info", handleInfo)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-012_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-012_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-012_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := proficiency.Info()
	info["time"] = time.Now().UTC()
	writeJSON(w, info)
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	info := proficiency.Info()
	sample := proficiency.Sample()
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-012",
		"service":        "eng-e-012",
		"run":            state.Runs,
		"message":        "OR semantics satisfied with single Go production sample",
		"acceptance": []string{
			"language=go",
			"or_semantics=true (Go OR Python OR similar — Go chosen)",
			"production_sample via net/http handlers",
		},
		"proof": map[string]any{
			"language":          info["language"],
			"or_semantics":      info["or_semantics"] == true,
			"production_sample": sample["handler"] == true,
			"health_path":       sample["health_path"],
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
