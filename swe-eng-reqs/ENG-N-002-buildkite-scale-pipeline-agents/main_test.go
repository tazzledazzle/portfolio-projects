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

func TestHandleAgents_POST_Register(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewReader([]byte(`{"id":"agent-test-1"}`)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleAgents_GET_Poll(t *testing.T) {
	mux := newMux()
	_ = agentPool.Register("poll-agent", nil)
	agentPool.Enqueue(&BKJob{ID: "poll-job-1", Status: "waiting", Group: "default"})
	req := httptest.NewRequest(http.MethodGet, "/v1/agents/poll-agent/poll", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandlePipelines_POST_Upload(t *testing.T) {
	mux := newMux()
	yaml := "steps:\n  - label: lint\n  - label: test\n"
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines/p1/upload", bytes.NewReader([]byte(yaml)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleConcurrency_GET(t *testing.T) {
	mux := newMux()
	_ = concMgr.Acquire("deploy", 2)
	req := httptest.NewRequest(http.MethodGet, "/v1/concurrency", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestHandleDemo_BuildkiteProof(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	proof, _ := resp["proof"].(map[string]any)
	if proof["dynamic_pipeline_supported"] != true {
		t.Fatalf("proof=%v", proof)
	}
}

func TestProofForDefault(t *testing.T) {
	p := proofFor("UNKNOWN")
	if p["vertical_slice"] != true {
		t.Fatalf("expected vertical_slice proof")
	}
}
