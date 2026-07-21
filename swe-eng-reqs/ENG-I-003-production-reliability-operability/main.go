package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ENG-I-003: Production reliability / operability
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-I-003",
	"service":        "eng-i-003",
	"title":          "Production reliability / operability",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewReliabilityStore("runbooks"))
	log.Printf("eng-i-003 listening on %s (requirement ENG-I-003)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(store *ReliabilityStore) http.Handler {
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
			"requirement_id": "ENG-I-003",
			"service":        "eng-i-003",
			"title":          "Production reliability / operability",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/slos/", func(w http.ResponseWriter, r *http.Request) {
		handleSLO(w, r, store)
	})
	mux.HandleFunc("/v1/golden-signals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.GoldenSignals())
	})
	mux.HandleFunc("/v1/runbooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		runbooks, err := store.ListRunbooks()
		if err != nil {
			http.Error(w, "runbook index unavailable", http.StatusInternalServerError)
			return
		}
		writeJSON(w, runbooks)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, store)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		_, _ = fmt.Fprintf(w, "# HELP eng_i_003_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_i_003_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_i_003_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleSLO(w http.ResponseWriter, r *http.Request, store *ReliabilityStore) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/slos/"), "/")
	var request struct {
		Objective float64 `json:"objective"`
		SLI       string  `json:"sli"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	slo, err := store.PutSLO(id, request.Objective, request.SLI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, slo)
}

func handleDemo(w http.ResponseWriter, r *http.Request, store *ReliabilityStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := fmt.Sprintf("demo-%d", time.Now().UTC().UnixNano())
	slo, err := store.PutSLO(id, 0.999, "successful_requests / total_requests")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runbooks, err := store.ListRunbooks()
	if err != nil {
		http.Error(w, "runbook index unavailable", http.StatusInternalServerError)
		return
	}
	signals := store.GoldenSignals()
	signalNames := []string{"latency", "traffic", "errors", "saturation"}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-003",
		"service":        "eng-i-003",
		"run":            state.Runs,
		"message":        "demo completed for Production reliability / operability",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = map[string]any{
		"slo":            slo.ID != "" && slo.Objective == 0.999,
		"slo_id":         slo.ID,
		"golden_signals": signalNames,
		"signal_metrics": signals,
		"runbook_count":  len(runbooks),
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
