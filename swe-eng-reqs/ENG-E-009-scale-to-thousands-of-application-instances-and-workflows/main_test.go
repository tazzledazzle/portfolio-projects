package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestHandleDemo_ScaleProof(t *testing.T) {
	mux := newMux()
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
	proof, ok := body["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %#v", body)
	}
	wf, ok := proof["workflows_simulated"].(float64)
	if !ok || wf < 1000 {
		t.Fatalf("expected workflows_simulated>=1000, got %#v", proof["workflows_simulated"])
	}
	if proof["backpressure"] != true {
		t.Fatalf("expected backpressure=true, got %#v", proof["backpressure"])
	}
	if _, ok := proof["p99_ms"]; !ok {
		t.Fatalf("expected p99_ms, got %#v", proof)
	}
	if _, ok := proof["queue_depth"]; !ok {
		t.Fatalf("expected queue_depth, got %#v", proof)
	}
}

func TestHandleSimulate(t *testing.T) {
	mux := newMux()
	body := `{"count":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/simulate", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok, got %#v", resp)
	}
}
