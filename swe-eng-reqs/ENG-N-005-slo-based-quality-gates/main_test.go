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

func TestHandleGate_EvaluateAllowDeny(t *testing.T) {
	gates = NewGateEngine()
	mux := newMux()

	sloBody, _ := json.Marshal(map[string]any{
		"objective": 0.999,
		"threshold": 14.4,
	})
	req := httptest.NewRequest(http.MethodPut, "/v1/slos/demo", bytes.NewReader(sloBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		t.Fatalf("put slo status=%d body=%s", rr.Code, rr.Body.String())
	}

	seriesBody, _ := json.Marshal(map[string]any{
		"errors_short": 0, "total_short": 1000,
		"errors_long":  1, "total_long": 100000,
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/series/demo", bytes.NewReader(seriesBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		t.Fatalf("series status=%d body=%s", rr.Code, rr.Body.String())
	}

	evalBody, _ := json.Marshal(map[string]any{
		"slo_id":    "demo",
		"burn_rate": 999.0, // client-supplied — must be ignored
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/gates/evaluate", bytes.NewReader(evalBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("evaluate status=%d body=%s", rr.Code, rr.Body.String())
	}
	var dec GateDecision
	if err := json.Unmarshal(rr.Body.Bytes(), &dec); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dec.Decision != "allow" {
		t.Fatalf("decision=%q, want allow (client burn_rate ignored)", dec.Decision)
	}
}

func TestHandleInfo_PromQLInspired(t *testing.T) {
	gates = NewGateEngine()
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("info status=%d", rr.Code)
	}
	var info map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if info["promql_inspired"] != true {
		t.Fatalf("promql_inspired=%v, want true", info["promql_inspired"])
	}
	if info["simulator"] != true {
		t.Fatalf("simulator=%v, want true", info["simulator"])
	}
}

func TestHandleDemo_GateProof(t *testing.T) {
	gates = NewGateEngine()
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
	if proof["promql_inspired"] != true {
		t.Fatalf("promql_inspired=%v", proof["promql_inspired"])
	}
	if proof["allow"] != true {
		t.Fatalf("allow path not demonstrated: %v", proof["allow"])
	}
	if proof["deny"] != true {
		t.Fatalf("deny path not demonstrated: %v", proof["deny"])
	}
	if _, ok := proof["burn_rate"]; !ok {
		t.Fatal("missing burn_rate in proof")
	}
	if _, ok := proof["evidence"]; !ok {
		t.Fatal("missing evidence in proof")
	}
}
