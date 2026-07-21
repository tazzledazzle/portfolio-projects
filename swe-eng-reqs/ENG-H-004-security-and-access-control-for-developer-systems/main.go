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

// ENG-H-004: Security and access control for developer systems
// OIDC-inspired claims + RBAC; no external IdP.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-H-004",
	"service":        "eng-h-004",
	"title":          "Security and access control for developer systems",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux(NewAuthzEngine())
	log.Printf("eng-h-004 listening on %s (requirement ENG-H-004)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newMux(eng *AuthzEngine) http.Handler {
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
		info := eng.Info()
		info["time"] = time.Now().UTC()
		writeJSON(w, info)
	})
	mux.HandleFunc("/v1/auth/evaluate", func(w http.ResponseWriter, r *http.Request) {
		handleEvaluate(w, r, eng)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, eng)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-h-004_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-h-004_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-h-004_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleEvaluate(w http.ResponseWriter, r *http.Request, eng *AuthzEngine) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Claims   Claims `json:"claims"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	decision := eng.Evaluate(request.Claims, request.Resource, request.Action)
	if !decision.Allow {
		w.WriteHeader(http.StatusForbidden)
	}
	writeJSON(w, decision)
}

func handleDemo(w http.ResponseWriter, r *http.Request, eng *AuthzEngine) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	now := time.Now().UTC()
	allowClaims := Claims{
		Iss: "https://idp.example.invalid", Sub: "alice", Aud: "eng-h-004",
		Exp: now.Add(time.Hour).Unix(), Roles: []string{"developer"},
	}
	denyClaims := Claims{
		Iss: "https://idp.example.invalid", Sub: "bob", Aud: "eng-h-004",
		Exp: now.Add(time.Hour).Unix(), Roles: []string{"viewer"},
	}
	expiredClaims := Claims{
		Iss: "https://idp.example.invalid", Sub: "carol", Aud: "eng-h-004",
		Exp: now.Add(-time.Minute).Unix(), Roles: []string{"admin"},
	}

	allow := eng.Evaluate(allowClaims, "pipelines", "write")
	deny := eng.Evaluate(denyClaims, "pipelines", "write")
	expired := eng.Evaluate(expiredClaims, "pipelines", "write")
	info := eng.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-H-004",
		"service":        "eng-h-004",
		"run":            state.Runs,
		"message":        "demo completed for OIDC-inspired RBAC (no external IdP)",
		"acceptance": []string{
			"healthz returns 200",
			"role allow produces rbac_allow",
			"missing role produces rbac_deny",
			"expired exp denies",
			"oidc_inspired simulator labeled",
		},
	}
	result["proof"] = map[string]any{
		"oidc_inspired": info["oidc_inspired"] == true,
		"simulator":     info["simulator"] == true,
		"external_idp":  false,
		"rbac_allow":    allow.Allow && allow.Reason == "rbac_allow",
		"rbac_deny":     !deny.Allow && deny.Reason == "rbac_deny",
		"exp_denied":    !expired.Allow,
		"allow_decision": allow,
		"deny_decision":  deny,
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
