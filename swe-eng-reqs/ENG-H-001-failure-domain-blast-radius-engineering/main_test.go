package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewBlastEngine())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	mux := newMux(NewBlastEngine())
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
	if proof["chaos_ran"] != true || proof["contained"] != true {
		t.Fatalf("expected chaos_ran+contained, got %#v", proof)
	}
	unaffected, _ := proof["unaffected_tenants"].([]any)
	if len(unaffected) == 0 {
		t.Fatalf("expected unaffected_tenants, got %#v", proof)
	}
}

func TestHandleBlastRadius_ServerComputed(t *testing.T) {
	eng := NewBlastEngine()
	_ = eng.RegisterDomain("fd-a", []string{"t1"})
	_ = eng.RegisterDomain("fd-b", []string{"t2"})
	_, _ = eng.RunChaos("fd-a", "crash")
	mux := newMux(eng)
	req := httptest.NewRequest(http.MethodGet, "/v1/blast-radius", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var radius BlastRadiusReport
	if err := json.Unmarshal(rr.Body.Bytes(), &radius); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !radius.Contained || !radius.ChaosRan {
		t.Fatalf("expected contained server report, got %#v", radius)
	}
}
