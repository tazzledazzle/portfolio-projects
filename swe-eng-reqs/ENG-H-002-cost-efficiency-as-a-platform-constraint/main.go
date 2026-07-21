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

// ENG-H-002: Cost/efficiency as a platform constraint
// Owns cost-per-build meter + cache-hit savings (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-H-002",
		"service":        "eng-h-002",
		"title":          "Cost/efficiency as a platform constraint",
	}}
	meter = NewCostMeter()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-h-002 listening on %s (requirement ENG-H-002)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(meter)))
}

func newMux(m *CostMeter) http.Handler {
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
			"requirement_id": "ENG-H-002",
			"service":        "eng-h-002",
			"title":          "Cost/efficiency as a platform constraint",
			"owns":           []string{"cost_per_build_meter", "cache_hit_savings"},
			"does_not_own":  []string{"cas_cache", "tenant_quotas"},
			"time":          time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/builds", func(w http.ResponseWriter, r *http.Request) {
		handleRecordBuild(w, r, m)
	})
	mux.HandleFunc("/v1/cost-report", func(w http.ResponseWriter, r *http.Request) {
		handleCostReport(w, r, m)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, m)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_h_002_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_h_002_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_h_002_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleRecordBuild(w http.ResponseWriter, r *http.Request, m *CostMeter) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// T-5-21: only accept duration/resources/cache_hit — ignore client cost/savings fields
	var raw map[string]any
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&raw); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	in := BuildInput{
		DurationSec: asFloat(raw["duration_sec"]),
		CPUCores:    asFloat(raw["cpu_cores"]),
		MemoryGB:    asFloat(raw["memory_gb"]),
		CacheHit:    asBool(raw["cache_hit"]),
	}
	rec, err := m.RecordBuild(in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, rec)
}

func handleCostReport(w http.ResponseWriter, r *http.Request, m *CostMeter) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, m.Report())
}

func handleDemo(w http.ResponseWriter, r *http.Request, _ *CostMeter) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewCostMeter()
	report, err := demo.SeedDemo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-H-002",
		"service":        "eng-h-002",
		"run":            run,
		"message":        "live cost-per-build + cache savings demo",
		"acceptance": []string{
			"cost_per_build_usd computed server-side from resources",
			"cache_savings_pct computed from recorded hits/misses",
			"does not own CAS cache (N-010)",
		},
		"proof": map[string]any{
			"cost_per_build_usd": report.CostPerBuildUSD,
			"cache_savings_pct":  report.CacheSavingsPct,
			"builds":             report.Builds,
			"cache_hits":         report.CacheHits,
			"not_cas_owner":      true,
		},
	})
}

func asFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func asBool(v any) bool {
	b, _ := v.(bool)
	return b
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
