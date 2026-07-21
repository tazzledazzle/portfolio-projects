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

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProofForDefault(t *testing.T) {
	p := proofFor("UNKNOWN")
	if p["vertical_slice"] != true {
		t.Fatalf("expected vertical_slice proof")
	}
}

func TestHandleBlobs_PUT_GET(t *testing.T) {
	resetStore()
	body := []byte("hello blob")
	req := httptest.NewRequest(http.MethodPut, "/v1/blobs", bytes.NewReader(body))
	req.Header.Set("X-Meta-content-type", "text/plain")
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("PUT status = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	digest, _ := resp["digest"].(string)
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
		t.Fatalf("bad digest %q", digest)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/blobs/"+digest, nil)
	getRR := httptest.NewRecorder()
	newMux().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", getRR.Code)
	}
	got, _ := io.ReadAll(getRR.Body)
	if !bytes.Equal(got, body) {
		t.Fatalf("GET body = %q, want %q", got, body)
	}
}

func TestHandleBlobs_Conflict(t *testing.T) {
	resetStore()
	orig := []byte("hello")
	digest, err := store.Put(orig, nil)
	if err != nil {
		t.Fatalf("seed Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/v1/blobs/"+digest, bytes.NewReader([]byte("world")))
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("conflict PUT status = %d, want 409; body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	resetStore()
	_, err := store.Put([]byte("demo-bytes"), map[string]string{"content-type": "application/octet-stream"})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo status = %d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	proof, ok := resp["proof"].(map[string]any)
	if !ok {
		t.Fatal("missing proof")
	}
	if proof["digest_immutable"] != true {
		t.Fatalf("digest_immutable = %v, want true", proof["digest_immutable"])
	}
	count, ok := proof["blob_count"].(float64)
	if !ok || count < 1 {
		t.Fatalf("blob_count = %v, want >= 1", proof["blob_count"])
	}
	keys, ok := proof["metadata_keys"].([]any)
	if !ok || len(keys) == 0 {
		t.Fatalf("metadata_keys = %v, want non-empty", proof["metadata_keys"])
	}
}
