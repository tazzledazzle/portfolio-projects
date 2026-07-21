package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ENG-E-007: Design platform subsystems — ADR ≥2 alternatives + thin skeleton.
// Does not own full prod metrics (E-008), HPA (E-013), or Phase 7 mentoring kits.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-007",
	"service":        "eng-e-007",
	"title":          "Design platform subsystems",
}}

var design *DesignStore

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	var err error
	design, err = NewDesignStore("adr")
	if err != nil {
		log.Fatalf("load ADRs: %v", err)
	}

	mux := newMux()
	log.Printf("eng-e-007 listening on %s (requirement ENG-E-007)", *addr)
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
	mux.HandleFunc("/v1/adrs", handleADRs)
	mux.HandleFunc("/v1/skeleton", handleSkeleton)
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-007_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-007_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-007_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func ensureDesign() *DesignStore {
	if design != nil {
		return design
	}
	s, err := NewDesignStore("adr")
	if err != nil {
		return &DesignStore{}
	}
	design = s
	return design
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := ensureDesign().Info()
	info["time"] = time.Now().UTC()
	writeJSON(w, info)
}

func handleADRs(w http.ResponseWriter, r *http.Request) {
	adrs := ensureDesign().ListADRs()
	// Strip full content from list response to keep payloads small.
	summaries := make([]map[string]any, 0, len(adrs))
	for _, a := range adrs {
		summaries = append(summaries, map[string]any{
			"id":                 a.ID,
			"title":              a.Title,
			"path":               a.Path,
			"alternative_count":  a.AlternativeCount,
			"decision_recorded":  a.DecisionRecorded,
		})
	}
	writeJSON(w, map[string]any{"adrs": summaries, "count": len(summaries)})
}

func handleSkeleton(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, ensureDesign().Skeleton())
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	store := ensureDesign()
	adrs := store.ListADRs()
	skel := store.Skeleton()
	altMax := 0
	for _, a := range adrs {
		if a.AlternativeCount > altMax {
			altMax = a.AlternativeCount
		}
	}
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-007",
		"service":        "eng-e-007",
		"run":            state.Runs,
		"message":        "demo completed for Design platform subsystems",
		"acceptance": []string{
			"ADR list ≥1",
			"≥2 alternatives documented",
			"skeleton references ADR ID",
		},
		"proof": map[string]any{
			"adr_count":          len(adrs),
			"alternatives":       altMax,
			"decision_recorded":  store.DecisionRecorded(),
			"skeleton_adr_id":    skel["adr_id"],
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
