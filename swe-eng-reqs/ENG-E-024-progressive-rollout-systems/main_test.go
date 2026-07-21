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

func TestHandleCanary_StepAbortPromote(t *testing.T) {
	canaries = NewCanaryStore()
	mux := newMux()
	body, _ := json.Marshal(map[string]string{"service": "api"})
	req := httptest.NewRequest(http.MethodPost, "/v1/canaries", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var c Canary
	if err := json.Unmarshal(rr.Body.Bytes(), &c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/v1/canaries/"+c.ID+"/step", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("step status=%d body=%s", rr.Code, rr.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/v1/canaries/"+c.ID+"/abort", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("abort status=%d body=%s", rr.Code, rr.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/v1/canaries/"+c.ID, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got Canary
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Status != "aborted" || got.Weight != 0 {
		t.Fatalf("got status=%q weight=%d", got.Status, got.Weight)
	}

	// promote path on a fresh canary
	body, _ = json.Marshal(map[string]string{"service": "api2"})
	req = httptest.NewRequest(http.MethodPost, "/v1/canaries", bytes.NewReader(body))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	var c2 Canary
	_ = json.Unmarshal(rr.Body.Bytes(), &c2)
	req = httptest.NewRequest(http.MethodPost, "/v1/canaries/"+c2.ID+"/promote", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("promote status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleDemo_CanaryProof(t *testing.T) {
	canaries = NewCanaryStore()
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
	if proof["abort_supported"] != true {
		t.Fatalf("abort_supported=%v", proof["abort_supported"])
	}
	weights, ok := proof["canary_weights"].([]any)
	if !ok || len(weights) != 4 {
		t.Fatalf("canary_weights=%v", proof["canary_weights"])
	}
	promoted := proof["promoted"] == true
	aborted := proof["aborted"] == true
	if !promoted && !aborted {
		t.Fatal("proof must show promoted or aborted")
	}
}
