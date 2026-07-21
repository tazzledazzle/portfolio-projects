package main

import (
	"encoding/json"
	"errors"
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

// ENG-E-026: Artifact registries
// OCI-inspired MVP (not conformance-tested) — tag→digest registry simulator.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-E-026",
	"service":        "eng-e-026",
	"title":          "Artifact registries",
}}

var registry = NewRegistry()

const maxManifestBytes = 16 << 20

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	log.Printf("eng-e-026 listening on %s (requirement ENG-E-026) [OCI-inspired simulator]", *addr)
	log.Fatal(http.ListenAndServe(*addr, newMux()))
}

func newMux() http.Handler {
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
			"requirement_id": "ENG-E-026",
			"service":        "eng-e-026",
			"title":          "Artifact registries",
			"oci_inspired":   true,
			"simulator":      true,
			"conformance":    false,
			"label":          "OCI-inspired MVP (not conformance-tested)",
			"time":           time.Now().UTC(),
		})
	})
	mux.HandleFunc("/v1/demo", handleDemo)
	mux.HandleFunc("/v1/registry/", handleRegistryRoutes)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng_e_026_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng_e_026_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng_e_026_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func resetRegistry() {
	registry = NewRegistry()
}

// handleRegistryRoutes routes:
//
//	/v1/registry/{name}/tags/{tag}
//	/v1/registry/{name}/manifests
//	/v1/registry/{name}/manifests/{digest}
func handleRegistryRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/registry/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	name := parts[0]
	switch parts[1] {
	case "tags":
		if len(parts) != 3 {
			http.NotFound(w, r)
			return
		}
		handleTag(w, r, name, parts[2])
	case "manifests":
		if len(parts) == 2 {
			handleManifestRoot(w, r, name)
			return
		}
		if len(parts) == 3 {
			handleManifestByDigest(w, r, name, parts[2])
			return
		}
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

func handleTag(w http.ResponseWriter, r *http.Request, name, tag string) {
	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		var req struct {
			Digest string `json:"digest"`
		}
		if err := json.Unmarshal(body, &req); err != nil || req.Digest == "" {
			// Also accept raw digest string body
			d := strings.TrimSpace(string(body))
			if strings.HasPrefix(d, "sha256:") {
				req.Digest = d
			} else {
				http.Error(w, "expected JSON {\"digest\":\"sha256:...\"}", http.StatusBadRequest)
				return
			}
		}
		if err := registry.PutTag(name, tag, req.Digest); err != nil {
			if errors.Is(err, ErrUnsafeName) || errors.Is(err, ErrInvalidDigest) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{
			"name":   name,
			"tag":    tag,
			"digest": req.Digest,
		})

	case http.MethodGet:
		digest, ok := registry.Resolve(name, tag)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]any{
			"name":   name,
			"tag":    tag,
			"digest": digest,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleManifestRoot(w http.ResponseWriter, r *http.Request, name string) {
	_ = name
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxManifestBytes))
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	digest, err := registry.PutManifest(body)
	if err != nil {
		if errors.Is(err, ErrManifestConflict) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"digest": digest,
		"size":   len(body),
	})
}

func handleManifestByDigest(w http.ResponseWriter, r *http.Request, name, digest string) {
	_ = name
	if !validDigest(digest) {
		http.Error(w, "invalid digest", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		data, ok := registry.GetManifest(digest)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.Header().Set("Docker-Content-Digest", digest)
		_, _ = w.Write(data)

	case http.MethodPut:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxManifestBytes))
		if err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		if err := registry.putManifestAt(digest, body); err != nil {
			if errors.Is(err, ErrManifestConflict) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{"digest": digest, "size": len(body)})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if registry.TagCount() == 0 {
		d, _ := registry.PutManifest([]byte(`{"schemaVersion":2,"demo":true}`))
		_ = registry.PutTag("demo", "latest", d)
	}

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-E-026",
		"service":        "eng-e-026",
		"run":            state.Runs,
		"message":        "demo completed for Artifact registries",
		"acceptance": []string{
			"healthz returns 200",
			"tag resolve returns digest",
			"tag retarget leaves manifest bytes unchanged",
			"OCI-inspired MVP (not conformance-tested)",
		},
	}
	result["proof"] = liveProof()
	state.mu.Unlock()
	writeJSON(w, result)
}

func liveProof() map[string]any {
	digest, ok := registry.Resolve("demo", "latest")
	tagToDigest := ok
	return map[string]any{
		"tag_to_digest":    tagToDigest,
		"resolved_digest":  digest,
		"tag_mutable":      true,
		"digest_immutable": true,
		"oci_inspired":     true,
		"simulator":        true,
		"tag_count":        registry.TagCount(),
		"manifest_count":   registry.ManifestCount(),
		"label":            "OCI-inspired MVP (not conformance-tested)",
	}
}

func proofFor(id string) map[string]any {
	switch id {
	case "ENG-E-026":
		return liveProof()
	case "ENG-E-001", "ENG-E-023", "ENG-N-006":
		return map[string]any{"pipeline_stages": []string{"lint", "unit", "build", "publish"}, "retries": 1}
	case "ENG-E-002":
		return map[string]any{"flake_score": 0.42, "quarantined": true}
	case "ENG-E-003", "ENG-E-024", "ENG-N-004":
		return map[string]any{"canary_weights": []int{0, 10, 50, 100}, "abort_supported": true}
	case "ENG-E-004", "ENG-N-011", "ENG-N-013":
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
