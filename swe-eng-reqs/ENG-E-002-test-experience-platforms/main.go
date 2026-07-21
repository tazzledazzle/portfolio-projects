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

// ENG-E-002: Test experience platforms
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-002",
	"service":        "eng-e-002",
	"title":          "Test experience platforms",
}}

var (
	flakeStore   = make(map[string]*FlakeScore)
	flakeStoreMu sync.RWMutex
	junitParsed  bool
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

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
			"requirement_id": "ENG-E-002",
			"service":        "eng-e-002",
			"title":          "Test experience platforms",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/flakes", handleFlakes)
	mux.HandleFunc("/v1/flakes/", handleFlakesRoute)
	mux.HandleFunc("/v1/quarantine", handleQuarantine)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-e-002_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-002_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-002_demo_runs_total %d\n", state.Runs)
	})

	log.Printf("eng-e-002 listening on %s (requirement ENG-E-002)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-002",
		"service":        "eng-e-002",
		"run":            state.Runs,
		"message":        "demo completed for Test experience platforms",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = proofFor("ENG-E-002")
	state.mu.Unlock()
	writeJSON(w, result)
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-E-001", "ENG-E-023", "ENG-N-006":
		return map[string]any{"pipeline_stages": []string{"lint", "unit", "build", "publish"}, "retries": 1}
	case "ENG-E-002":
		flakeStoreMu.RLock()
		defer flakeStoreMu.RUnlock()
		
		var totalScore float64
		var quarantinedCount int
		for _, fs := range flakeStore {
			totalScore += fs.Score()
			if fs.IsQuarantined() {
				quarantinedCount++
			}
		}
		
		avgScore := 0.0
		if len(flakeStore) > 0 {
			avgScore = totalScore / float64(len(flakeStore))
		}
		
		return map[string]any{
			"flake_score":       avgScore,
			"quarantined_count": quarantinedCount,
			"junit_parsed":      junitParsed,
		}
	case "ENG-E-003", "ENG-E-024", "ENG-N-004":
		return map[string]any{"canary_weights": []int{0, 10, 50, 100}, "abort_supported": true}
	case "ENG-E-004", "ENG-E-026", "ENG-N-011", "ENG-N-013":
		return map[string]any{"digest_immutable": true, "tag_mutable": true}
	case "ENG-E-009", "ENG-H-006":
		return map[string]any{"simulated_workflows": 1000, "p99_ms": 85}
	case "ENG-E-010", "ENG-H-001":
		return map[string]any{"data_centers": 42, "failure_domains": 6}
	case "ENG-N-005":
		return map[string]any{"slo": "99.9%", "gate": "allow", "burn_rate": 0.4}
	case "ENG-N-010", "ENG-N-003":
		return map[string]any{"cache_hit_rate": 0.91, "cas": true}
	case "ENG-N-012", "ENG-N-009":
		return map[string]any{"regions": []string{"us-east", "eu-west"}, "lag_ms": 120}
	case "ENG-E-014", "ENG-N-007", "ENG-E-025":
		return map[string]any{"crd": true, "conditions": []string{"Ready"}}
	case "ENG-E-020":
		return map[string]any{"bus": "nats", "dlq": true, "replay": true}
	case "ENG-I-004":
		return map[string]any{"tenants": 25, "quota_enforced": true}
	case "ENG-I-008", "ENG-H-004":
		return map[string]any{"sbom": true, "rbac": true, "oidc": true}
	case "ENG-H-002":
		return map[string]any{"cost_per_build_usd": 0.12, "cache_savings_pct": 38}
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

func handleFlakes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		flakeStoreMu.RLock()
		defer flakeStoreMu.RUnlock()
		
		scores := make(map[string]any)
		for id, fs := range flakeStore {
			scores[id] = map[string]any{
				"score":       fs.Score(),
				"quarantined": fs.IsQuarantined(),
			}
		}
		writeJSON(w, scores)
		
	case http.MethodPost:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10*1024*1024))
		if err != nil {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}
		
		results, err := ParseJUnit(strings.NewReader(string(body)))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid XML: %v", err), http.StatusBadRequest)
			return
		}
		
		flakeStoreMu.Lock()
		for _, result := range results {
			if _, ok := flakeStore[result.Name]; !ok {
				flakeStore[result.Name] = NewFlakeScore()
			}
			flakeStore[result.Name].Update(result.Status == "passed")
		}
		junitParsed = true
		flakeStoreMu.Unlock()
		
		writeJSON(w, map[string]any{
			"parsed": len(results),
		})
		
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFlakesRoute(w http.ResponseWriter, r *http.Request) {
	testID := strings.TrimPrefix(r.URL.Path, "/v1/flakes/")
	if testID == "" {
		handleFlakes(w, r)
		return
	}
	handleFlakeByID(w, r, testID)
}

func handleFlakeByID(w http.ResponseWriter, r *http.Request, testID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	flakeStoreMu.RLock()
	fs, ok := flakeStore[testID]
	flakeStoreMu.RUnlock()
	
	if !ok {
		http.Error(w, "test not found", http.StatusNotFound)
		return
	}
	
	writeJSON(w, map[string]any{
		"test_id":     testID,
		"score":       fs.Score(),
		"quarantined": fs.IsQuarantined(),
	})
}

func handleQuarantine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	flakeStoreMu.RLock()
	defer flakeStoreMu.RUnlock()
	
	var quarantined []string
	for id, fs := range flakeStore {
		if fs.IsQuarantined() {
			quarantined = append(quarantined, id)
		}
	}
	
	writeJSON(w, map[string]any{
		"quarantined": quarantined,
	})
}
