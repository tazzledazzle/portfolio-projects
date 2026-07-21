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

// ENG-I-009: Support AI/GPU-adjacent developer workloads as platform consumers
// Long-running job + chunked large-artifact upload (Boundary Matrix D-03).
// Does NOT own: H-006 durable workflow engine; E-013 HPA packaging.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state    = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-I-009",
		"service":        "eng-i-009",
		"title":          "Support AI/GPU-adjacent developer workloads as platform consumers",
	}}
	consumer = NewGPUConsumer()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-i-009 listening on %s (requirement ENG-I-009)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(consumer)))
}

func newMux(c *GPUConsumer) http.Handler {
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
			"requirement_id":    "ENG-I-009",
			"service":           "eng-i-009",
			"title":             "Support AI/GPU-adjacent developer workloads as platform consumers",
			"owns":              []string{"long_running_job", "chunked_large_artifact_upload"},
			"does_not_own":      []string{"durable_workflow_engine", "gpu_kernel_scheduling", "hpa_packaging"},
			"max_chunk_bytes":   DefaultMaxChunkBytes,
			"max_total_bytes":   DefaultMaxTotalBytes,
			"simulator":         true,
			"time":              time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		handleJobs(w, r, c)
	})
	mux.HandleFunc("/v1/jobs/", func(w http.ResponseWriter, r *http.Request) {
		handleJobChunks(w, r, c)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, c)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_i_009_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_i_009_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_i_009_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleJobs(w http.ResponseWriter, r *http.Request, c *GPUConsumer) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name          string `json:"name"`
		TimeoutMS     int64  `json:"timeout_ms"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "unnamed"
	}
	timeout := DefaultJobTimeout
	if req.TimeoutMS > 0 {
		timeout = time.Duration(req.TimeoutMS) * time.Millisecond
	}
	job, err := c.StartJob(req.Name, timeout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, job)
}

func handleJobChunks(w http.ResponseWriter, r *http.Request, c *GPUConsumer) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/jobs/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "chunks" {
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
	// Chunk body capped above typical MaxBytesReader 1<<20 for multi-chunk demos,
	// but still bounded by DefaultMaxChunkBytes in UploadChunk (T-5-15).
	limited := http.MaxBytesReader(w, r.Body, int64(DefaultMaxChunkBytes)+1024)
	data, err := io.ReadAll(limited)
	if err != nil {
		http.Error(w, "chunk too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}
	index := 0
	if v := r.URL.Query().Get("index"); v != "" {
		fmt.Sscanf(v, "%d", &index)
	}
	job, err := c.UploadChunk(id, data, index)
	if err != nil {
		status := http.StatusBadRequest
		switch err {
		case ErrJobNotFound:
			status = http.StatusNotFound
		case ErrJobTimedOut:
			status = http.StatusGone
		case ErrChunkTooLarge, ErrTotalCapExceeded:
			status = http.StatusRequestEntityTooLarge
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, job)
}

func handleDemo(w http.ResponseWriter, r *http.Request, c *GPUConsumer) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	job, err := c.StartJob(fmt.Sprintf("demo-%d", time.Now().UnixNano()), 5*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunks := [][]byte{
		[]byte("gpu-artifact-part-1|"),
		[]byte("gpu-artifact-part-2|"),
		[]byte("gpu-artifact-part-3"),
	}
	var last *GPUJob
	for i, ch := range chunks {
		last, err = c.UploadChunk(job.ID, ch, i)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Prove timeout path with a separate short-lived job (T-5-15 companion).
	short, err := c.StartJob("demo-timeout", 20*time.Millisecond)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	time.Sleep(35 * time.Millisecond)
	timeoutRejected := false
	if _, err := c.UploadChunk(short.ID, []byte("late"), 0); err == ErrJobTimedOut {
		timeoutRejected = true
	}
	shortGot, _ := c.GetJob(short.ID)

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-009",
		"service":        "eng-i-009",
		"run":            run,
		"message":        "GPU-adjacent long job + chunked upload demo",
		"acceptance": []string{
			"long_running job with timeout deadline",
			"chunked_upload accumulates bytes_received",
			"expired job rejects further chunks",
		},
		"proof": map[string]any{
			"long_running":     last.LongRunning,
			"timeout":          shortGot != nil && (shortGot.TimedOut || shortGot.Status == "timeout"),
			"timeout_rejected": timeoutRejected,
			"chunked_upload":   last.ChunkedUpload,
			"bytes_received":   last.BytesReceived,
			"chunks":           last.Chunks,
			"job_id":           last.ID,
			"max_chunk_bytes":  DefaultMaxChunkBytes,
			"not_durable_wf":   true,
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
