package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleFindMissing_POST(t *testing.T) {
	resetBazelCAS()
	
	d1 := bazelCAS.Write([]byte("present"))
	missingDigest := "0000000000000000000000000000000000000000000000000000000000000000"
	
	body, _ := json.Marshal(map[string]any{
		"digests": []string{d1, missingDigest},
	})
	
	req := httptest.NewRequest(http.MethodPost, "/v1/cas/find_missing", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	handleFindMissing(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("POST /v1/cas/find_missing status = %d, want 200", rec.Code)
	}
	
	var resp struct {
		Missing []string `json:"missing"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if len(resp.Missing) != 1 || resp.Missing[0] != missingDigest {
		t.Errorf("missing = %v, want [%q]", resp.Missing, missingDigest)
	}
}

func TestHandleBatchRead_POST(t *testing.T) {
	resetBazelCAS()
	
	d1 := bazelCAS.Write([]byte("read me"))
	
	body, _ := json.Marshal(map[string]any{
		"digests": []string{d1},
	})
	
	req := httptest.NewRequest(http.MethodPost, "/v1/cas/batch_read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	handleBatchRead(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("POST /v1/cas/batch_read status = %d, want 200", rec.Code)
	}
	
	var resp struct {
		Blobs map[string]string `json:"blobs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if resp.Blobs[d1] != "cmVhZCBtZQ==" {
		t.Logf("Blob content (base64): %q", resp.Blobs[d1])
	}
}

func TestHandleBatchWrite_POST(t *testing.T) {
	resetBazelCAS()
	
	body, _ := json.Marshal(map[string]any{
		"blobs": map[string]string{
			"0000000000000000000000000000000000000000000000000000000000000001": "dGVzdA==",
		},
	})
	
	req := httptest.NewRequest(http.MethodPost, "/v1/cas/batch_write", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	handleBatchWrite(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("POST /v1/cas/batch_write status = %d, want 200", rec.Code)
	}
	
	var resp struct {
		Results map[string]bool `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	if !resp.Results["0000000000000000000000000000000000000000000000000000000000000001"] {
		t.Error("expected write success for digest")
	}
}

func TestHandleDemo_BazelProof(t *testing.T) {
	resetBazelCAS()
	
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rec := httptest.NewRecorder()
	
	handleDemo(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/demo status = %d, want 200", rec.Code)
	}
	
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	
	proof, ok := resp["proof"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'proof' object")
	}
	
	if _, ok := proof["hermetic_inputs"]; !ok {
		t.Error("proof missing 'hermetic_inputs'")
	}
	if _, ok := proof["find_missing_supported"]; !ok {
		t.Error("proof missing 'find_missing_supported'")
	}
	if _, ok := proof["bazel_compatible"]; !ok {
		t.Error("proof missing 'bazel_compatible'")
	}
}

func resetBazelCAS() {
	bazelCAS = NewBazelCAS(10000)
}
