package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestSpan_StartEnd_AppearsInExport(t *testing.T) {
	store := NewOTelStore()
	started, err := store.StartSpan("deployment.reconcile")
	if err != nil {
		t.Fatalf("StartSpan: %v", err)
	}
	if _, err := store.EndSpan(started.TraceID); err != nil {
		t.Fatalf("EndSpan: %v", err)
	}
	exported := store.ExportTraces()
	if len(exported.Spans) != 1 {
		t.Fatalf("expected one span, got %#v", exported)
	}
	if exported.Spans[0].Name != "deployment.reconcile" || exported.Spans[0].TraceID == "" {
		t.Fatalf("span lacks name or trace id: %#v", exported.Spans[0])
	}
	if exported.Spans[0].EndedAt.IsZero() {
		t.Fatalf("span was not ended: %#v", exported.Spans[0])
	}
}

func TestAlert_FiresOnThreshold(t *testing.T) {
	result, err := NewOTelStore().EvaluateAlerts("high-error-rate", []float64{0.01, 0.08, 0.12})
	if err != nil {
		t.Fatalf("EvaluateAlerts: %v", err)
	}
	if !result.Fired || result.Status != "alert_fired" {
		t.Fatalf("expected fired alert, got %#v", result)
	}
}

func TestAlert_QuietWhenHealthy(t *testing.T) {
	result, err := NewOTelStore().EvaluateAlerts("high-error-rate", []float64{0.01, 0.02, 0.03})
	if err != nil {
		t.Fatalf("EvaluateAlerts: %v", err)
	}
	if result.Fired || result.Status != "alert_ok" {
		t.Fatalf("expected healthy alert, got %#v", result)
	}
}

func TestOTel_InfoFlags(t *testing.T) {
	info := NewOTelStore().Info()
	if info["otel_inspired"] != true || info["simulator"] != true {
		t.Fatalf("missing honesty flags: %#v", info)
	}
	if info["instrumentation_model"] != "otel-inspired" || info["collector"] != "none" {
		t.Fatalf("unexpected instrumentation claims: %#v", info)
	}
}

func TestOTel_ConcurrentSpans(t *testing.T) {
	store := NewOTelStore()
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			span, err := store.StartSpan(fmt.Sprintf("span-%d", i))
			if err != nil {
				t.Errorf("StartSpan: %v", err)
				return
			}
			_ = store.ExportTraces()
			if _, err := store.EndSpan(span.TraceID); err != nil {
				t.Errorf("EndSpan: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := len(store.ExportTraces().Spans); got != 30 {
		t.Fatalf("expected 30 spans, got %d", got)
	}
}
