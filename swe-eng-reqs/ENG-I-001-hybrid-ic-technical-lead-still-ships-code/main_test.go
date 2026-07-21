package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewHybridStore("artifacts/leadership"))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_HybridProof(t *testing.T) {
	mux := newMux(NewHybridStore("artifacts/leadership"))
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	proof, ok := body["proof"].(map[string]any)
	if !ok {
		t.Fatalf("expected proof object, got %#v", body["proof"])
	}
	for _, key := range []string{"code_shipped", "leadership_artifacts", "hybrid_ic"} {
		v, present := proof[key]
		if !present {
			t.Fatalf("proof missing %s", key)
		}
		b, ok := v.(bool)
		if !ok || !b {
			t.Fatalf("proof.%s must be true, got %#v", key, v)
		}
	}
}
