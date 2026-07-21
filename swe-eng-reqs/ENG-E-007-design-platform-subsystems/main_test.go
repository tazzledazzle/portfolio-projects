package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	design = mustStore(t)
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_ADRProof(t *testing.T) {
	design = mustStore(t)
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
	adrCount, _ := proof["adr_count"].(float64)
	if adrCount < 1 {
		t.Fatalf("expected adr_count≥1, got %#v", proof["adr_count"])
	}
	alts, _ := proof["alternatives"].(float64)
	if alts < 2 {
		t.Fatalf("expected alternatives≥2, got %#v", proof["alternatives"])
	}
	if proof["decision_recorded"] != true {
		t.Fatalf("expected decision_recorded=true, got %#v", proof["decision_recorded"])
	}
}

func mustStore(t *testing.T) *DesignStore {
	t.Helper()
	s, err := NewDesignStore("adr")
	if err != nil {
		t.Fatalf("NewDesignStore: %v", err)
	}
	return s
}
