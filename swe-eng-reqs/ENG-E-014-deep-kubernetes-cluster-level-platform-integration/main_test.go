package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewCIJobController())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDemo_CIJobProof(t *testing.T) {
	mux := newMux(NewCIJobController())
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, key := range []string{`"kind": "CIJob"`, `"job_scheduled": true`, `"conditions"`} {
		if !strings.Contains(body, key) {
			t.Fatalf("demo missing %s: %s", key, body)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	proof, _ := payload["proof"].(map[string]any)
	if proof == nil {
		t.Fatalf("missing proof: %s", body)
	}
	if _, ok := proof["finalizer_cleared"]; ok {
		t.Fatal("demo must not use finalizer_cleared as primary proof")
	}
	complete, _ := proof["complete"].(bool)
	failed, _ := proof["failed"].(bool)
	if !complete && !failed {
		t.Fatalf("demo must prove complete or failed path: %+v", proof)
	}
	// Live demo covers both terminal paths.
	both, _ := proof["complete_and_failed_paths"].(bool)
	if !both {
		t.Fatalf("demo must prove both complete and failed paths: %+v", proof)
	}
}
