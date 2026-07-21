package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewController())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleWorkload_ReconcileFinalize(t *testing.T) {
	mux := newMux(NewController())

	create := httptest.NewRequest(http.MethodPost, "/v1/workloads",
		bytes.NewBufferString(`{"id":"orders","replicas":2,"image":"example/orders:v1"}`))
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}

	reconcile := httptest.NewRequest(http.MethodPost, "/v1/reconcile",
		bytes.NewBufferString(`{"id":"orders"}`))
	reconciled := httptest.NewRecorder()
	mux.ServeHTTP(reconciled, reconcile)
	if reconciled.Code != http.StatusOK {
		t.Fatalf("reconcile status=%d body=%s", reconciled.Code, reconciled.Body.String())
	}

	remove := httptest.NewRequest(http.MethodDelete, "/v1/workloads/orders", nil)
	deleting := httptest.NewRecorder()
	mux.ServeHTTP(deleting, remove)
	if deleting.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleting.Code, deleting.Body.String())
	}

	finalize := httptest.NewRequest(http.MethodPost, "/v1/workloads/orders/finalize", nil)
	finalized := httptest.NewRecorder()
	mux.ServeHTTP(finalized, finalize)
	if finalized.Code != http.StatusOK {
		t.Fatalf("finalize status=%d body=%s", finalized.Code, finalized.Body.String())
	}

	var proof map[string]any
	if err := json.Unmarshal(finalized.Body.Bytes(), &proof); err != nil {
		t.Fatalf("decode finalize response: %v", err)
	}
	if proof["finalizer_cleared"] != true {
		t.Fatalf("expected finalizer_cleared=true, got %#v", proof)
	}
}

func TestHandleDemo_OperatorProof(t *testing.T) {
	mux := newMux(NewController())
	request := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("demo status=%d body=%s", response.Code, response.Body.String())
	}

	var result struct {
		Proof map[string]any `json:"proof"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode demo: %v", err)
	}
	for _, key := range []string{"crd", "reconciled", "conditions", "finalizer_cleared"} {
		if _, ok := result.Proof[key]; !ok {
			t.Errorf("proof missing %q: %#v", key, result.Proof)
		}
	}
	if result.Proof["crd"] != true || result.Proof["reconciled"] != true || result.Proof["finalizer_cleared"] != true {
		t.Fatalf("operator proof incomplete: %#v", result.Proof)
	}
}
