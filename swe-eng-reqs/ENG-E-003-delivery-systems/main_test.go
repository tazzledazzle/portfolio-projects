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

func TestHandleRelease_PromoteAudit(t *testing.T) {
	store = NewDeliveryStore()
	mux := newMux()
	body, _ := json.Marshal(map[string]string{
		"app":     "demo-app",
		"version": "1.0.0",
		"stage":   "dev",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/releases", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var rel Release
	if err := json.Unmarshal(rr.Body.Bytes(), &rel); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/v1/releases/"+rel.ID+"/promote", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("promote status=%d body=%s", rr.Code, rr.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/v1/releases/"+rel.ID+"/audit", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("audit status=%d body=%s", rr.Code, rr.Body.String())
	}
	var entries []AuditEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	if len(entries) < 1 {
		t.Fatalf("audit entries=%d, want ≥1", len(entries))
	}
}

func TestHandleRollback(t *testing.T) {
	store = NewDeliveryStore()
	mux := newMux()
	rel, err := store.Create("rb-app", "1.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Promote(rel.ID); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/releases/"+rel.ID+"/rollback", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("rollback status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	store = NewDeliveryStore()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo status=%d body=%s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %v", result)
	}
	if proof["promoted"] != true {
		t.Fatalf("proof.promoted=%v, want true", proof["promoted"])
	}
	entries, ok := proof["audit_entries"].(float64)
	if !ok || entries < 2 {
		t.Fatalf("proof.audit_entries=%v, want ≥2", proof["audit_entries"])
	}
	if proof["rollback_invoked"] != true {
		t.Fatalf("proof.rollback_invoked=%v, want true", proof["rollback_invoked"])
	}
	if _, has := proof["canary_weights"]; has {
		t.Fatal("demo must not use static scaffold canary_weights proof")
	}
}
