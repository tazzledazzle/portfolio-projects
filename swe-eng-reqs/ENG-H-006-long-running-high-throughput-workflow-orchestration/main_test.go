package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewWorkflowEngine())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_DurableProof(t *testing.T) {
	eng := NewWorkflowEngine()
	mux := newMux(eng)
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
	if proof["durable"] != true {
		t.Fatalf("durable want true, got %v", proof["durable"])
	}
	if proof["replay_safe"] != true {
		t.Fatalf("replay_safe want true, got %v", proof["replay_safe"])
	}
	tp, ok := proof["throughput_per_s"].(float64)
	if !ok || tp <= 0 {
		t.Fatalf("throughput_per_s want > 0, got %v", proof["throughput_per_s"])
	}
	sc, ok := proof["steps_completed"].(float64)
	if !ok || sc < 4 {
		t.Fatalf("steps_completed want >= 4, got %v", proof["steps_completed"])
	}
}

func TestHandleWorkflows_StartAndSignal(t *testing.T) {
	eng := NewWorkflowEngine()
	mux := newMux(eng)

	create := httptest.NewRequest(http.MethodPost, "/v1/workflows", strings.NewReader(`{"name":"api-wf","steps":["a","b"]}`))
	create.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, create)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var wf Workflow
	if err := json.Unmarshal(rr.Body.Bytes(), &wf); err != nil {
		t.Fatalf("decode: %v", err)
	}

	sig := httptest.NewRequest(http.MethodPost, "/v1/workflows/"+wf.ID+"/signal", strings.NewReader(`{"action":"advance","event_id":"e1"}`))
	sig.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, sig)
	if rr2.Code != 200 {
		t.Fatalf("signal: %d %s", rr2.Code, rr2.Body.String())
	}
	var out Workflow
	if err := json.Unmarshal(rr2.Body.Bytes(), &out); err != nil {
		t.Fatalf("signal decode: %v", err)
	}
	if out.StepsCompleted != 1 {
		t.Fatalf("steps_completed want 1, got %d", out.StepsCompleted)
	}
}
