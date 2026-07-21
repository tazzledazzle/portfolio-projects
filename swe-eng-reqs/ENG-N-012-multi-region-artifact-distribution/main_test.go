package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewReplicator([]string{"us-east", "eu-west"}, time.Millisecond))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleRegions_PUT_GET(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, 50*time.Millisecond)
	mux := newMux(r)
	putRR := httptest.NewRecorder()
	mux.ServeHTTP(putRR, httptest.NewRequest(http.MethodPut, "/v1/regions/us-east/blobs", strings.NewReader("regional")))
	if putRR.Code != http.StatusCreated {
		t.Fatalf("PUT status=%d body=%s", putRR.Code, putRR.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(putRR.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	digest := response["digest"].(string)
	missingRR := httptest.NewRecorder()
	mux.ServeHTTP(missingRR, httptest.NewRequest(http.MethodGet, "/v1/regions/eu-west/blobs/"+digest, nil))
	if missingRR.Code != http.StatusNotFound {
		t.Fatalf("pre-sync GET status=%d", missingRR.Code)
	}
	r.Wait()
	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/v1/regions/eu-west/blobs/"+digest, nil))
	if getRR.Code != http.StatusOK || getRR.Body.String() != "regional" {
		t.Fatalf("GET status=%d body=%q", getRR.Code, getRR.Body.String())
	}
}

func TestHandleReplicationStatus(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, 50*time.Millisecond)
	_, _ = r.Put("us-east", []byte("pending"))
	mux := newMux(r)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/replication/status", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"lag_ms"`) {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	r.Wait()
}

func TestHandleDemo_ReplicationProof(t *testing.T) {
	mux := newMux(NewReplicator([]string{"us-east", "eu-west"}, time.Millisecond))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	for _, key := range []string{`"regions"`, `"lag_ms"`, `"replicated": true`} {
		if !strings.Contains(rr.Body.String(), key) {
			t.Fatalf("demo missing %s: %s", key, rr.Body.String())
		}
	}
}
