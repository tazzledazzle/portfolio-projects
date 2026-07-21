package main

import (
	"encoding/json"
	"testing"
)

func TestService_RequestID_Propagated(t *testing.T) {
	rt := NewServiceRuntime()
	id := rt.WithRequestID("")
	if id == "" {
		t.Fatal("WithRequestID must return a non-empty request_id when none provided")
	}
	id2 := rt.WithRequestID("client-req-1")
	if id2 != "client-req-1" {
		t.Fatalf("WithRequestID must preserve client id, got %q", id2)
	}
}

func TestService_StructuredError_JSONShape(t *testing.T) {
	se := StructuredError{Code: "bad_request", Message: "invalid echo payload"}
	b, err := json.Marshal(se)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["code"] != "bad_request" {
		t.Fatalf("expected code field, got %#v", m)
	}
	if m["message"] != "invalid echo payload" {
		t.Fatalf("expected message field, got %#v", m)
	}
	if _, ok := m["stack"]; ok {
		t.Fatal("structured errors must not leak stack traces")
	}
	if _, ok := m["secret"]; ok {
		t.Fatal("structured errors must not include secrets")
	}
}

func TestService_Metrics_Increment(t *testing.T) {
	rt := NewServiceRuntime()
	before := rt.MetricsCount()
	rt.IncrementEcho()
	rt.IncrementDemo()
	after := rt.MetricsCount()
	if after != before+2 {
		t.Fatalf("metrics counter should increment on echo/demo, before=%d after=%d", before, after)
	}
}
