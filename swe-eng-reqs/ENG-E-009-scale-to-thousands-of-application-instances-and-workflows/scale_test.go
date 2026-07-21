package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestScale_Simulate_AtLeast1000(t *testing.T) {
	s := NewScaleSim(100)
	res, err := s.Simulate(1000)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	if res.WorkflowsSimulated < 1000 {
		t.Fatalf("expected workflows_simulated >= 1000, got %d", res.WorkflowsSimulated)
	}
}

func TestScale_Backpressure_WhenSaturated(t *testing.T) {
	s := NewScaleSim(10) // tiny queue cap
	res, err := s.Simulate(50)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	if !res.Backpressure {
		t.Fatal("expected backpressure when queue capacity saturated")
	}
	if res.Rejected < 1 && res.Delayed < 1 {
		t.Fatalf("expected reject or delay signal, got rejected=%d delayed=%d", res.Rejected, res.Delayed)
	}
}

func TestScale_SLO_P99Present(t *testing.T) {
	s := NewScaleSim(200)
	_, err := s.Simulate(100)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	m := s.Metrics()
	if _, ok := m["p99_ms"]; !ok {
		t.Fatalf("expected p99_ms in metrics, got %#v", m)
	}
	if _, ok := m["queue_depth"]; !ok {
		t.Fatalf("expected queue_depth in metrics, got %#v", m)
	}
}

func TestScale_NotDurableWorkflow(t *testing.T) {
	s := NewScaleSim(50)
	rt := reflect.TypeOf(s)
	for i := 0; i < rt.NumMethod(); i++ {
		name := rt.Method(i).Name
		lower := strings.ToLower(name)
		if strings.Contains(lower, "signal") || strings.Contains(lower, "durable") || strings.Contains(lower, "replay") {
			t.Fatalf("E-009 must not expose durable workflow APIs; found method %s (H-006 ownership)", name)
		}
	}
	info := s.Info()
	if info["durable_workflow"] == true {
		t.Fatal("must not claim durable_workflow")
	}
}
