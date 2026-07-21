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

// ENG-N-005: SLO-based quality gates
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-005",
	"service":        "eng-n-005",
	"title":          "SLO-based quality gates",
}}

var gates = NewGateEngine()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-005 listening on %s (requirement ENG-N-005)", *addr)
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
	mux.HandleFunc("/v1/slos/", handleSLO)
	mux.HandleFunc("/v1/series/", handleSeries)
	mux.HandleFunc("/v1/gates/evaluate", handleEvaluate)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-005_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-005_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-005_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := gates.Info()
	info["time"] = time.Now().UTC()
	writeJSON(w, info)
}

func handleSLO(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/slos/"), "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Objective float64 `json:"objective"`
		Threshold float64 `json:"threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := gates.PutSLO(id, body.Objective, body.Threshold); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": id, "objective": body.Objective, "threshold": body.Threshold})
}

func handleSeries(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/series/"), "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body SeriesSample
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := gates.IngestSeries(id, body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "slo_id": id})
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		SloID    string  `json:"slo_id"`
		BurnRate float64 `json:"burn_rate"` // ignored — evidence is server-side (T-4-03)
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	_ = body.BurnRate // explicitly discard client-supplied burn
	dec, err := gates.Evaluate(body.SloID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, dec)
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewGateEngine()
	_ = demo.PutSLO("allow-slo", 0.999, 14.4)
	_ = demo.IngestSeries("allow-slo", SeriesSample{
		ErrorsShort: 0, TotalShort: 1000,
		ErrorsLong:  1, TotalLong: 100000,
	})
	allowDec, err := demo.Evaluate("allow-slo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = demo.PutSLO("deny-slo", 0.999, 14.4)
	_ = demo.IngestSeries("deny-slo", SeriesSample{
		ErrorsShort: 50, TotalShort: 100,
		ErrorsLong:  500, TotalLong: 1000,
	})
	denyDec, err := demo.Evaluate("deny-slo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info := demo.Info()
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-005",
		"service":        "eng-n-005",
		"run":            state.Runs,
		"message":        "demo completed for SLO-based quality gates",
		"acceptance": []string{
			"healthz returns 200",
			"multi-window AND burn-rate allow/deny",
			"evidence computed server-side (promql_inspired simulator)",
		},
		"proof": map[string]any{
			"allow":           allowDec.Decision == "allow",
			"deny":            denyDec.Decision == "deny",
			"gate":            denyDec.Decision,
			"burn_rate":       denyDec.BurnRate,
			"evidence":        denyDec.Evidence,
			"promql_inspired": info["promql_inspired"] == true,
			"simulator":       info["simulator"] == true,
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
