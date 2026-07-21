package main

import (
	"bytes"
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

func TestHandleOrch_Tick(t *testing.T) {
	store = NewOrchestrator(NewStubGate("allow"))
	mux := newMux()
	body, _ := json.Marshal(map[string]any{"slo_id": "checkout-slo"})
	req := httptest.NewRequest(http.MethodPost, "/v1/orchestrations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var orc Orchestration
	if err := json.Unmarshal(rr.Body.Bytes(), &orc); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/v1/orchestrations/"+orc.ID+"/tick", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("tick status=%d body=%s", rr.Code, rr.Body.String())
	}
	var ticked Orchestration
	if err := json.Unmarshal(rr.Body.Bytes(), &ticked); err != nil {
		t.Fatalf("decode tick: %v", err)
	}
	if ticked.CurrentStage != "staging" {
		t.Fatalf("after tick stage=%q, want staging", ticked.CurrentStage)
	}
}

func TestHandleDemo_ReleaseStageProof(t *testing.T) {
	store = NewOrchestrator(NewStubGate("allow"))
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	var result map[string]any
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %v", result)
	}
	for _, key := range []string{"stages_advanced", "blocked_on_deny", "gate_required"} {
		if _, has := proof[key]; !has {
			t.Fatalf("proof missing required key %q: %v", key, proof)
		}
	}
	if proof["blocked_on_deny"] != true {
		t.Fatalf("blocked_on_deny=%v, want true", proof["blocked_on_deny"])
	}
	if proof["gate_required"] != true {
		t.Fatalf("gate_required=%v, want true", proof["gate_required"])
	}
	// D-09: proof must NOT contain CI stage vocabulary.
	for _, ci := range []string{`"lint"`, `"unit"`, `"build"`, `"publish"`, "pipeline_stages"} {
		if strings.Contains(body, ci) {
			t.Fatalf("demo proof must not contain CI vocabulary %q", ci)
		}
	}
}
