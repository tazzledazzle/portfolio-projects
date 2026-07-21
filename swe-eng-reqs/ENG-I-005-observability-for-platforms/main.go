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

// ENG-I-005: Observability for platforms
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-I-005",
	"service":        "eng-i-005",
	"title":          "Observability for platforms",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewOTelStore())
	log.Printf("eng-i-005 listening on %s (requirement ENG-I-005)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(store *OTelStore) http.Handler {
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
		info := store.Info()
		info["time"] = time.Now().UTC()
		writeJSON(w, info)
	})
	mux.HandleFunc("/v1/trace", func(w http.ResponseWriter, r *http.Request) {
		handleTrace(w, r, store)
	})
	mux.HandleFunc("/v1/traces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.ExportTraces())
	})
	mux.HandleFunc("/v1/alerts/evaluate", func(w http.ResponseWriter, r *http.Request) {
		handleAlert(w, r, store)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, store)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		runs := state.Runs
		state.mu.Unlock()
		_, _ = fmt.Fprintf(w, "# HELP eng_i_005_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_i_005_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_i_005_demo_runs_total %d\n", runs)
		for name, value := range store.Metrics() {
			_, _ = fmt.Fprintf(w, "eng_i_005_%s %d\n", name, value)
		}
	})
	return mux
}

func handleTrace(w http.ResponseWriter, r *http.Request, store *OTelStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Name string `json:"name"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	span, err := store.StartSpan(request.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ended, err := store.EndSpan(span.TraceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, ended)
}

func handleAlert(w http.ResponseWriter, r *http.Request, store *OTelStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		RuleID  string    `json:"rule_id"`
		Samples []float64 `json:"samples"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	result, err := store.EvaluateAlerts(request.RuleID, request.Samples)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

func handleDemo(w http.ResponseWriter, r *http.Request, store *OTelStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	span, err := store.StartSpan("platform.demo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := store.EndSpan(span.TraceID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	export := store.ExportTraces()
	alert, err := store.EvaluateAlerts("high-error-rate", []float64{0.01, 0.08})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	metrics := store.Metrics()
	info := store.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-005",
		"service":        "eng-i-005",
		"run":            state.Runs,
		"message":        "demo completed for Observability for platforms",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = map[string]any{
		"spans_exported":        len(export.Spans),
		"metrics_exported":      len(metrics),
		"alert_rules":           store.AlertRuleCount(),
		"alert_status":          alert.Status,
		"otel_inspired":         info["otel_inspired"],
		"simulator":             info["simulator"],
		"instrumentation_model": info["instrumentation_model"],
		"collector":             info["collector"],
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
