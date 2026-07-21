package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthz(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.0.3\n"), 0o644)
	mux := newMux(NewAPIEngine(10, "resources:read"), dir)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleInfo_OpenAPIInspired(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.0.3\ninfo:\n  title: t\n"), 0o644)
	mux := newMux(NewAPIEngine(10, "resources:read"), dir)
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("info %d", rr.Code)
	}
	var info map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["openapi_inspired"] != true && info["openapi"] != true {
		t.Fatalf("expected openapi honesty label: %#v", info)
	}
}

func TestHandleDemo_APIProof(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.0.3\npaths: {}\n"), 0o644)
	mux := newMux(NewAPIEngine(2, "resources:read"), dir)
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("demo %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	proof, _ := result["proof"].(map[string]any)
	if proof["openapi"] != true || proof["authz_enforced"] != true ||
		proof["rate_limited"] != true || proof["compat_pass"] != true {
		t.Fatalf("proof incomplete: %#v", proof)
	}
}
