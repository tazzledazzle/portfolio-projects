package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewMultiDC())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	mux := newMux(NewMultiDC())
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
	if proof == nil {
		t.Fatalf("missing proof: %#v", body)
	}
	if int(proof["data_centers"].(float64)) < 40 {
		t.Fatalf("expected ≥40 data_centers, got %#v", proof["data_centers"])
	}
	if proof["fanout_ok"] != true {
		t.Fatalf("expected fanout_ok, got %#v", proof)
	}
	if proof["multi_dc_simulator"] != true {
		t.Fatalf("expected multi_dc_simulator, got %#v", proof)
	}
}

func TestHandleInfo_Simulator(t *testing.T) {
	mux := newMux(NewMultiDC())
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var info map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if info["simulator"] != true || info["multi_dc_simulator"] != true {
		t.Fatalf("expected simulator labels, got %#v", info)
	}
}
