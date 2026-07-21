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

// ENG-E-009: Scale to thousands of workflows — load + backpressure + SLO latency.
// Boundary: not durable engine (H-006), not queue idempotency (E-019), not multi-DC (E-010).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-009",
	"service":        "eng-e-009",
	"title":          "Scale to thousands of application instances and workflows",
}}

var scaleSim = NewScaleSim(256)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-e-009 listening on %s (requirement ENG-E-009)", *addr)
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
	mux.HandleFunc("/v1/simulate", handleSimulate)
	mux.HandleFunc("/v1/metrics", handleJSONMetrics)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		runs := state.Runs
		state.mu.Unlock()
		m := scaleSim.Metrics()
		fmt.Fprintf(w, "# HELP eng_e_009_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_009_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_009_demo_runs_total %d\n", runs)
		fmt.Fprintf(w, "# HELP eng_e_009_p99_ms Admission latency p99\n")
		fmt.Fprintf(w, "# TYPE eng_e_009_p99_ms gauge\n")
		fmt.Fprintf(w, "eng_e_009_p99_ms %v\n", m["p99_ms"])
		fmt.Fprintf(w, "# HELP eng_e_009_queue_depth Simulated queue depth\n")
		fmt.Fprintf(w, "# TYPE eng_e_009_queue_depth gauge\n")
		fmt.Fprintf(w, "eng_e_009_queue_depth %v\n", m["queue_depth"])
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := scaleSim.Info()
	info["time"] = time.Now().UTC()
	info["metrics"] = scaleSim.Metrics()
	writeJSON(w, info)
}

func handleSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Count <= 0 {
		req.Count = 1000
	}
	res, err := scaleSim.Simulate(req.Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "result": res})
}

func handleJSONMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, scaleSim.Metrics())
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fresh sim with capacity that saturates under 1000 → backpressure proof.
	demo := NewScaleSim(128)
	res, err := demo.Simulate(1000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m := demo.Metrics()
	info := demo.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-009",
		"service":        "eng-e-009",
		"run":            state.Runs,
		"message":        "load simulation ≥1000 workflows with backpressure and SLO latency",
		"acceptance": []string{
			"workflows_simulated >= 1000",
			"backpressure when queue saturated",
			"p99_ms and queue_depth present",
			"not a durable workflow engine",
		},
		"proof": map[string]any{
			"workflows_simulated": res.WorkflowsSimulated,
			"backpressure":        res.Backpressure,
			"p99_ms":              res.P99Ms,
			"queue_depth":         m["queue_depth"],
			"rejected":            res.Rejected,
			"delayed":             res.Delayed,
			"accepted":            res.Accepted,
			"durable_workflow":    false,
			"not_durable_engine":  info["durable_workflow"] == false,
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
