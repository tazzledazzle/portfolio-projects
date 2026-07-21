package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewAuthzEngine())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_AuthzProof(t *testing.T) {
	mux := newMux(NewAuthzEngine())
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, key := range []string{`"oidc_inspired": true`, `"rbac_allow": true`, `"rbac_deny": true`, `"simulator": true`} {
		if !strings.Contains(body, key) {
			t.Fatalf("demo missing %s: %s", key, body)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	proof, _ := payload["proof"].(map[string]any)
	if proof["external_idp"] == true {
		t.Fatal("demo must not claim external IdP")
	}
}
