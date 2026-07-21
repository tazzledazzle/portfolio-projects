package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// ENG-I-001: Hybrid IC + technical lead (still ships code)
// Shipped service + leadership artifacts in the same folder (not Phase 7 soft-skill kits).

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	store := NewHybridStore("artifacts/leadership")
	mux := newMux(store)
	log.Printf("eng-i-001 listening on %s (requirement ENG-I-001)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(store *HybridStore) http.Handler {
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
			"requirement_id": "ENG-I-001",
			"service":        "eng-i-001",
			"title":          "Hybrid IC + technical lead (still ships code)",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/leadership", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]any{
			"artifacts": store.Leadership(),
			"status":    store.HybridStatus(),
		})
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, store)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "# HELP eng_i_001_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_i_001_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_i_001_demo_runs_total %d\n", store.DemoCount())
	})
	return mux
}

func handleDemo(w http.ResponseWriter, r *http.Request, store *HybridStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	run := store.IncrementDemo()
	status := store.HybridStatus()
	sample := store.Sample()
	artifacts := store.Leadership()

	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-001",
		"service":        "eng-i-001",
		"run":            run,
		"message":        "demo completed for Hybrid IC + technical lead (still ships code)",
		"acceptance": []string{
			"service code ships in this folder",
			"leadership artifacts under artifacts/leadership/",
			"hybrid_ic requires both code and leadership",
		},
		"proof": map[string]any{
			"code_shipped":         sample["code_shipped"] == true,
			"leadership_artifacts": status["leadership_artifacts"] == true,
			"hybrid_ic":            status["hybrid_ic"] == true,
			"artifact_paths":       artifacts,
		},
	}
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
