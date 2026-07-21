package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ENG-I-008: Supply-chain awareness for build/artifacts
// ed25519 sign + SPDX-inspired SBOM + registry scopes; sigstore=false.

type DemoState struct {
	mu       sync.Mutex
	Runs     int            `json:"runs"`
	LastDemo string         `json:"last_demo"`
	Meta     map[string]any `json:"meta"`
}

var state = &DemoState{Meta: map[string]any{
	"requirement_id": "ENG-I-008",
	"service":        "eng-i-008",
	"title":          "Supply-chain awareness for build/artifacts",
}}

func main() {
	addr := flag.String("addr", getenv("ADDR", ":8080"), "listen address")
	flag.Parse()

	sc, err := loadOrGenerateSupplyChain()
	if err != nil {
		log.Fatal(err)
	}
	mux := newMux(sc)
	log.Printf("eng-i-008 listening on %s (requirement ENG-I-008)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func loadOrGenerateSupplyChain() (*SupplyChain, error) {
	privPath := filepath.Join("testdata", "keys", "ed25519.priv")
	pubPath := filepath.Join("testdata", "keys", "ed25519.pub")
	if _, err := os.Stat(privPath); err == nil {
		return LoadSupplyChainFromFiles(privPath, pubPath)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewSupplyChain(priv, pub), nil
}

func newMux(sc *SupplyChain) http.Handler {
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
		info := sc.Info()
		info["time"] = time.Now().UTC()
		writeJSON(w, info)
	})
	mux.HandleFunc("/v1/sign", func(w http.ResponseWriter, r *http.Request) {
		handleSign(w, r, sc)
	})
	mux.HandleFunc("/v1/sbom", func(w http.ResponseWriter, r *http.Request) {
		handleSBOM(w, r, sc)
	})
	mux.HandleFunc("/v1/push", func(w http.ResponseWriter, r *http.Request) {
		handlePush(w, r, sc)
	})
	mux.HandleFunc("/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		handleDemo(w, r, sc)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()
		fmt.Fprintf(w, "# HELP eng-i-008_demo_runs_total Demo invocations\n")
		fmt.Fprintf(w, "# TYPE eng-i-008_demo_runs_total counter\n")
		fmt.Fprintf(w, "eng-i-008_demo_runs_total %d\n", state.Runs)
	})
	return mux
}

func handleSign(w http.ResponseWriter, r *http.Request, sc *SupplyChain) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Digest string `json:"digest"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	sig, err := sc.Sign(request.Digest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"digest": request.Digest, "signature": sig, "signing": "ed25519"})
}

func handleSBOM(w http.ResponseWriter, r *http.Request, sc *SupplyChain) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
	}
	if r.Method == http.MethodPost {
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
	if request.Name == "" {
		request.Name = "demo-artifact"
	}
	if request.Digest == "" {
		sum := sha256.Sum256([]byte(request.Name))
		request.Digest = "sha256:" + hex.EncodeToString(sum[:])
	}
	writeJSON(w, sc.SBOM(request.Name, request.Digest))
}

func handlePush(w http.ResponseWriter, r *http.Request, sc *SupplyChain) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Name      string   `json:"name"`
		Digest    string   `json:"digest"`
		Signature string   `json:"signature"`
		Scopes    []string `json:"scopes"`
	}
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := sc.Push(request.Name, request.Digest, request.Signature, request.Scopes); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "name": request.Name, "digest": request.Digest})
}

func handleDemo(w http.ResponseWriter, r *http.Request, sc *SupplyChain) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sum := sha256.Sum256([]byte("eng-i-008-demo"))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	sig, err := sc.Sign(digest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ok, err := sc.Verify(digest, sig)
	if err != nil || !ok {
		http.Error(w, "verify failed", http.StatusInternalServerError)
		return
	}
	sbom := sc.SBOM("eng-i-008-demo", digest)
	denied, _ := sc.AuthorizePush([]string{"artifacts:read"})
	allowed, _ := sc.AuthorizePush([]string{"artifacts:push"})
	if err := sc.Push("eng-i-008-demo", digest, sig, []string{"artifacts:push"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	info := sc.Info()

	state.mu.Lock()
	state.Runs++
	state.LastDemo = time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"ok":             true,
		"requirement_id": "ENG-I-008",
		"service":        "eng-i-008",
		"run":            state.Runs,
		"message":        "demo completed for ed25519 sign + SPDX-inspired SBOM + scopes",
		"acceptance": []string{
			"healthz returns 200",
			"ed25519 sign/verify round-trip",
			"SPDX-inspired SBOM labeled",
			"push scope default-deny",
			"sigstore=false",
		},
	}
	result["proof"] = map[string]any{
		"signed":             ok,
		"sbom_spdx_inspired": sbom["spdx_inspired"] == true,
		"scope_enforced":     !denied && allowed,
		"sigstore":           false,
		"signing":            "ed25519",
		"spdx_inspired":      true,
		"digest":             digest,
		"simulator":          info["simulator"] == true,
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
