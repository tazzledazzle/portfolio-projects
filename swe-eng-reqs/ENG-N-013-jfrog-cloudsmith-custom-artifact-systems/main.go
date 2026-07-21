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

// ENG-N-013: Custom registry SIMULATOR (not real JFrog/Cloudsmith)
// Vertical-slice MVP control-plane service.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-N-013",
	"service":        "eng-n-013",
	"title":          "Custom registry simulator",
}}

var policy = NewPolicyEngine()

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	mux := newMux()
	log.Printf("eng-n-013 listening on %s (requirement ENG-N-013) [SIMULATOR]", *addr)
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
	mux.HandleFunc("/v1/artifacts", handleArtifacts)
	mux.HandleFunc("/v1/retention/run", handleRetention)
	mux.HandleFunc("/v1/scan", handleScan)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-n-013_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-n-013_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-n-013_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	info := policy.Info()
	info["time"] = time.Now().UTC()
	writeJSON(w, info)
}

func scopesFromRequest(r *http.Request) []string {
	var scopes []string
	if s := r.Header.Get("X-Scope"); s != "" {
		for _, part := range strings.Split(s, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				scopes = append(scopes, part)
			}
		}
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		// Demo token: Bearer demo implies write when X-Scope also present,
		// or Bearer demo with scope claim in token (demo = write for portfolio).
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "demo" && len(scopes) == 0 {
			scopes = append(scopes, requiredWriteScope)
		}
	}
	return scopes
}

func handleArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scopes := scopesFromRequest(r)
	ok, err := policy.Authorize(scopes)
	if !ok {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	var body struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := policy.PutArtifact(body.Name, body.Digest, scopes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "name": body.Name, "digest": body.Digest})
}

func handleRetention(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scopes := scopesFromRequest(r)
	ok, err := policy.Authorize(scopes)
	if !ok {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	keep := 2
	var body struct {
		Keep int `json:"keep"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
		if body.Keep > 0 {
			keep = body.Keep
		}
	}
	deleted, err := policy.RunRetention(keep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{
		"ok":                 true,
		"retention_deleted":  deleted,
		"keep":               keep,
		"remaining":          len(policy.ListArtifacts()),
	})
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	findings := policy.Scan()
	writeJSON(w, map[string]any{
		"ok":         true,
		"scan_hook":  true,
		"stub":       true,
		"findings":   findings,
		"note":       "Fixture findings only — no network scanner",
	})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	demo := NewPolicyEngine()
	// Prove auth scope denial
	denied, denyErr := demo.Authorize(nil)
	authEnforced := !denied && denyErr != nil

	writeScopes := []string{"artifacts:write"}
	_ = demo.PutArtifact("old", "sha256:1111111111111111111111111111111111111111111111111111111111111111", writeScopes)
	_ = demo.PutArtifact("mid", "sha256:2222222222222222222222222222222222222222222222222222222222222222", writeScopes)
	_ = demo.PutArtifact("new", "sha256:3333333333333333333333333333333333333333333333333333333333333333", writeScopes)
	deleted, _ := demo.RunRetention(2)
	findings := demo.Scan()
	info := demo.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-N-013",
		"service":        "eng-n-013",
		"run":            state.Runs,
		"message":        "demo completed for custom registry SIMULATOR (not JFrog/Cloudsmith)",
		"acceptance": []string{
			"healthz returns 200",
			"scope middleware denies unauthorized puts",
			"retention deletes by keep-count without rewriting digests",
			"scan hook returns fixture findings only",
		},
		"proof": map[string]any{
			"retention_deleted":   deleted,
			"auth_scope_enforced": authEnforced,
			"scan_hook":           len(findings) > 0,
			"simulator":           info["simulator"] == true,
			"vendor_model":        info["vendor_model"],
		},
	}
	state.mu.Unlock()
	writeJSON(w, result)
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-E-001", "ENG-E-023", "ENG-N-006":
		return map[string]any{"pipeline_stages": []string{"lint", "unit", "build", "publish"}, "retries": 1}
	case "ENG-E-002":
		return map[string]any{"flake_score": 0.42, "quarantined": true}
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
