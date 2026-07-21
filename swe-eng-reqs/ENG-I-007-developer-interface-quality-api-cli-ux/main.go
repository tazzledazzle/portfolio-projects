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

// ENG-I-007: Developer interface quality (API/CLI UX)
// Owns CLI --json, clear errors, golden snapshots — not OpenAPI/rate-limit (E-021).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-I-007",
	"service":        "eng-i-007",
	"title":          "Developer interface quality (API/CLI UX)",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	// CLI mode when non-flag args remain after flag.Parse (e.g. status --json).
	if args := flag.Args(); len(args) > 0 {
		stdout, err := RunCLI(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, stdout)
		return
	}

	mux := newMux()
	log.Printf("eng-i-007 listening on %s (requirement ENG-I-007)", *addr)
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
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-i-007_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-i-007_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-i-007_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"requirement_id": "ENG-I-007",
		"service":        "eng-i-007",
		"title":          "Developer interface quality (API/CLI UX)",
		"cli":            true,
		"owns":           []string{"cli_json", "clear_errors", "golden_snapshots"},
		"does_not_own":   []string{"openapi", "rate_limit", "idp_catalog"},
		"time":           time.Now().UTC(),
	})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stdout, err := RunCLI([]string{"status", "--json"})
	cliJSON := err == nil && json.Valid([]byte(stdout))

	_, unknownErr := RunCLI([]string{"not-a-command"})
	clearErrors := unknownErr != nil && unknownErr.Error() != ""

	goldenMatch := false
	if cliJSON {
		want, readErr := os.ReadFile("testdata/status.json.golden")
		if readErr == nil {
			got := normalizeJSONWhitespace(stdout)
			expected := normalizeJSONWhitespace(string(want))
			goldenMatch = got == expected
		}
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-007",
		"service":        "eng-i-007",
		"run":            state.Runs,
		"message":        "demo completed for Developer interface quality (API/CLI UX)",
		"acceptance": []string{
			"CLI status --json emits valid JSON",
			"unknown command yields clear error without panic",
			"stdout matches testdata/status.json.golden",
		},
		"proof": map[string]any{
			"cli_json":     cliJSON,
			"clear_errors": clearErrors,
			"golden_match": goldenMatch,
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
