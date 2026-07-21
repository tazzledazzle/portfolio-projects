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

func TestHandleInfo_Honesty(t *testing.T) {
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
	if body["nats_inspired"] != true {
		t.Fatalf("expected nats_inspired=true, got %#v", body["nats_inspired"])
	}
	if body["simulator"] != true {
		t.Fatalf("expected simulator=true, got %#v", body["simulator"])
	}
	if body["nats_connected"] == true {
		t.Fatal("must NOT claim live NATS connectivity")
	}
	if body["bus"] == "nats" {
		t.Fatal("must not claim bus=nats")
	}
}

func TestHandleDemo_BusProof(t *testing.T) {
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
	for _, key := range []string{"nats_inspired", "simulator", "dlq", "replay", "schema_envelope"} {
		if proof[key] != true {
			t.Fatalf("expected proof.%s=true, got %#v", key, proof[key])
		}
	}
}
