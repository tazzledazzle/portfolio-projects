package main

import (
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

func TestHandleInfo_OR(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["language"] != "go" {
		t.Fatalf("expected language=go, got %#v", body["language"])
	}
	if body["or_semantics"] != true {
		t.Fatalf("expected or_semantics=true, got %#v", body["or_semantics"])
	}
}

func TestHandleDemo_ORProof(t *testing.T) {
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
	if proof["language"] != "go" {
		t.Fatalf("expected proof.language=go, got %#v", proof["language"])
	}
	if proof["or_semantics"] != true {
		t.Fatalf("expected proof.or_semantics=true, got %#v", proof["or_semantics"])
	}
	if proof["production_sample"] != true {
		t.Fatalf("expected proof.production_sample=true, got %#v", proof["production_sample"])
	}
}
