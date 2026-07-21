package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewFacade(3, 2))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleObjects_PUT_GET(t *testing.T) {
	mux := newMux(NewFacade(3, 2))
	put := httptest.NewRequest(http.MethodPut, "/v1/objects", strings.NewReader("handler blob"))
	putRR := httptest.NewRecorder()
	mux.ServeHTTP(putRR, put)
	if putRR.Code != http.StatusCreated {
		t.Fatalf("PUT status = %d, body=%s", putRR.Code, putRR.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(putRR.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	digest := response["digest"].(string)

	get := httptest.NewRequest(http.MethodGet, "/v1/objects/"+digest, nil)
	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, get)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET status = %d", getRR.Code)
	}
	got, _ := io.ReadAll(getRR.Body)
	if string(got) != "handler blob" {
		t.Fatalf("GET body = %q", got)
	}
	if getRR.Header().Get("X-Artifact-Replicas") == "" {
		t.Fatal("missing durability replica header")
	}
}

func TestHandleDurability(t *testing.T) {
	f := NewFacade(3, 2)
	digest, _ := f.Put([]byte("metadata"))
	mux := newMux(f)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/durability/"+digest, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	for _, key := range []string{"replicas", "checksum", "healthy_nodes"} {
		if !strings.Contains(rr.Body.String(), key) {
			t.Fatalf("response missing %q: %s", key, rr.Body.String())
		}
	}
}

func TestHandleDemo_DurableProof(t *testing.T) {
	mux := newMux(NewFacade(3, 2))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	for _, proof := range []string{`"nodes": 3`, `"replicas": 3`, `"durable": true`} {
		if !strings.Contains(rr.Body.String(), proof) {
			t.Fatalf("demo missing %s: %s", proof, rr.Body.String())
		}
	}
}
