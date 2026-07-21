package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewTenantScheduler())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	mux := newMux(NewTenantScheduler())
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, _ := body["proof"].(map[string]any)
	if proof["quota_enforced"] != true {
		t.Fatalf("expected quota_enforced, got %#v", proof)
	}
	if proof["noisy_neighbor_limited"] != true {
		t.Fatalf("expected noisy_neighbor_limited, got %#v", proof)
	}
}

func TestHandleSchedule_RequiresTenant(t *testing.T) {
	mux := newMux(NewTenantScheduler())
	req := httptest.NewRequest(http.MethodPost, "/v1/schedule", strings.NewReader(`{"units":1}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code == 200 {
		t.Fatal("expected rejection without tenant_id")
	}
}
