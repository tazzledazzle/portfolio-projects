package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewReliabilityStore("runbooks"))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleSLO_Golden_Runbooks(t *testing.T) {
	runbookDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(runbookDir, "error-budget.md"), []byte("# Error budget"), 0o600); err != nil {
		t.Fatalf("write runbook: %v", err)
	}
	mux := newMux(NewReliabilityStore(runbookDir))

	put := httptest.NewRequest(http.MethodPut, "/v1/slos/api",
		bytes.NewBufferString(`{"objective":0.999,"sli":"successful_requests / total_requests"}`))
	putResult := httptest.NewRecorder()
	mux.ServeHTTP(putResult, put)
	if putResult.Code != http.StatusOK {
		t.Fatalf("PUT SLO status=%d body=%s", putResult.Code, putResult.Body.String())
	}

	for _, path := range []string{"/v1/golden-signals", "/v1/runbooks"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}

func TestHandleDemo_ReliabilityProof(t *testing.T) {
	mux := newMux(NewReliabilityStore("runbooks"))
	request := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("demo status=%d body=%s", response.Code, response.Body.String())
	}

	var result struct {
		Proof map[string]any `json:"proof"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode demo: %v", err)
	}
	if result.Proof["slo"] != true {
		t.Fatalf("expected slo proof, got %#v", result.Proof)
	}
	signals, ok := result.Proof["golden_signals"].([]any)
	if !ok || len(signals) != 4 {
		t.Fatalf("expected four golden signals, got %#v", result.Proof["golden_signals"])
	}
	if count, ok := result.Proof["runbook_count"].(float64); !ok || count < 1 {
		t.Fatalf("expected runbook_count >= 1, got %#v", result.Proof["runbook_count"])
	}
}
