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

// ENG-E-020: Event-driven architecture — NATS-inspired in-memory bus (D-10).
// Honesty: nats_inspired + simulator; never bus:"nats" or live connectivity.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-020",
	"service":        "eng-e-020",
	"title":          "Event-driven architecture design",
}}

var bus = NewBus()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-e-020 listening on %s (requirement ENG-E-020)", *addr)
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
	mux.HandleFunc("/v1/publish", handlePublish)
	mux.HandleFunc("/v1/consume", handleConsume)
	mux.HandleFunc("/v1/replay", handleReplay)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		runs := state.Runs
		state.mu.Unlock()
		info := bus.Info()
		fmt.Fprintf(w, "# HELP eng_e_020_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_020_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_020_demo_runs_total %d\n", runs)
		fmt.Fprintf(w, "# HELP eng_e_020_log_length Bus log length\n")
		fmt.Fprintf(w, "# TYPE eng_e_020_log_length gauge\n")
		fmt.Fprintf(w, "eng_e_020_log_length %v\n", info["log_length"])
		fmt.Fprintf(w, "# HELP eng_e_020_dlq_depth Dead-letter queue depth\n")
		fmt.Fprintf(w, "# TYPE eng_e_020_dlq_depth gauge\n")
		fmt.Fprintf(w, "eng_e_020_dlq_depth %v\n", info["dlq_depth"])
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := bus.Info()
	info["time"] = time.Now().UTC()
	writeJSON(w, info)
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Subject string         `json:"subject"`
		Schema  string         `json:"schema"`
		Payload map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	env, err := bus.Publish(req.Subject, req.Schema, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "envelope": env})
}

func handleConsume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	consumer := r.URL.Query().Get("consumer")
	if consumer == "" {
		consumer = "default"
	}
	var failReason string
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Consumer string `json:"consumer"`
			Fail     string `json:"fail"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Consumer != "" {
			consumer = req.Consumer
		}
		failReason = req.Fail
	}
	env, err := bus.Consume(consumer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if env != nil && failReason != "" {
		_ = bus.Fail(env.ID, failReason)
	}
	writeJSON(w, map[string]any{"ok": true, "envelope": env, "dlq": failReason != ""})
}

func handleReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	offset := 0
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			FromOffset int `json:"from_offset"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		offset = req.FromOffset
	}
	msgs, err := bus.Replay(offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{
		"ok":           true,
		"from_offset":  offset,
		"replayed":     msgs,
		"count":        len(msgs),
		"nats_inspired": true,
		"simulator":    true,
		"note":         "Replay from in-memory log offset; not JetStream",
	})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewBus()
	env, _ := demo.Publish("demo.events", "demo.v1", map[string]any{"seq": 1})
	_, _ = demo.Publish("demo.events", "demo.v1", map[string]any{"seq": 2})
	got, _ := demo.Consume("demo-consumer")
	_ = demo.Fail(got.ID, "simulated handler failure")
	dlq := demo.DLQ()
	replayed, _ := demo.Replay(0)
	info := demo.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-020",
		"service":        "eng-e-020",
		"run":            state.Runs,
		"message":        "NATS-inspired bus demo (simulator; no live NATS)",
		"acceptance": []string{
			"schema envelopes with id/subject/schema/payload/ts",
			"consumers deliver published messages",
			"handler failure → DLQ",
			"replay from log offset",
			"nats_inspired + simulator honesty labels",
		},
		"proof": map[string]any{
			"nats_inspired":   info["nats_inspired"] == true,
			"simulator":       info["simulator"] == true,
			"schema_envelope": env != nil && env.Schema == "demo.v1" && env.ID != "",
			"dlq":             len(dlq) >= 1,
			"replay":          len(replayed) >= 2,
			"nats_connected":  false,
			"envelope_id":     env.ID,
			"dlq_reason":      dlq[0].Reason,
			"replay_count":    len(replayed),
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
