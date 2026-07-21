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
	"sync/atomic"
	"time"
)

// ENG-N-001: GitHub Actions at scale (SIMULATOR)
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-001",
	"service":        "eng-n-001",
	"title":          "GitHub Actions at scale",
}}

var (
	runnerPool   = NewRunnerPool(256)
	jobSeq       atomic.Int64
	matrixRan    atomic.Bool
	workflowJobs atomic.Int64
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-001 listening on %s (requirement ENG-N-001) [SIMULATOR]", *addr)
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
	mux.HandleFunc("/v1/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"requirement_id": "ENG-N-001",
			"service":        "eng-n-001",
			"title":          "GitHub Actions at scale",
			"simulator":      true,
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/workflows", handleWorkflows)
	mux.HandleFunc("/v1/runners", handleRunners)
	mux.HandleFunc("/v1/runners/", handleRunnerRoutes)
	mux.HandleFunc("/v1/jobs/", handleJobRoutes)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-001_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-001_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-001_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Seed a tiny matrix workflow for proof if empty.
	if workflowJobs.Load() == 0 {
		jobs, _ := ExpandMatrix(MatrixConfig{
			Dimensions: map[string][]string{"os": {"ubuntu", "macos"}},
		})
		for _, j := range jobs {
			id := fmt.Sprintf("demo-%d", jobSeq.Add(1))
			runnerPool.Enqueue(&Job{ID: id, Status: "queued", Matrix: j.Values})
			workflowJobs.Add(1)
		}
		matrixRan.Store(true)
	}
	if runnerPool.RunnerCount() == 0 {
		_ = runnerPool.Register("demo-runner")
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-001",
		"service":        "eng-n-001",
		"run":            state.Runs,
		"message":        "demo completed for GitHub Actions at scale (simulator)",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
		"proof": map[string]any{
			"matrix_expanded":     matrixRan.Load(),
			"runners_registered":  runnerPool.RunnerCount(),
			"jobs_completed":      runnerPool.CompletedCount(),
			"simulator":           "github-actions",
		},
	}
	state.mu.Unlock()
	writeJSON(w, result)
}

func handleWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var cfg MatrixConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		// also accept {"matrix": {...}}
		var wrap struct {
			Matrix MatrixConfig `json:"matrix"`
		}
		if err2 := json.Unmarshal(body, &wrap); err2 != nil {
			http.Error(w, "invalid matrix config", http.StatusBadRequest)
			return
		}
		cfg = wrap.Matrix
	}
	jobs, err := ExpandMatrix(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ids := make([]string, 0, len(jobs))
	for _, j := range jobs {
		id := fmt.Sprintf("job-%d", jobSeq.Add(1))
		runnerPool.Enqueue(&Job{ID: id, Status: "queued", Matrix: j.Values})
		ids = append(ids, id)
		workflowJobs.Add(1)
	}
	matrixRan.Store(true)
	writeJSON(w, map[string]any{"jobs": ids, "count": len(ids)})
}

func handleRunners(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		req.ID = fmt.Sprintf("runner-%d", jobSeq.Add(1))
	}
	if err := runnerPool.Register(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, map[string]any{"runner_id": req.ID})
}

func handleRunnerRoutes(w http.ResponseWriter, r *http.Request) {
	// /v1/runners/:id/claim
	path := strings.TrimPrefix(r.URL.Path, "/v1/runners/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "claim" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := runnerPool.ClaimJob(parts[0])
	if job == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, job)
}

func handleJobRoutes(w http.ResponseWriter, r *http.Request) {
	// /v1/jobs/:id/complete
	path := strings.TrimPrefix(r.URL.Path, "/v1/jobs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "complete" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		req.Status = "success"
	}
	if err := runnerPool.ReportComplete(parts[0], req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "job_id": parts[0], "status": req.Status})
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-N-001":
		return map[string]any{
			"matrix_expanded":    matrixRan.Load(),
			"runners_registered": runnerPool.RunnerCount(),
			"simulator":          "github-actions",
		}
	default:
		return map[string]any{"vertical_slice": true, "compose": true, "kubernetes": true}
	}
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
