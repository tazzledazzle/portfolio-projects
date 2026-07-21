package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleCache_PUT(t *testing.T) {
	resetCache()
	
	blob := []byte("test blob content")
	digest := computeDigest(blob)
	
	req := httptest.NewRequest(http.MethodPut, "/cache/cas/"+digest, bytes.NewReader(blob))
	rec := httptest.NewRecorder()
	
	handleCacheCAS(rec, req)
	
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Errorf("PUT /cache/cas/:digest status = %d, want 200 or 201", rec.Code)
	}
}

func TestHandleCache_GET_Hit(t *testing.T) {
	resetCache()
	
	blob := []byte("test blob content")
	digest := cache.Put("default", blob)
	
	req := httptest.NewRequest(http.MethodGet, "/cache/cas/"+digest, nil)
	rec := httptest.NewRecorder()
	
	handleCacheCAS(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("GET /cache/cas/:digest status = %d, want 200", rec.Code)
	}
	
	body, _ := io.ReadAll(rec.Body)
	if !bytes.Equal(body, blob) {
		t.Errorf("GET body = %q, want %q", string(body), string(blob))
	}
}

func TestHandleCache_GET_Miss(t *testing.T) {
	resetCache()
	
	req := httptest.NewRequest(http.MethodGet, "/cache/cas/invaliddigesthere1234567890123456789012345678901234", nil)
	rec := httptest.NewRecorder()
	
	handleCacheCAS(rec, req)
	
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /cache/cas/invalid status = %d, want 404", rec.Code)
	}
}

func TestHandleCache_Metrics(t *testing.T) {
	resetCache()
	
	blob := []byte("metrics test")
	digest := cache.Put("default", blob)
	cache.Get("default", digest)
	cache.Get("default", "miss")
	
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	
	handleMetrics(rec, req)
	
	body := rec.Body.String()
	
	if !strings.Contains(body, "cache_hits_total") {
		t.Error("metrics missing cache_hits_total")
	}
	if !strings.Contains(body, "cache_misses_total") {
		t.Error("metrics missing cache_misses_total")
	}
}

func TestHandleDemo_CacheProof(t *testing.T) {
	resetCache()
	
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
	
	if _, ok := proof["cache_hit_rate"]; !ok {
		t.Error("proof missing 'cache_hit_rate'")
	}
	if _, ok := proof["cas_enabled"]; !ok {
		t.Error("proof missing 'cas_enabled'")
	}
}

func resetCache() {
	cache = NewCASCache(1000)
}

var _ = bytes.Buffer{}
