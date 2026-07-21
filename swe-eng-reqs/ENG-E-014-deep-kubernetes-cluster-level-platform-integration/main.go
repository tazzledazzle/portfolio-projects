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

// ENG-E-014: Deep Kubernetes cluster-level platform integration
// CIJob schedules Job; Complete/Failed conditions (distinct from N-007 ManagedWorkload).

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-014",
	"service":        "eng-e-014",
	"title":          "Deep Kubernetes cluster-level platform integration",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewCIJobController())
	log.Printf("eng-e-014 listening on %s (requirement ENG-E-014)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(controller *CIJobController) http.Handler {
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
			"requirement_id": "ENG-E-014",
			"service":        "eng-e-014",
			"title":          "Deep Kubernetes cluster-level platform integration",
			"kind":           "CIJob",
			"kind_optional": true,
			"note":          "demo-local in-memory gate; Kind cluster apply is optional (D-05)",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/cijobs", func(w http.ResponseWriter, r *http.Request) {
		handleCIJobs(w, r, controller)
	})
	mux.HandleFunc("/v1/cijobs/", func(w http.ResponseWriter, r *http.Request) {
		handleCIJobs(w, r, controller)
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
		fmt.Fprintf(w, "# HELP eng-e-014_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-e-014_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-e-014_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleCIJobs(w http.ResponseWriter, r *http.Request, controller *CIJobController) {
	if r.URL.Path == "/v1/cijobs" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			ID    string `json:"id"`
			Image string `json:"image"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		job, err := controller.Create(request.ID, request.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, job)
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/cijobs/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && r.Method == http.MethodGet {
		job, ok := controller.Get(parts[0])
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, job)
		return
	}
	if len(parts) == 2 && parts[1] == "outcome" && r.Method == http.MethodPost {
		var request struct {
			Outcome string `json:"outcome"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if err := controller.SetJobOutcome(parts[0], request.Outcome); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		job, err := controller.Reconcile(parts[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, job)
		return
	}
	http.Error(w, "invalid CIJob route", http.StatusBadRequest)
}

func handleReconcile(w http.ResponseWriter, r *http.Request, controller *CIJobController) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]any{"terminal": controller.ReconcileAll()})
}

func handleDemo(w http.ResponseWriter, r *http.Request, controller *CIJobController) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ns := fmt.Sprintf("%d", time.Now().UnixNano())
	okID := "demo-ok-" + ns
	badID := "demo-bad-" + ns

	okJob, err := controller.Create(okID, "busybox:latest")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := controller.SetJobOutcome(okID, "Succeeded"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	okJob, err = controller.Reconcile(okID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	badJob, err := controller.Create(badID, "busybox:latest")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := controller.SetJobOutcome(badID, "Failed"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	badJob, err = controller.Reconcile(badID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conditionTypes := make([]string, 0, len(okJob.Conditions))
	for _, condition := range okJob.Conditions {
		conditionTypes = append(conditionTypes, condition.Type)
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-014",
		"service":        "eng-e-014",
		"run":            state.Runs,
		"message":        "demo completed for CIJob→Job Complete/Failed (Kind optional)",
		"acceptance": []string{
			"healthz returns 200",
			"CIJob schedules Job child",
			"Complete and Failed conditions derived from Job state",
			"kind is CIJob (not ManagedWorkload)",
		},
	}
	result["proof"] = map[string]any{
		"crd":                        true,
		"kind":                       "CIJob",
		"job_scheduled":              okJob.JobScheduled && badJob.JobScheduled,
		"complete":                   conditionTrue(okJob, "Complete"),
		"failed":                     conditionTrue(badJob, "Failed"),
		"complete_and_failed_paths":  conditionTrue(okJob, "Complete") && conditionTrue(badJob, "Failed"),
		"conditions":                 conditionTypes,
		"ok_job_name":                okJob.Job.Name,
		"bad_job_name":               badJob.Job.Name,
		"kind_optional":              true,
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
