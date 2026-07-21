package main

import (
	"testing"
	"time"
)

func TestWorkflow_Start_PersistsSteps(t *testing.T) {
	eng := NewWorkflowEngine()
	steps := []string{"validate", "provision", "run", "finalize"}
	wf, err := eng.Start("demo-wf", steps)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if wf.ID == "" {
		t.Fatal("expected workflow id")
	}
	if !wf.Durable {
		t.Fatal("expected durable=true")
	}
	if len(wf.Steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(wf.Steps))
	}
	if wf.StepsCompleted != 0 {
		t.Fatalf("expected 0 steps completed at start, got %d", wf.StepsCompleted)
	}

	// Signal advances state; Get must return same durable record (not ephemeral).
	if _, err := eng.Signal(wf.ID, "advance", "evt-1"); err != nil {
		t.Fatalf("Signal: %v", err)
	}
	got, ok := eng.Get(wf.ID)
	if !ok {
		t.Fatal("workflow disappeared after Signal — not durable")
	}
	if got.StepsCompleted < 1 {
		t.Fatalf("expected persisted steps_completed >= 1, got %d", got.StepsCompleted)
	}
}

func TestWorkflow_Signal_AdvancesStep(t *testing.T) {
	eng := NewWorkflowEngine()
	wf, err := eng.Start("advance-wf", []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	before := wf.StepsCompleted
	out, err := eng.Signal(wf.ID, "advance", "sig-1")
	if err != nil {
		t.Fatalf("Signal: %v", err)
	}
	if out.StepsCompleted != before+1 {
		t.Fatalf("steps_completed want %d, got %d", before+1, out.StepsCompleted)
	}
	out2, err := eng.Signal(wf.ID, "advance", "sig-2")
	if err != nil {
		t.Fatalf("Signal 2: %v", err)
	}
	if out2.StepsCompleted != before+2 {
		t.Fatalf("steps_completed want %d, got %d", before+2, out2.StepsCompleted)
	}
}

func TestWorkflow_ReplaySafe(t *testing.T) {
	eng := NewWorkflowEngine()
	wf, err := eng.Start("replay-wf", []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	first, err := eng.Signal(wf.ID, "advance", "same-event-id")
	if err != nil {
		t.Fatalf("Signal: %v", err)
	}
	second, err := eng.Signal(wf.ID, "advance", "same-event-id")
	if err != nil {
		t.Fatalf("Replay Signal: %v", err)
	}
	if second.StepsCompleted != first.StepsCompleted {
		t.Fatalf("replay double-applied: first=%d second=%d", first.StepsCompleted, second.StepsCompleted)
	}
	if !second.ReplaySafe {
		t.Fatal("expected replay_safe=true after idempotent replay")
	}
}

func TestWorkflow_Throughput_Positive(t *testing.T) {
	eng := NewWorkflowEngine()
	const n = 50
	start := time.Now()
	for i := 0; i < n; i++ {
		id := "batch-" + string(rune('A'+i%26)) + "-" + itoa(i)
		wf, err := eng.Start(id, []string{"s1", "s2"})
		if err != nil {
			t.Fatalf("Start %d: %v", i, err)
		}
		if _, err := eng.Signal(wf.ID, "advance", "e1-"+itoa(i)); err != nil {
			t.Fatalf("Signal %d: %v", i, err)
		}
		if _, err := eng.Signal(wf.ID, "advance", "e2-"+itoa(i)); err != nil {
			t.Fatalf("Signal2 %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	tp := eng.Throughput()
	if tp.ThroughputPerS <= 0 {
		t.Fatalf("expected throughput_per_s > 0, got %v (elapsed %v)", tp.ThroughputPerS, elapsed)
	}
	if tp.SignalsApplied < n*2 {
		t.Fatalf("expected signals_applied >= %d, got %d", n*2, tp.SignalsApplied)
	}
}

func TestWorkflow_NotLoadOnlySim(t *testing.T) {
	eng := NewWorkflowEngine()
	// Load-only sims expose Simulate(N) without durable multi-step state.
	// H-006 must own durable Start/Signal — not a bare Simulate API.
	type loadOnly interface {
		Simulate(n int) map[string]any
	}
	if _, ok := any(eng).(loadOnly); ok {
		t.Fatal("WorkflowEngine must not be a load-only Simulate(N) API (E-009 owns that)")
	}
	wf, err := eng.Start("owned-wf", []string{"prepare", "execute", "cleanup"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !wf.Durable {
		t.Fatal("durable workflow required — distinct from E-009 load sim")
	}
	if _, err := eng.Signal(wf.ID, "advance", "owned-1"); err != nil {
		t.Fatalf("Signal: %v", err)
	}
	got, ok := eng.Get(wf.ID)
	if !ok || got.StepsCompleted < 1 {
		t.Fatal("durable state missing after signal — looks like load-only sim")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
