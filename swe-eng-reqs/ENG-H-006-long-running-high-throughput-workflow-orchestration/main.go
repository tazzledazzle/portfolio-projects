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

// ENG-H-006: Long-running / high-throughput workflow orchestration
// Durable multi-step workflow MVP + throughput (Boundary Matrix D-03).
// Does NOT own: E-009 load-only sim, E-019 queue partitions, I-009 GPU chunk upload.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state  = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-H-006",
		"service":        "eng-h-006",
		"title":          "Long-running / high-throughput workflow orchestration",
	}}
	engine = NewWorkflowEngine()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-h-006 listening on %s (requirement ENG-H-006)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(engine)))
}

func newMux(eng *WorkflowEngine) http.Handler {
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
			"requirement_id": "ENG-H-006",
			"service":        "eng-h-006",
			"title":          "Long-running / high-throughput workflow orchestration",
			"owns":           []string{"durable_workflow", "throughput"},
			"does_not_own":  []string{"load_only_sim", "queue_partitions", "gpu_chunk_upload"},
			"simulator":     true,
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		handleWorkflows(w, r, eng)
	})
	mux.HandleFunc("/v1/workflows/", func(w http.ResponseWriter, r *http.Request) {
		handleWorkflowSignal(w, r, eng)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, eng)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		tp := eng.Throughput()
		fmt.Fprintf(w, "# HELP eng_h_006_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_h_006_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_h_006_demo_runs_total %d\n", state.Runs)
		fmt.Fprintf(w, "# HELP eng_h_006_signals_applied_total Durable signal applications\n")
		fmt.Fprintf(w, "# TYPE eng_h_006_signals_applied_total counter\n")
		fmt.Fprintf(w, "eng_h_006_signals_applied_total %d\n", tp.SignalsApplied)
	})
	return mux
}

func handleWorkflows(w http.ResponseWriter, r *http.Request, eng *WorkflowEngine) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name  string   `json:"name"`
		Steps []string `json:"steps"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Steps) == 0 {
		req.Steps = []string{"validate", "provision", "run", "finalize"}
	}
	if req.Name == "" {
		req.Name = "unnamed"
	}
	wf, err := eng.Start(req.Name, req.Steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, wf)
}

func handleWorkflowSignal(w http.ResponseWriter, r *http.Request, eng *WorkflowEngine) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/workflows/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "signal" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id := parts[0]
	if id == "" || strings.ContainsAny(id, `/\`) {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Action  string `json:"action"`
		EventID string `json:"event_id"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.EventID == "" {
		http.Error(w, "event_id required", http.StatusBadRequest)
		return
	}
	wf, err := eng.Signal(id, req.Action, req.EventID)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrWorkflowNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, wf)
}

func handleDemo(w http.ResponseWriter, r *http.Request, eng *WorkflowEngine) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	steps := []string{"validate", "provision", "run", "finalize"}
	wf, err := eng.Start(fmt.Sprintf("demo-%d", time.Now().UnixNano()), steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := 0; i < len(steps); i++ {
		eventID := fmt.Sprintf("demo-evt-%s-%d", wf.ID, i)
		if _, err := eng.Signal(wf.ID, "advance", eventID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Replay same event — must not double-apply (T-5-14).
		if _, err := eng.Signal(wf.ID, "advance", eventID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	got, ok := eng.Get(wf.ID)
	if !ok {
		http.Error(w, "workflow lost", http.StatusInternalServerError)
		return
	}
	tp := eng.Throughput()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-H-006",
		"service":        "eng-h-006",
		"run":            run,
		"message":        "durable workflow + throughput demo",
		"acceptance": []string{
			"durable multi-step workflow persists across Signal",
			"replay_safe: duplicate event_id does not double-apply",
			"throughput_per_s > 0 after batch signals",
		},
		"proof": map[string]any{
			"durable":          got.Durable,
			"throughput_per_s": tp.ThroughputPerS,
			"steps_completed":  got.StepsCompleted,
			"replay_safe":      got.ReplaySafe,
			"status":           got.Status,
			"workflow_id":      got.ID,
			"signals_applied":  tp.SignalsApplied,
			"not_load_only":    true,
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
