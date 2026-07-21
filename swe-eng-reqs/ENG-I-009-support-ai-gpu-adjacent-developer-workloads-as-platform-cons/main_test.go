package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewGPUConsumer())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_ChunkedProof(t *testing.T) {
	c := NewGPUConsumer()
	mux := newMux(c)
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	proof, _ := body["proof"].(map[string]any)
	if proof == nil {
		t.Fatal("missing proof")
	}
	if proof["long_running"] != true {
		t.Fatalf("long_running want true, got %v", proof["long_running"])
	}
	if proof["chunked_upload"] != true {
		t.Fatalf("chunked_upload want true, got %v", proof["chunked_upload"])
	}
	if proof["timeout"] != true {
		t.Fatalf("timeout want true, got %v", proof["timeout"])
	}
	br, ok := proof["bytes_received"].(float64)
	if !ok || br <= 0 {
		t.Fatalf("bytes_received want > 0, got %v", proof["bytes_received"])
	}
}

func TestHandleJobs_StartAndChunk(t *testing.T) {
	c := NewGPUConsumer()
	mux := newMux(c)

	create := httptest.NewRequest(http.MethodPost, "/v1/jobs", strings.NewReader(`{"name":"api-job","timeout_ms":5000}`))
	create.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, create)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var job GPUJob
	if err := json.Unmarshal(rr.Body.Bytes(), &job); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !job.LongRunning {
		t.Fatal("expected long_running")
	}

	chunk := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+job.ID+"/chunks?index=0", bytes.NewReader([]byte("payload")))
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, chunk)
	if rr2.Code != 200 {
		t.Fatalf("chunk: %d %s", rr2.Code, rr2.Body.String())
	}
	var out GPUJob
	if err := json.Unmarshal(rr2.Body.Bytes(), &out); err != nil {
		t.Fatalf("chunk decode: %v", err)
	}
	if out.BytesReceived != 7 {
		t.Fatalf("bytes_received want 7, got %d", out.BytesReceived)
	}
}

func TestHandleChunk_TimeoutGone(t *testing.T) {
	c := NewGPUConsumer()
	mux := newMux(c)
	job, err := c.StartJob("http-timeout", 25*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(40 * time.Millisecond)
	chunk := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+job.ID+"/chunks", bytes.NewReader([]byte("late")))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, chunk)
	if rr.Code != http.StatusGone {
		t.Fatalf("want 410 Gone, got %d %s", rr.Code, rr.Body.String())
	}
}
