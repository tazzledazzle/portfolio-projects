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

// ENG-E-019: Distributed systems design — task queue (retries, idempotency, partitions).
// Boundary: not event bus/DLQ (E-020), not durable workflows (H-006), not load SLO (E-009).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-019",
	"service":        "eng-e-019",
	"title":          "Distributed systems design",
}}

var queue = NewQueue(8)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-e-019 listening on %s (requirement ENG-E-019)", *addr)
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
	mux.HandleFunc("/v1/tasks", handleTasks)
	mux.HandleFunc("/v1/ack", handleAck)
	mux.HandleFunc("/v1/nack", handleNack)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		runs := state.Runs
		state.mu.Unlock()
		stats := queue.Stats()
		fmt.Fprintf(w, "# HELP eng_e_019_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_019_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_019_demo_runs_total %d\n", runs)
		fmt.Fprintf(w, "# HELP eng_e_019_duplicate_suppressed_total Idempotent duplicates suppressed\n")
		fmt.Fprintf(w, "# TYPE eng_e_019_duplicate_suppressed_total counter\n")
		fmt.Fprintf(w, "eng_e_019_duplicate_suppressed_total %d\n", stats["duplicate_suppressed"])
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := queue.Info()
	info["time"] = time.Now().UTC()
	info["stats"] = queue.Stats()
	writeJSON(w, info)
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Payload        string `json:"payload"`
		IdempotencyKey string `json:"idempotency_key"`
		Partition      int    `json:"partition"`
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	task, err := queue.Enqueue(req.Payload, req.IdempotencyKey, req.Partition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "task": task})
}

func handleAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := queue.Ack(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "id": req.ID, "status": "acked"})
}

func handleNack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := queue.Nack(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	task := queue.Get(req.ID)
	writeJSON(w, map[string]any{"ok": true, "task": task})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demoQ := NewQueue(4)
	t1, _ := demoQ.Enqueue("work-a", "demo-idem-1", 1)
	t2, _ := demoQ.Enqueue("work-a-dup", "demo-idem-1", 1)
	claimed := demoQ.Claim()
	_ = demoQ.Nack(claimed.ID)
	afterNack := demoQ.Get(claimed.ID)
	claimed2 := demoQ.Claim()
	_ = demoQ.Ack(claimed2.ID)
	stats := demoQ.Stats()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-019",
		"service":        "eng-e-019",
		"run":            state.Runs,
		"message":        "task queue demo: retries, idempotency keys, partitions",
		"acceptance": []string{
			"enqueue assigns partition",
			"duplicate idempotency key → duplicate_suppressed",
			"nack increments attempts and requeues",
			"ack completes task",
		},
		"proof": map[string]any{
			"idempotent":           t2.DuplicateSuppressed && t2.ID == t1.ID,
			"duplicate_suppressed": t2.DuplicateSuppressed && stats["duplicate_suppressed"].(int) >= 1,
			"partitions":           t1.Partition == 1 && stats["partitions"].(int) == 4,
			"retries":              afterNack != nil && afterNack.Attempts >= 1,
			"task_id":              t1.ID,
			"partition":            t1.Partition,
			"attempts_after_nack":  afterNack.Attempts,
			"stats":                stats,
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
