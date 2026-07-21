package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewController())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDevEnvs_POST_GET_Tick(t *testing.T) {
	mux := newMux(NewController())
	postRR := httptest.NewRecorder()
	body := `{"id":"preview-1","ttl_seconds":60}`
	mux.ServeHTTP(postRR, httptest.NewRequest(http.MethodPost, "/v1/devenvs", strings.NewReader(body)))
	if postRR.Code != http.StatusCreated {
		t.Fatalf("POST status=%d body=%s", postRR.Code, postRR.Body.String())
	}
	var env DevEnv
	if err := json.Unmarshal(postRR.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}

	tickRR := httptest.NewRecorder()
	tickBody := `{"now":"` + env.CreatedAt.Add(61_000_000_000).Format("2006-01-02T15:04:05Z07:00") + `"}`
	mux.ServeHTTP(tickRR, httptest.NewRequest(http.MethodPost, "/v1/devenvs/preview-1/tick", strings.NewReader(tickBody)))
	if tickRR.Code != http.StatusOK || !strings.Contains(tickRR.Body.String(), `"reclaimed": true`) {
		t.Fatalf("tick status=%d body=%s", tickRR.Code, tickRR.Body.String())
	}

	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/v1/devenvs/preview-1", nil))
	if getRR.Code != http.StatusOK || !strings.Contains(getRR.Body.String(), `"Expired"`) {
		t.Fatalf("GET status=%d body=%s", getRR.Code, getRR.Body.String())
	}
}

func TestHandleReconcile(t *testing.T) {
	c := NewController()
	mux := newMux(c)
	_, _ = c.Create("expired", 1, nowUTC().Add(-2_000_000_000))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/reconcile", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"reclaimed": 1`) {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleDemo_TTLProof(t *testing.T) {
	mux := newMux(NewController())
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/demo", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	for _, key := range []string{`"crd": true`, `"ttl_reclaimed": true`, `"Ready"`, `"Expired"`} {
		if !strings.Contains(rr.Body.String(), key) {
			t.Fatalf("demo missing %s: %s", key, rr.Body.String())
		}
	}
}
