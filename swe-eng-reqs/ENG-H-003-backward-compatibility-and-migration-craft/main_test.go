package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewMigrateStore())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_MigrateProof(t *testing.T) {
	mux := newMux(NewMigrateStore())
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
	if proof["dual_write"] != true {
		t.Fatalf("expected dual_write, got %#v", proof)
	}
	if proof["v1_readable"] != true {
		t.Fatalf("expected v1_readable, got %#v", proof)
	}
	if proof["v2_readable"] != true {
		t.Fatalf("expected v2_readable, got %#v", proof)
	}
	if proof["compat_pass"] != true {
		t.Fatalf("expected compat_pass, got %#v", proof)
	}
}
