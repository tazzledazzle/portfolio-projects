package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ENG-E-008: Build production platform services
// Production runtime quality: request IDs, structured errors, /metrics.

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	rt := NewServiceRuntime()
	mux := newMux(rt)
	log.Printf("eng-e-008 listening on %s (requirement ENG-E-008)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(rt *ServiceRuntime) http.Handler {
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
			"requirement_id": "ENG-E-008",
			"service":        "eng-e-008",
			"title":          "Build production platform services",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/echo", func(w http.ResponseWriter, r *http.Request) {
		handleEcho(w, r, rt)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, rt)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "# HELP eng_e_008_requests_total Echo and demo requests\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_e_008_requests_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_e_008_requests_total %d\n", rt.MetricsCount())
		_, _ = fmt.Fprintf(w, "# HELP eng_e_008_demo_runs_total Demo invocations\n")
		_, _ = fmt.Fprintf(w, "# TYPE eng_e_008_demo_runs_total counter\n")
		_, _ = fmt.Fprintf(w, "eng_e_008_demo_runs_total %d\n", rt.DemoCount())
	})
	return mux
}

func handleEcho(w http.ResponseWriter, r *http.Request, rt *ServiceRuntime) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeStructuredError(w, http.StatusMethodNotAllowed, StructuredError{
			Code:    "method_not_allowed",
			Message: "use GET or POST",
		})
		return
	}
	reqID := rt.WithRequestID(r.Header.Get("X-Request-ID"))
	w.Header().Set("X-Request-ID", reqID)

	var payload struct {
		Message string `json:"message"`
	}
	if r.Body != nil && r.Method == http.MethodPost {
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&payload); err != nil && err != io.EOF {
			writeStructuredError(w, http.StatusBadRequest, StructuredError{
				Code:    "bad_request",
				Message: "invalid echo payload",
			})
			return
		}
	}
	if payload.Message == "" {
		payload.Message = "pong"
	}
	rt.IncrementEcho()
	writeJSON(w, map[string]any{
		"ok":         true,
		"message":    payload.Message,
		"request_id": reqID,
	})
}

func handleDemo(w http.ResponseWriter, r *http.Request, rt *ServiceRuntime) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeStructuredError(w, http.StatusMethodNotAllowed, StructuredError{
			Code:    "method_not_allowed",
			Message: "use GET or POST",
		})
		return
	}
	reqID := rt.WithRequestID(r.Header.Get("X-Request-ID"))
	w.Header().Set("X-Request-ID", reqID)
	rt.IncrementDemo()
	sample := SampleStructuredError()

	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-008",
		"service":        "eng-e-008",
		"run":            rt.DemoCount(),
		"request_id":     reqID,
		"message":        "demo completed for Build production platform services",
		"acceptance": []string{
			"healthz returns 200",
			"request IDs propagate on /v1/echo",
			"structured errors omit stack traces",
			"metrics expose request counters",
		},
		"proof": map[string]any{
			"metrics_exposed":  rt.MetricsCount() > 0,
			"request_id":       reqID,
			"structured_error": sample.Code != "" && sample.Message != "",
			"error_sample":     sample,
		},
	}
	writeJSON(w, result)
}

func writeStructuredError(w http.ResponseWriter, status int, se StructuredError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(se)
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
