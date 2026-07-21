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

// ENG-H-001: Failure-domain / blast-radius engineering
// Owns chaos containment proof (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-H-001",
		"service":        "eng-h-001",
		"title":          "Failure-domain / blast-radius engineering",
	}}
	store = NewBlastEngine()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-h-001 listening on %s (requirement ENG-H-001)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(store)))
}

func newMux(eng *BlastEngine) http.Handler {
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
			"requirement_id": "ENG-H-001",
			"service":        "eng-h-001",
			"title":          "Failure-domain / blast-radius engineering",
			"owns":           []string{"chaos_containment", "blast_radius"},
			"simulator":      true,
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/chaos", func(w http.ResponseWriter, r *http.Request) {
		handleChaos(w, r, eng)
	})
	mux.HandleFunc("/v1/blast-radius", func(w http.ResponseWriter, r *http.Request) {
		handleBlastRadius(w, r, eng)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, eng)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_h_001_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_h_001_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_h_001_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleChaos(w http.ResponseWriter, r *http.Request, eng *BlastEngine) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Domain   string   `json:"domain"`
		Scenario string   `json:"scenario"`
		Tenants  []string `json:"tenants"`
		Seed     bool     `json:"seed"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Seed || eng.BlastRadius().ChaosRan == false && len(req.Tenants) > 0 {
		_ = eng.RegisterDomain(req.Domain, req.Tenants)
	}
	result, err := eng.RunChaos(req.Domain, req.Scenario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

func handleBlastRadius(w http.ResponseWriter, r *http.Request, eng *BlastEngine) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, eng.BlastRadius())
}

func handleDemo(w http.ResponseWriter, r *http.Request, _ *BlastEngine) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewBlastEngine()
	chaos, err := demo.SeedDemo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	radius := demo.BlastRadius()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-H-001",
		"service":        "eng-h-001",
		"run":            run,
		"message":        "live chaos blast-radius containment demo",
		"acceptance": []string{
			"chaos runs in one failure domain",
			"blast radius contained with unaffected domains/tenants",
			"affected/unaffected sets are server-computed",
		},
		"proof": map[string]any{
			"chaos_ran":           chaos.ChaosRan,
			"contained":           chaos.Contained && radius.Contained,
			"affected_domains":    radius.AffectedDomains,
			"unaffected_domains":  radius.UnaffectedDomains,
			"unaffected_tenants":  radius.UnaffectedTenants,
			"affected_tenants":    radius.AffectedTenants,
			"scenario":            chaos.Scenario,
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
