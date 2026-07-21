package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ENG-I-002: Platform-as-product mindset
// Owns SLAs, adoption metrics, golden-path template (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-I-002",
		"service":        "eng-i-002",
		"title":          "Platform-as-product mindset",
	}}
	productStore *ProductStore
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	templateDir := getenv("TEMPLATE_DIR", "templates")
	productStore = NewProductStore(templateDir)

	log.Printf("eng-i-002 listening on %s (requirement ENG-I-002)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(productStore)))
}

func newMux(store *ProductStore) http.Handler {
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
			"requirement_id": "ENG-I-002",
			"service":        "eng-i-002",
			"title":          "Platform-as-product mindset",
			"owns":           []string{"sla", "adoption", "golden_path"},
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/sla", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.SLA())
	})
	mux.HandleFunc("/v1/adoption", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.Adoption())
	})
	mux.HandleFunc("/v1/golden-path", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.GoldenPath())
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, store)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_i_002_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_i_002_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_i_002_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleDemo(w http.ResponseWriter, r *http.Request, store *ProductStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store.RecordAdoption("demo-team")
	sla := store.SLA()
	adoption := store.Adoption()
	gp := store.GoldenPath()
	gpOK := gp.Content != "" && filepath.Base(gp.Path) == "golden-path.md"

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-002",
		"service":        "eng-i-002",
		"run":            run,
		"message":        "live platform-as-product demo",
		"acceptance": []string{
			"SLA objectives published",
			"adoption metrics tracked",
			"golden-path template present",
		},
		"proof": map[string]any{
			"sla":         sla.Availability != "" && sla.LatencyP99MS > 0,
			"adoption":    adoption.Teams >= 1 && adoption.Active >= 1,
			"golden_path": gpOK,
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
