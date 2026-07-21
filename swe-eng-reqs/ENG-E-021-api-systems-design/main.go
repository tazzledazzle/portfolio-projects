package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ENG-E-021: API systems design
// Owns OpenAPI + authz + rate limits + compat (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-021",
	"service":        "eng-e-021",
	"title":          "API systems design",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	docDir := getenv("OPENAPI_DIR", ".")
	eng := NewAPIEngine(5, "resources:read")
	eng.SetDocPath(filepath.Join(docDir, "openapi.yaml"))

	log.Printf("eng-e-021 listening on %s (requirement ENG-E-021)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(eng, docDir)))
}

func newMux(eng *APIEngine, docDir string) http.Handler {
	eng.SetDocPath(filepath.Join(docDir, "openapi.yaml"))
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
			"requirement_id":   "ENG-E-021",
			"service":          "eng-e-021",
			"title":            "API systems design",
			"openapi":          true,
			"openapi_inspired": true,
			"note":             "Hand-authored OpenAPI 3.x; no OAS codegen or external gateway",
			"owns":             []string{"openapi", "authz", "rate_limits", "compat"},
			"time":             time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		doc, err := eng.OpenAPIDoc()
		if err != nil {
			http.Error(w, "openapi unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte(doc))
	})
	mux.HandleFunc("/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		handleGuardedResource(w, r, eng, true)
	})
	mux.HandleFunc("/v2/resources", func(w http.ResponseWriter, r *http.Request) {
		handleGuardedResource(w, r, eng, false)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, eng)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_e_021_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_021_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_021_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func scopesFromRequest(r *http.Request) []string {
	raw := r.Header.Get("X-Scope")
	if raw == "" {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			raw = strings.TrimSpace(auth[7:])
		}
	}
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func handleGuardedResource(w http.ResponseWriter, r *http.Request, eng *APIEngine, v1 bool) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	scopes := scopesFromRequest(r)
	ok, err := eng.Authorize(scopes)
	if !ok {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	subject := r.Header.Get("X-Subject")
	if subject == "" {
		subject = "anonymous"
	}
	if !eng.Allow(subject) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		http.Error(w, "rate_limited", http.StatusTooManyRequests)
		return
	}
	if v1 {
		writeJSON(w, eng.ResourceV1())
		return
	}
	writeJSON(w, eng.ResourceV2())
}

func handleDemo(w http.ResponseWriter, r *http.Request, eng *APIEngine) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	okNil, _ := eng.Authorize(nil)
	okScope, authErr := eng.Authorize([]string{"resources:read"})
	authzEnforced := !okNil && okScope && authErr == nil

	// Prove rate limiting with a dedicated engine (server-side counters).
	rl := NewAPIEngine(2, "resources:read")
	_ = rl.Allow("demo-user")
	_ = rl.Allow("demo-user")
	rateLimited := !rl.Allow("demo-user")

	doc, err := eng.OpenAPIDoc()
	openapiOK := err == nil && strings.Contains(doc, "openapi:")
	compat := eng.Compat()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-021",
		"service":        "eng-e-021",
		"run":            run,
		"message":        "live OpenAPI authz rate-limit compat demo",
		"acceptance": []string{
			"OpenAPI document present",
			"authz default-deny",
			"server-side rate limits",
			"v1/v2 compat shapes",
		},
		"proof": map[string]any{
			"openapi":          openapiOK,
			"openapi_inspired": true,
			"authz_enforced":   authzEnforced,
			"rate_limited":     rateLimited,
			"compat_pass":      compat.Pass,
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
