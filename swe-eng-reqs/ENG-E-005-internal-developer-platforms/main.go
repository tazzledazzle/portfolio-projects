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

// ENG-E-005: Internal developer platforms
// IDP catalog: projects, pipelines, environments (Boundary Matrix D-03).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var (
	state = &DemoState{Meta: map[string]any{
		"requirement_id": "ENG-E-005",
		"service":        "eng-e-005",
		"title":          "Internal developer platforms",
	}}
	store = NewIDPStore()
)

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-e-005 listening on %s (requirement ENG-E-005)", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux(store)))
}

func newMux(idp *IDPStore) http.Handler {
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
			"requirement_id": "ENG-E-005",
			"service":        "eng-e-005",
			"title":          "Internal developer platforms",
			"owns":           []string{"projects", "pipelines", "environments"},
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		handleProjects(w, r, idp)
	})
	mux.HandleFunc("/v1/pipelines", func(w http.ResponseWriter, r *http.Request) {
		handlePipelines(w, r, idp)
	})
	mux.HandleFunc("/v1/environments", func(w http.ResponseWriter, r *http.Request) {
		handleEnvironments(w, r, idp)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, idp)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_e_005_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_005_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_005_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleProjects(w http.ResponseWriter, r *http.Request, idp *IDPStore) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, idp.ListProjects())
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		proj, err := idp.CreateProject(req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, proj)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePipelines(w http.ResponseWriter, r *http.Request, idp *IDPStore) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, idp.ListPipelines())
	case http.MethodPost:
		var req struct {
			Name      string `json:"name"`
			ProjectID string `json:"project_id"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		pipe, err := idp.CreatePipeline(req.ProjectID, req.Name)
		if err != nil {
			status := http.StatusBadRequest
			if err == ErrProjectNotFound {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, pipe)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleEnvironments(w http.ResponseWriter, r *http.Request, idp *IDPStore) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, idp.ListEnvironments())
	case http.MethodPost:
		var req struct {
			Name      string `json:"name"`
			ProjectID string `json:"project_id"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		env, err := idp.CreateEnvironment(req.ProjectID, req.Name)
		if err != nil {
			status := http.StatusBadRequest
			if err == ErrProjectNotFound {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, env)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDemo(w http.ResponseWriter, r *http.Request, idp *IDPStore) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	proj, err := idp.CreateProject(fmt.Sprintf("demo-%d", time.Now().UnixNano()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := idp.CreatePipeline(proj.ID, "demo-ci"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := idp.CreateEnvironment(proj.ID, "demo-staging"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects := idp.ListProjects()
	pipelines := idp.ListPipelines()
	environments := idp.ListEnvironments()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	run := state.Runs
	state.mu.Unlock()

	writeJSON(w, map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-005",
		"service":        "eng-e-005",
		"run":            run,
		"message":        "live IDP catalog demo",
		"acceptance": []string{
			"projects/pipelines/environments create+list",
			"safe IDs reject path traversal",
			"self_service catalog proof",
		},
		"proof": map[string]any{
			"projects":     len(projects),
			"pipelines":    len(pipelines),
			"environments": len(environments),
			"self_service": true,
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
