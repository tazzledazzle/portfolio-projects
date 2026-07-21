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

// ENG-I-004: Multi-tenant consumer isolation
// Owns tenant quotas + noisy-neighbor limits (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-I-004",
		"service":        "eng-i-004",
		"title":          "Multi-tenant consumer isolation",
	}}
	store = NewTenantScheduler()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-i-004 listening on %s (requirement ENG-I-004)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(store)))
}

func newMux(sched *TenantScheduler) http.Handler {
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
			"requirement_id": "ENG-I-004",
			"service":        "eng-i-004",
			"title":          "Multi-tenant consumer isolation",
			"owns":           []string{"tenant_quotas", "noisy_neighbor_limits"},
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/quotas", func(w http.ResponseWriter, r *http.Request) {
		handleQuotas(w, r, sched)
	})
	mux.HandleFunc("/v1/schedule", func(w http.ResponseWriter, r *http.Request) {
		handleSchedule(w, r, sched)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, sched)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_i_004_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_i_004_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_i_004_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleQuotas(w http.ResponseWriter, r *http.Request, sched *TenantScheduler) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TenantID  string `json:"tenant_id"`
		Quota     int    `json:"quota"`
		RateLimit int    `json:"rate_limit"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := sched.SetQuota(req.TenantID, req.Quota); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.RateLimit > 0 {
		if err := sched.SetRateLimit(req.TenantID, req.RateLimit); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	writeJSON(w, map[string]any{
		"tenant_id":  req.TenantID,
		"quota":      req.Quota,
		"rate_limit": req.RateLimit,
	})
}

func handleSchedule(w http.ResponseWriter, r *http.Request, sched *TenantScheduler) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TenantID string `json:"tenant_id"`
		Units    int    `json:"units"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Units == 0 {
		req.Units = 1
	}
	result, err := sched.Schedule(req.TenantID, req.Units)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrQuotaExceeded || err == ErrRateLimited {
			status = http.StatusTooManyRequests
		}
		if err == ErrMissingTenantID {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, result)
}

func handleDemo(w http.ResponseWriter, r *http.Request, _ *TenantScheduler) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewTenantScheduler()
	if err := demo.SeedDemo(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proof := demo.Proof()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-004",
		"service":        "eng-i-004",
		"run":            run,
		"message":        "live tenant quota + noisy-neighbor isolation demo",
		"acceptance": []string{
			"schedule beyond quota denied",
			"noisy tenant rate-limited while quiet tenant schedules",
			"missing tenant id rejected",
		},
		"proof": proof,
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
