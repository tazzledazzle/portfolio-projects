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

// ENG-N-007: Custom Kubernetes operators
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-007",
	"service":        "eng-n-007",
	"title":          "Custom Kubernetes operators",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewController())
	log.Printf("eng-n-007 listening on %s (requirement ENG-N-007)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(controller *Controller) http.Handler {
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
			"requirement_id": "ENG-N-007",
			"service":        "eng-n-007",
			"title":          "Custom Kubernetes operators",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		handleWorkloads(w, r, controller)
	})
	mux.HandleFunc("/v1/workloads/", func(w http.ResponseWriter, r *http.Request) {
		handleWorkloads(w, r, controller)
	})
	mux.HandleFunc("/v1/reconcile", func(w http.ResponseWriter, r *http.Request) {
		handleReconcile(w, r, controller)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, controller)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		_, _ = fmt.Fprintf(w, "# HELP eng_n_007_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_n_007_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_n_007_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleWorkloads(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.URL.Path == "/v1/workloads" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			ID       string `json:"id"`
			Replicas int    `json:"replicas"`
			Image    string `json:"image"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		workload, err := controller.Create(request.ID, request.Replicas, request.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, workload)
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/workloads/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && r.Method == http.MethodDelete {
		workload, err := controller.Delete(parts[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, workload)
		return
	}
	if len(parts) == 2 && parts[1] == "finalize" && r.Method == http.MethodPost {
		workload, err := controller.Finalize(parts[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, present := controller.Get(parts[0])
		writeJSON(w, map[string]any{
			"workload":          workload,
			"finalizer_cleared": len(workload.Finalizers) == 0 && !present,
		})
		return
	}
	http.Error(w, "invalid ManagedWorkload route", http.StatusBadRequest)
}

func handleReconcile(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		ID string `json:"id"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	workload, err := controller.Reconcile(request.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, workload)
}

func handleDemo(w http.ResponseWriter, r *http.Request, controller *Controller) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := fmt.Sprintf("demo-%d", time.Now().UTC().UnixNano())
	if _, err := controller.Create(id, 2, "example/managed-workload:v1"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reconciled, err := controller.Reconcile(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := controller.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	finalized, err := controller.Finalize(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, present := controller.Get(id)
	conditionTypes := make([]string, 0, len(reconciled.Conditions))
	for _, condition := range reconciled.Conditions {
		conditionTypes = append(conditionTypes, condition.Type)
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-007",
		"service":        "eng-n-007",
		"run":            state.Runs,
		"message":        "demo completed for Custom Kubernetes operators",
		"acceptance": []string{
			"healthz returns 200",
			"demo increments run counter",
			"metrics expose demo_runs_total",
		},
	}
	result["proof"] = map[string]any{
		"crd":               true,
		"kind":              "ManagedWorkload",
		"reconciled":        conditionTrue(reconciled, "Ready"),
		"conditions":        conditionTypes,
		"finalizer_cleared": len(finalized.Finalizers) == 0 && !present,
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
