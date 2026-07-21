package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewSelfServiceStore(3))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleSelfService_RequestAndMetrics(t *testing.T) {
	mux := newMux(NewSelfServiceStore(4))
	body, _ := json.Marshal(map[string]string{"kind": "env", "summary": "staging"})
	req := httptest.NewRequest(http.MethodPost, "/v1/requests", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /v1/requests: %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /v1/metrics: %d", rr.Code)
	}
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["ticket_removed_path"] != true {
		t.Fatalf("expected ticket_removed_path: %#v", m)
	}
}

func TestHandleDemo_SelfServiceProof(t *testing.T) {
	mux := newMux(NewSelfServiceStore(8))
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	proof, _ := result["proof"].(map[string]any)
	if proof["ticket_removed"] != true || proof["self_service"] != true {
		t.Fatalf("proof incomplete: %#v", proof)
	}
	if proof["automation_metrics"] == nil {
		t.Fatalf("missing automation_metrics: %#v", proof)
	}
}
