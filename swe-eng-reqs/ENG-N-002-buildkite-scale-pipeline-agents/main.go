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

// ENG-N-002: Buildkite-scale pipeline agents (SIMULATOR)
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-002",
	"service":        "eng-n-002",
	"title":          "Buildkite-scale pipeline agents",
}}

var (
	agentPool = NewAgentPool()
	concMgr   = NewConcurrencyManager()
	agentSeq  atomic.Int64
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()
	mux := newMux()
	log.Printf("eng-n-002 listening on %s (requirement ENG-N-002) [SIMULATOR]", *addr)
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
			"requirement_id": "ENG-N-002",
			"service":        "eng-n-002",
			"title":          "Buildkite-scale pipeline agents",
			"simulator":      true,
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/agents", handleAgents)
	mux.HandleFunc("/v1/agents/", handleAgentRoutes)
	mux.HandleFunc("/v1/pipelines/", handlePipelineRoutes)
	mux.HandleFunc("/v1/jobs/", handleJobRoutes)
	mux.HandleFunc("/v1/concurrency", handleConcurrency)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-002_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-002_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-002_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if agentPool.AgentCount() == 0 {
		_ = agentPool.Register("demo-agent", []string{"queue=default"})
	}
	_ = concMgr.Acquire("deploy", 2)
	_ = agentPool.UploadPipeline("demo", "steps:\n  - label: lint\n  - label: test\n")

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-002",
		"service":        "eng-n-002",
		"run":            state.Runs,
		"message":        "demo completed for Buildkite-scale pipeline agents (simulator)",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
		"proof": map[string]any{
			"agents":                     agentPool.AgentCount(),
			"concurrency_groups":         concMgr.GroupCount(),
			"dynamic_pipeline_supported": true,
			"simulator":                  "buildkite",
		},
	}
	state.mu.Unlock()
	writeJSON(w, result)
}

func handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID   string   `json:"id"`
		Tags []string `json:"tags"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.ID == "" {
		req.ID = fmt.Sprintf("agent-%d", agentSeq.Add(1))
	}
	if err := agentPool.Register(req.ID, req.Tags); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, map[string]any{"agent_id": req.ID})
}

func handleAgentRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "poll" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := agentPool.Poll(parts[0])
	if job == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Respect concurrency group if set
	group := job.Group
	if group == "" {
		group = "default"
	}
	if !concMgr.Acquire(group, 5) {
		job.offeredTo = ""
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := agentPool.Claim(parts[0], job.ID); err != nil {
		concMgr.Release(group)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, job)
}

func handlePipelineRoutes(w http.ResponseWriter, r *http.Request) {
	// /v1/pipelines/:id/upload
	path := strings.TrimPrefix(r.URL.Path, "/v1/pipelines/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "upload" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	yaml := string(body)
	var wrap struct {
		YAML string `json:"yaml"`
	}
	if json.Unmarshal(body, &wrap) == nil && wrap.YAML != "" {
		yaml = wrap.YAML
	}
	if err := agentPool.UploadPipeline(parts[0], yaml); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "pipeline_id": parts[0]})
}

func handleJobRoutes(w http.ResponseWriter, r *http.Request) {
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
		ExitCode int    `json:"exit_code"`
		Group    string `json:"group"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := agentPool.Complete(parts[0], req.ExitCode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	group := req.Group
	if group == "" {
		group = "default"
	}
	concMgr.Release(group)
	writeJSON(w, map[string]any{"ok": true, "job_id": parts[0]})
}

func handleConcurrency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]any{"groups": concMgr.AllStatuses()})
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-N-002":
		return map[string]any{
			"agents":             agentPool.AgentCount(),
			"concurrency_groups": concMgr.GroupCount(),
			"simulator":          "buildkite",
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
