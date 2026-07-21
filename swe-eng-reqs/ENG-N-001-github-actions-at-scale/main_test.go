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

func TestHandleWorkflows_POST(t *testing.T) {
	mux := newMux()
	cfg := MatrixConfig{Dimensions: map[string][]string{"os": {"ubuntu"}, "node": {"14", "16"}}}
	raw, _ := json.Marshal(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/workflows", bytes.NewReader(raw))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleRunners_POST_Register(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/runners", bytes.NewReader([]byte(`{"id":"r-test-1"}`)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleRunners_GET_Claim(t *testing.T) {
	mux := newMux()
	_ = runnerPool.Register("r-claim")
	runnerPool.Enqueue(&Job{ID: "claim-job-1", Status: "queued"})
	req := httptest.NewRequest(http.MethodGet, "/v1/runners/r-claim/claim", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleJobs_POST_Complete(t *testing.T) {
	mux := newMux()
	_ = runnerPool.Register("r-complete")
	runnerPool.Enqueue(&Job{ID: "complete-job-1", Status: "queued"})
	_ = runnerPool.ClaimJob("r-complete")
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/complete-job-1/complete", bytes.NewReader([]byte(`{"status":"success"}`)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleDemo_ActionsProof(t *testing.T) {
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
	if proof["matrix_expanded"] != true {
		t.Fatalf("proof=%v", proof)
	}
}

func TestProofForDefault(t *testing.T) {
	p := proofFor("UNKNOWN")
	if p["vertical_slice"] != true {
		t.Fatalf("expected vertical_slice proof")
	}
}
