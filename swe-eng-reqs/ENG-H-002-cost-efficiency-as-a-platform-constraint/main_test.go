package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewCostMeter())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_CostProof(t *testing.T) {
	mux := newMux(NewCostMeter())
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
	proof, _ := body["proof"].(map[string]any)
	cost, _ := proof["cost_per_build_usd"].(float64)
	if cost <= 0 {
		t.Fatalf("expected cost_per_build_usd > 0, got %#v", proof)
	}
	savings, ok := proof["cache_savings_pct"].(float64)
	if !ok {
		t.Fatalf("expected cache_savings_pct number, got %#v", proof)
	}
	if savings < 0 {
		t.Fatalf("expected non-negative cache_savings_pct, got %v", savings)
	}
}
