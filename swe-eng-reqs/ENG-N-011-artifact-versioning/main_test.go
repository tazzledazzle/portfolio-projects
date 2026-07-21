package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleVersionCreate(t *testing.T) {
	versions = NewVersionStore()
	mux := newMux()
	body := map[string]string{
		"name":   "api",
		"digest": digestA,
		"stage":  "dev",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/versions", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got Version
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Digest != digestA || got.Stage != "dev" {
		t.Fatalf("unexpected version: %+v", got)
	}
}

func TestHandlePromote(t *testing.T) {
	versions = NewVersionStore()
	mux := newMux()
	v, err := versions.Create("promo", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/versions/"+v.ID+"/promote", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got Version
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Stage != "staging" || got.Digest != digestA {
		t.Fatalf("promote result: %+v", got)
	}
}

func TestHandleTag(t *testing.T) {
	versions = NewVersionStore()
	mux := newMux()
	body := map[string]string{"digest": digestB}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/v1/tags/release", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	got, ok := versions.GetTag("release")
	if !ok || got != digestB {
		t.Fatalf("GetTag=%q ok=%v", got, ok)
	}
}

func TestHandleDemo_PromoteProof(t *testing.T) {
	versions = NewVersionStore()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %v", result)
	}
	for _, key := range []string{"promoted", "digest_unchanged", "tag_mutable"} {
		if proof[key] != true {
			t.Fatalf("proof[%s]=%v, want true", key, proof[key])
		}
	}
}

func TestProofForDefault(t *testing.T) {
	p := proofFor("UNKNOWN")
	if p["vertical_slice"] != true {
		t.Fatalf("expected vertical_slice proof")
	}
}
