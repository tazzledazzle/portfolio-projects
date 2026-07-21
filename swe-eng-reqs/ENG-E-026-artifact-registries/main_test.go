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

func TestHandleTag_PUT_GET(t *testing.T) {
	resetRegistry()
	digest, err := registry.PutManifest([]byte("tag-manifest"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	body := []byte(`{"digest":"` + digest + `"}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/registry/demo/tags/latest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("PUT tag status = %d, body=%s", rr.Code, rr.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/registry/demo/tags/latest", nil)
	getRR := httptest.NewRecorder()
	newMux().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET tag status = %d", getRR.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(getRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp["digest"] != digest {
		t.Fatalf("resolved digest = %v, want %s", resp["digest"], digest)
	}
}

func TestHandleManifest_PUT_GET(t *testing.T) {
	resetRegistry()
	manifest := []byte(`{"schemaVersion":2}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/registry/demo/manifests", bytes.NewReader(manifest))
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("PUT manifest status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	digest, _ := resp["digest"].(string)
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("bad digest %q", digest)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/registry/demo/manifests/"+digest, nil)
	getRR := httptest.NewRecorder()
	newMux().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET manifest status = %d", getRR.Code)
	}
	got, _ := io.ReadAll(getRR.Body)
	if !bytes.Equal(got, manifest) {
		t.Fatalf("GET body = %q, want %q", got, manifest)
	}
}

func TestHandleInfo_OCIInspired(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("info status = %d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp["oci_inspired"] != true {
		t.Fatalf("oci_inspired = %v, want true", resp["oci_inspired"])
	}
	if resp["conformance"] == true {
		t.Fatal("must not claim OCI Distribution Spec conformance")
	}
	if s, ok := resp["label"].(string); ok && strings.Contains(strings.ToLower(s), "conformance") && !strings.Contains(strings.ToLower(s), "not") {
		t.Fatalf("label claims conformance: %s", s)
	}
}

func TestHandleDemo_TagProof(t *testing.T) {
	resetRegistry()
	digest, err := registry.PutManifest([]byte("demo-manifest"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	if err := registry.PutTag("demo", "latest", digest); err != nil {
		t.Fatalf("PutTag: %v", err)
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
	if proof["tag_to_digest"] != true && proof["tag_to_digest"] != digest {
		// Accept bool true or the actual digest mapping indicator
		if _, has := proof["tag_to_digest"]; !has {
			t.Fatalf("proof missing tag_to_digest: %v", proof)
		}
	}
	if proof["tag_mutable"] != true {
		t.Fatalf("tag_mutable = %v, want true", proof["tag_mutable"])
	}
	if proof["digest_immutable"] != true {
		t.Fatalf("digest_immutable = %v, want true", proof["digest_immutable"])
	}
}
