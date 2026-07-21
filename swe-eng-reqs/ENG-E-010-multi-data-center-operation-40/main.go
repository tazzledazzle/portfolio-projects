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

// ENG-E-010: Multi-data-center operation (40+)
// Owns ≥40 DC topology + failure domains + fan-out (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-E-010",
		"service":        "eng-e-010",
		"title":          "Multi-data-center operation (40+)",
	}}
	store = NewMultiDC()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-e-010 listening on %s (requirement ENG-E-010)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(store)))
}

func newMux(mdc *MultiDC) http.Handler {
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
		info := mdc.Info()
		info["requirement_id"] = "ENG-E-010"
		info["service"] = "eng-e-010"
		info["title"] = "Multi-data-center operation (40+)"
		info["owns"] = []string{"data_centers", "failure_domains", "fanout"}
		info["time"] = time.Now().UTC()
		writeJSON(w, info)
	})
	mux.HandleFunc("/v1/dcs", func(w http.ResponseWriter, r *http.Request) {
		handleDCs(w, r, mdc)
	})
	mux.HandleFunc("/v1/fanout", func(w http.ResponseWriter, r *http.Request) {
		handleFanout(w, r, mdc)
	})
	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		handleDomains(w, r, mdc)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, mdc)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_e_010_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_010_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_010_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleDCs(w http.ResponseWriter, r *http.Request, mdc *MultiDC) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, mdc.List())
	case http.MethodPost:
		var req struct {
			ID      string `json:"id"`
			Domain  string `json:"failure_domain"`
			Healthy *bool  `json:"healthy"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		healthy := true
		if req.Healthy != nil {
			healthy = *req.Healthy
		}
		if err := mdc.RegisterDC(req.ID, req.Domain, healthy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{"id": req.ID, "failure_domain": req.Domain, "healthy": healthy})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFanout(w http.ResponseWriter, r *http.Request, mdc *MultiDC) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var cfg map[string]any
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&cfg); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	result, err := mdc.Fanout(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

func handleDomains(w http.ResponseWriter, r *http.Request, mdc *MultiDC) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]any{
		"failure_domains": mdc.Domains(),
		"data_centers":    mdc.Count(),
	})
}

func handleDemo(w http.ResponseWriter, r *http.Request, mdc *MultiDC) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewMultiDC()
	if err := demo.SeedDemo(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fanout, err := demo.Fanout(map[string]any{
		"revision": "demo-cfg",
		"feature":  "multidc-fanout",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains := demo.Domains()
	info := demo.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-010",
		"service":        "eng-e-010",
		"run":            run,
		"message":        "live multi-DC simulator demo (≥40 DCs + fan-out)",
		"acceptance": []string{
			"≥40 simulated data centers registered",
			"failure domains group DCs",
			"fan-out reports partial success for unhealthy DCs",
			"labeled multi_dc_simulator (not physical)",
		},
		"proof": map[string]any{
			"data_centers":       demo.Count(),
			"failure_domains":    len(domains),
			"fanout_ok":          fanout.FanoutOK,
			"fanout_pushed":      fanout.Pushed,
			"fanout_failed":      fanout.Failed,
			"simulator":          info["simulator"] == true,
			"multi_dc_simulator": info["multi_dc_simulator"] == true,
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
