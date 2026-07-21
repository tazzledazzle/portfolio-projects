package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleFlakes_GET_Empty(t *testing.T) {
	resetFlakeStore()
	
	req := httptest.NewRequest(http.MethodGet, "/v1/flakes", nil)
	rec := httptest.NewRecorder()
	
	handleFlakes(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/flakes status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if len(resp) != 0 {
		t.Errorf("expected empty map, got %v", resp)
	}
}

func TestHandleFlakes_POST_JUnit(t *testing.T) {
	resetFlakeStore()
	
	xml := `<?xml version="1.0"?>
<testsuite name="suite1" tests="2">
  <testcase name="test1" classname="pkg.Test" time="0.1"/>
  <testcase name="test2" classname="pkg.Test" time="0.2">
    <failure message="oops">failed</failure>
  </testcase>
</testsuite>`
	
	req := httptest.NewRequest(http.MethodPost, "/v1/flakes", strings.NewReader(xml))
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()
	
	handleFlakes(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("POST /v1/flakes status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if _, ok := resp["parsed"]; !ok {
		t.Error("response missing 'parsed' field")
	}
}

func TestHandleFlakes_GET_TestID(t *testing.T) {
	resetFlakeStore()
	
	xml := `<?xml version="1.0"?>
<testsuite name="suite1" tests="1">
  <testcase name="test1" classname="pkg.Test" time="0.1"/>
</testsuite>`
	
	postReq := httptest.NewRequest(http.MethodPost, "/v1/flakes", strings.NewReader(xml))
	postReq.Header.Set("Content-Type", "application/xml")
	postRec := httptest.NewRecorder()
	handleFlakes(postRec, postReq)
	
	req := httptest.NewRequest(http.MethodGet, "/v1/flakes/test1", nil)
	rec := httptest.NewRecorder()
	
	handleFlakeByID(rec, req, "test1")
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/flakes/test1 status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var resp struct {
		TestID      string  `json:"test_id"`
		Score       float64 `json:"score"`
		Quarantined bool    `json:"quarantined"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if resp.TestID != "test1" {
		t.Errorf("test_id = %q, want %q", resp.TestID, "test1")
	}
}

func TestHandleFlakes_POST_InvalidXML(t *testing.T) {
	resetFlakeStore()
	
	req := httptest.NewRequest(http.MethodPost, "/v1/flakes", strings.NewReader("<invalid"))
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()
	
	handleFlakes(rec, req)
	
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/flakes with invalid XML status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleQuarantine_GET(t *testing.T) {
	resetFlakeStore()
	
	flakeStore["flaky_test"] = NewFlakeScore()
	flakeStore["flaky_test"].Update(false)
	flakeStore["flaky_test"].Update(false)
	
	flakeStore["stable_test"] = NewFlakeScore()
	flakeStore["stable_test"].Update(true)
	
	req := httptest.NewRequest(http.MethodGet, "/v1/quarantine", nil)
	rec := httptest.NewRecorder()
	
	handleQuarantine(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/quarantine status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var resp struct {
		Quarantined []string `json:"quarantined"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	found := false
	for _, id := range resp.Quarantined {
		if id == "flaky_test" {
			found = true
		}
		if id == "stable_test" {
			t.Error("stable_test should not be quarantined")
		}
	}
	if !found {
		t.Error("flaky_test should be in quarantine list")
	}
}

func TestHandleDemo_FlakeProof(t *testing.T) {
	resetFlakeStore()
	
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rec := httptest.NewRecorder()
	
	handleDemo(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/demo status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	proof, ok := resp["proof"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'proof' object")
	}
	
	if _, ok := proof["flake_score"]; !ok {
		t.Error("proof missing 'flake_score'")
	}
	if _, ok := proof["quarantined_count"]; !ok {
		t.Error("proof missing 'quarantined_count'")
	}
	if _, ok := proof["junit_parsed"]; !ok {
		t.Error("proof missing 'junit_parsed'")
	}
}

func resetFlakeStore() {
	flakeStoreMu.Lock()
	defer flakeStoreMu.Unlock()
	flakeStore = make(map[string]*FlakeScore)
	junitParsed = false
}

var _ = bytes.Buffer{}
