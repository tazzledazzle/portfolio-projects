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

func TestHandleDemo_QueueProof(t *testing.T) {
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
	if proof["idempotent"] != true {
		t.Fatalf("expected proof.idempotent=true, got %#v", proof["idempotent"])
	}
	if proof["duplicate_suppressed"] != true {
		t.Fatalf("expected proof.duplicate_suppressed=true, got %#v", proof["duplicate_suppressed"])
	}
	if proof["partitions"] != true {
		t.Fatalf("expected proof.partitions=true, got %#v", proof["partitions"])
	}
	if proof["retries"] != true {
		t.Fatalf("expected proof.retries=true, got %#v", proof["retries"])
	}
}

func TestHandleTasks_Enqueue(t *testing.T) {
	mux := newMux()
	body := `{"payload":"hello","idempotency_key":"http-key-1","partition":3}`
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	task, ok := resp["task"].(map[string]any)
	if !ok {
		t.Fatalf("missing task: %#v", resp)
	}
	if task["partition"].(float64) != 3 {
		t.Fatalf("expected partition 3, got %#v", task["partition"])
	}
}
