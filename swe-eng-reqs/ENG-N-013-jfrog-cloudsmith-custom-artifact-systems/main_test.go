package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleArtifacts_Unauthorized(t *testing.T) {
	policy = NewPolicyEngine()
	mux := newMux()
	body := map[string]string{"name": "pkg", "digest": digestX}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/v1/artifacts", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want 401/403 body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleRetention_Run(t *testing.T) {
	policy = NewPolicyEngine()
	mux := newMux()
	scopes := []string{"artifacts:write"}
	_ = policy.PutArtifact("a", digestX, scopes)
	_ = policy.PutArtifact("b", digestY, scopes)
	_ = policy.PutArtifact("c", digestZ, scopes)

	body := map[string]int{"keep": 2}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/retention/run", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer demo")
	req.Header.Set("X-Scope", "artifacts:write")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["retention_deleted"] != float64(1) && got["retention_deleted"] != 1 {
		t.Fatalf("retention_deleted=%v, want 1", got["retention_deleted"])
	}
}

func TestHandleScan(t *testing.T) {
	policy = NewPolicyEngine()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/scan", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["scan_hook"] != true {
		t.Fatalf("scan_hook=%v, want true", got["scan_hook"])
	}
	findings, ok := got["findings"].([]any)
	if !ok || len(findings) == 0 {
		t.Fatalf("findings missing: %v", got)
	}
}

func TestHandleInfo_Simulator(t *testing.T) {
	policy = NewPolicyEngine()
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["simulator"] != true {
		t.Fatalf("simulator=%v", got["simulator"])
	}
	if got["vendor_model"] != "custom-registry" {
		t.Fatalf("vendor_model=%v", got["vendor_model"])
	}
}

func TestHandleDemo_SimulatorProof(t *testing.T) {
	policy = NewPolicyEngine()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %v", result)
	}
	if n, ok := proof["retention_deleted"].(float64); !ok || n < 1 {
		t.Fatalf("proof[retention_deleted]=%v, want >= 1", proof["retention_deleted"])
	}
	for _, key := range []string{"auth_scope_enforced", "scan_hook", "simulator"} {
		if proof[key] != true {
			t.Fatalf("proof[%s]=%v, want true", key, proof[key])
		}
	}
}

func TestProofForDefault(t *testing.T) {
	p := proofFor("UNKNOWN")
	if p["vertical_slice"] != true {
		t.Fatalf("expected vertical_slice proof")
	}
}
