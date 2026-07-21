package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ENG-N-008: Heterogeneous deployment environments.
// Profile abstraction + scheduling the SAME workload across >=3 heterogeneous
// profiles (k8s-standard, k8s-gpu, vm-bake) per D-03/D-11. No canary weights,
// burn-rate gates, or finalizers here.

type DemoState struct {
	mu       sync.Mutex
	Runs     int    `json:"runs"`
	LastDemo string `json:"last_demo"`
}

var state = &DemoState{}

var store = NewProfileStore()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-008 listening on %s (requirement ENG-N-008)", *addr)
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
			"requirement_id": "ENG-N-008",
			"service":        "eng-n-008",
			"title":          "Heterogeneous deployment environments",
			"owns":           "profile abstraction + schedule same workload across >=3 profiles",
			"profiles":       []string{"k8s-standard", "k8s-gpu", "vm-bake"},
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/profiles/", handleProfiles)
	mux.HandleFunc("/v1/workloads/", handleWorkloads)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-008_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-008_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-008_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleProfiles(w http.ResponseWriter, r *http.Request) {
	name := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/profiles/"), "/")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Runtime     string            `json:"runtime"`
		Constraints map[string]string `json:"constraints"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	prof, err := store.UpsertProfile(name, body.Runtime, body.Constraints)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, prof)
}

func handleWorkloads(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/workloads/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	if parts[1] == "schedule" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var body struct {
			Profiles []string `json:"profiles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		placements, err := store.Schedule(id, body.Profiles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, placements)
		return
	}
	if parts[1] == "placements" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		placements, err := store.Placements(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, placements)
		return
	}
	http.NotFound(w, r)
}

// handleDemo registers all three profiles then schedules one workload across
// them, proving heterogeneous placement of the same workload identity.
func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewProfileStore()
	seed := []struct {
		name, runtime string
		constraints   map[string]string
	}{
		{"k8s-standard", "kubernetes", map[string]string{"gpu": "none"}},
		{"k8s-gpu", "kubernetes", map[string]string{"gpu": "nvidia"}},
		{"vm-bake", "vm", map[string]string{"image": "baked"}},
	}
	for _, p := range seed {
		if _, err := demo.UpsertProfile(p.name, p.runtime, p.constraints); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	const workload = "checkout-api"
	placements, err := demo.Schedule(workload, []string{"k8s-standard", "k8s-gpu", "vm-bake"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sameWorkload := true
	names := make([]string, 0, len(placements))
	for _, p := range placements {
		if p.WorkloadID != workload {
			sameWorkload = false
		}
		names = append(names, p.Profile)
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-008",
		"service":        "eng-n-008",
		"run":            state.Runs,
		"message":        "demo completed for Heterogeneous deployment environments",
		"acceptance": []string{
			"healthz returns 200",
			"register >=3 heterogeneous profiles",
			"schedule same workload across all profiles",
		},
		"proof": map[string]any{
			"profiles":       len(demo.Profiles()),
			"same_workload":  sameWorkload,
			"placements":     len(placements),
			"profile_names":  names,
			"workload":       workload,
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
