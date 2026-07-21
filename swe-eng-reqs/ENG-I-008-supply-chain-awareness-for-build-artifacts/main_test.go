package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mustTestSupplyChain(t *testing.T) *SupplyChain {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return NewSupplyChain(priv, pub)
}

func TestHealthz(t *testing.T) {
	mux := newMux(mustTestSupplyChain(t))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_SupplyChainProof(t *testing.T) {
	mux := newMux(mustTestSupplyChain(t))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, key := range []string{`"signed": true`, `"sbom_spdx_inspired": true`, `"scope_enforced": true`, `"sigstore": false`} {
		if !strings.Contains(body, key) {
			t.Fatalf("demo missing %s: %s", key, body)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	proof, _ := payload["proof"].(map[string]any)
	if proof["signing"] != "ed25519" {
		t.Fatalf("proof.signing=%v", proof["signing"])
	}
}
