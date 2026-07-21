package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewServiceRuntime())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleEcho_RequestID(t *testing.T) {
	mux := newMux(NewServiceRuntime())
	req := httptest.NewRequest(http.MethodPost, "/v1/echo", strings.NewReader(`{"message":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "echo-req-42")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("X-Request-ID"); got != "echo-req-42" {
		t.Fatalf("expected X-Request-ID header echo-req-42, got %q", got)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["request_id"] != "echo-req-42" {
		t.Fatalf("expected body request_id, got %#v", body)
	}
}

func TestHandleDemo_ProdProof(t *testing.T) {
	mux := newMux(NewServiceRuntime())
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
	for _, key := range []string{"metrics_exposed", "request_id", "structured_error"} {
		v, present := proof[key]
		if !present {
			t.Fatalf("proof missing %s", key)
		}
		switch typed := v.(type) {
		case bool:
			if !typed {
				t.Fatalf("proof.%s must be true", key)
			}
		case string:
			if typed == "" {
				t.Fatalf("proof.%s must be non-empty", key)
			}
		default:
			t.Fatalf("proof.%s unexpected type %#v", key, v)
		}
	}
}
