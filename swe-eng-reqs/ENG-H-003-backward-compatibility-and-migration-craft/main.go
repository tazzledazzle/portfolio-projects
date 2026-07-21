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

// ENG-H-003: Backward compatibility and migration craft
// Owns API v1→v2 dual-write + compat tests (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-H-003",
		"service":        "eng-h-003",
		"title":          "Backward compatibility and migration craft",
	}}
	store = NewMigrateStore()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-h-003 listening on %s (requirement ENG-H-003)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(store)))
}

func newMux(s *MigrateStore) http.Handler {
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
			"requirement_id": "ENG-H-003",
			"service":        "eng-h-003",
			"title":          "Backward compatibility and migration craft",
			"owns":           []string{"dual_write_migration", "compat_tests"},
			"does_not_own":  []string{"openapi_craft", "rate_limits", "idp_catalog"},
			"dual_write":    s.DualWriteEnabled(),
			"time":          time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/migrate/dual-write", func(w http.ResponseWriter, r *http.Request) {
		handleEnableDualWrite(w, r, s)
	})
	mux.HandleFunc("/v1/items", func(w http.ResponseWriter, r *http.Request) {
		handleItemsV1(w, r, s)
	})
	mux.HandleFunc("/v2/items", func(w http.ResponseWriter, r *http.Request) {
		handleItemsV2(w, r, s)
	})
	mux.HandleFunc("/v1/items/", func(w http.ResponseWriter, r *http.Request) {
		handleGetItemV1(w, r, s)
	})
	mux.HandleFunc("/v2/items/", func(w http.ResponseWriter, r *http.Request) {
		handleGetItemV2(w, r, s)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, s)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_h_003_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_h_003_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_h_003_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleEnableDualWrite(w http.ResponseWriter, r *http.Request, s *MigrateStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	s.EnableDualWrite(req.Enabled)
	writeJSON(w, map[string]any{"dual_write": s.DualWriteEnabled()})
}

func handleItemsV1(w http.ResponseWriter, r *http.Request, s *MigrateStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var fields map[string]any
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&fields); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	id, err := s.Put(fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, _ := s.GetV1(id)
	writeJSON(w, item)
}

func handleItemsV2(w http.ResponseWriter, r *http.Request, s *MigrateStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var fields map[string]any
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&fields); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	id, err := s.Put(fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, ok := s.GetV2(id)
	if !ok {
		http.Error(w, ErrDualWriteOff.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, item)
}

func handleGetItemV1(w http.ResponseWriter, r *http.Request, s *MigrateStore) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/items/")
	if !safeID(id) {
		http.Error(w, ErrUnsafeID.Error(), http.StatusBadRequest)
		return
	}
	item, ok := s.GetV1(id)
	if !ok {
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, item)
}

func handleGetItemV2(w http.ResponseWriter, r *http.Request, s *MigrateStore) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v2/items/")
	if !safeID(id) {
		http.Error(w, ErrUnsafeID.Error(), http.StatusBadRequest)
		return
	}
	item, ok := s.GetV2(id)
	if !ok {
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, item)
}

func handleDemo(w http.ResponseWriter, r *http.Request, _ *MigrateStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewMigrateStore()
	report, err := demo.SeedDemo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-H-003",
		"service":        "eng-h-003",
		"run":            run,
		"message":        "live v1→v2 dual-write migration compat demo",
		"acceptance": []string{
			"dual-write keeps v1 and v2 readable",
			"field rename name→display_name under single mutex",
			"compat_pass when both versions present",
		},
		"proof": map[string]any{
			"dual_write":   report.DualWrite,
			"v1_readable":  report.V1Readable,
			"v2_readable":  report.V2Readable,
			"compat_pass":  report.CompatPass,
			"field_rename": report.FieldRename,
			"item_id":      report.ItemID,
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
