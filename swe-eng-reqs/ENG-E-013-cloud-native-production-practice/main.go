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

// ENG-E-013: Cloud-native production practice
// Owns Dockerfile + probes + HPA-ready + compose parity (not IDP/queue/operator).

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	pkg := NewPackaging(".", "k8s/deploy.yaml")
	mux := newMux(pkg)
	log.Printf("eng-e-013 listening on %s (requirement ENG-E-013)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(pkg *Packaging) http.Handler {
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
			"requirement_id": "ENG-E-013",
			"service":        "eng-e-013",
			"title":          "Cloud-native production practice",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/packaging", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, pkg.Packaging())
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, pkg)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "# HELP eng_e_013_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_e_013_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_e_013_demo_runs_total %d\n", pkg.DemoCount())
	})
	return mux
}

func handleDemo(w http.ResponseWriter, r *http.Request, pkg *Packaging) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	run := pkg.IncrementDemo()
	facts := pkg.Packaging()
	probes := pkg.Probes()
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-013",
		"service":        "eng-e-013",
		"run":            run,
		"message":        "demo completed for Cloud-native production practice",
		"acceptance": []string{
			"healthz/readyz probes present",
			"HPA-ready manifests",
			"Dockerfile + compose parity",
		},
		"proof": map[string]any{
			"probes":         probes,
			"hpa_ready":      pkg.HPAReady(),
			"compose_parity": facts["compose_parity"] == true,
			"dockerfile":     facts["dockerfile"] == true,
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
