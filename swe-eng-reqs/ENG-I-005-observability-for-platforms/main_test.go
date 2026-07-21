package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewOTelStore())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleTrace_Alerts(t *testing.T) {
	mux := newMux(NewOTelStore())

	traceRequest := httptest.NewRequest(http.MethodPost, "/v1/trace",
		bytes.NewBufferString(`{"name":"release.promote"}`))
	traceResponse := httptest.NewRecorder()
	mux.ServeHTTP(traceResponse, traceRequest)
	if traceResponse.Code != http.StatusCreated {
		t.Fatalf("trace status=%d body=%s", traceResponse.Code, traceResponse.Body.String())
	}

	exportRequest := httptest.NewRequest(http.MethodGet, "/v1/traces", nil)
	exportResponse := httptest.NewRecorder()
	mux.ServeHTTP(exportResponse, exportRequest)
	if exportResponse.Code != http.StatusOK {
		t.Fatalf("traces status=%d body=%s", exportResponse.Code, exportResponse.Body.String())
	}

	alertRequest := httptest.NewRequest(http.MethodPost, "/v1/alerts/evaluate",
		bytes.NewBufferString(`{"rule_id":"high-error-rate","samples":[0.01,0.09]}`))
	alertResponse := httptest.NewRecorder()
	mux.ServeHTTP(alertResponse, alertRequest)
	if alertResponse.Code != http.StatusOK {
		t.Fatalf("alert status=%d body=%s", alertResponse.Code, alertResponse.Body.String())
	}
	var alert AlertResult
	if err := json.Unmarshal(alertResponse.Body.Bytes(), &alert); err != nil {
		t.Fatalf("decode alert: %v", err)
	}
	if !alert.Fired {
		t.Fatalf("expected alert fired, got %#v", alert)
	}
}

func TestHandleInfo_OTelInspired(t *testing.T) {
	mux := newMux(NewOTelStore())
	request := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	var info map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode info: %v", err)
	}
	if info["otel_inspired"] != true || info["instrumentation_model"] != "otel-inspired" || info["collector"] != "none" {
		t.Fatalf("inaccurate info labels: %#v", info)
	}
}

func TestHandleDemo_OTelProof(t *testing.T) {
	mux := newMux(NewOTelStore())
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
	if result.Proof["otel_inspired"] != true {
		t.Fatalf("missing otel_inspired proof: %#v", result.Proof)
	}
	for _, key := range []string{"spans_exported", "metrics_exported", "alert_rules"} {
		value, ok := result.Proof[key].(float64)
		if !ok || value < 1 {
			t.Errorf("expected %s >= 1, got %#v", key, result.Proof[key])
		}
	}
}
