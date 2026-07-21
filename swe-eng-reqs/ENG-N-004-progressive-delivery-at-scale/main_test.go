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

func TestHandlePlan_EvaluatePromote(t *testing.T) {
	store = NewScaleStore()
	mux := newMux()
	body, _ := json.Marshal(map[string]any{
		"envs":     []string{"dev", "staging", "prod"},
		"criteria": map[string]float64{"max_error_rate": 0.01, "min_success": 0.99},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var plan Plan
	if err := json.Unmarshal(rr.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	evalBody, _ := json.Marshal(map[string]float64{"error_rate": 0.001, "success_rate": 0.999})
	req = httptest.NewRequest(http.MethodPost, "/v1/plans/"+plan.ID+"/evaluate", bytes.NewReader(evalBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("evaluate status=%d body=%s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/plans/"+plan.ID+"/promote", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("promote status=%d body=%s", rr.Code, rr.Body.String())
	}
	var promoted Plan
	if err := json.Unmarshal(rr.Body.Bytes(), &promoted); err != nil {
		t.Fatalf("decode promote: %v", err)
	}
	if promoted.CurrentEnv != "staging" {
		t.Fatalf("after promote env=%q, want staging", promoted.CurrentEnv)
	}
}

func TestHandleDemo_ScaleProof(t *testing.T) {
	store = NewScaleStore()
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
	envs, ok := proof["environments"].(float64)
	if !ok || envs < 2 {
		t.Fatalf("proof.environments=%v, want >=2", proof["environments"])
	}
	if proof["criteria_passed"] != true {
		t.Fatalf("proof.criteria_passed=%v, want true", proof["criteria_passed"])
	}
	if proof["auto_promoted"] != true {
		t.Fatalf("proof.auto_promoted=%v, want true", proof["auto_promoted"])
	}
}
