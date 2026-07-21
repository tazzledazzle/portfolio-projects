package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDAGValidate_LinearDAG(t *testing.T) {
	dag := map[string][]string{
		"lint":  {},
		"test":  {"lint"},
		"build": {"test"},
	}
	p := NewPipeline(dag)
	if err := p.Validate(); err != nil {
		t.Errorf("linear DAG should be valid, got error: %v", err)
	}
}

func TestDAGValidate_CyclicDAG(t *testing.T) {
	dag := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	p := NewPipeline(dag)
	err := p.Validate()
	if err == nil {
		t.Error("cyclic DAG should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("error should contain 'cycle detected', got: %v", err)
	}
}

func TestDAGValidate_MissingDependency(t *testing.T) {
	dag := map[string][]string{
		"test": {"nonexistent"},
	}
	p := NewPipeline(dag)
	err := p.Validate()
	if err == nil {
		t.Error("missing dependency should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "stage not found") {
		t.Errorf("error should contain 'stage not found', got: %v", err)
	}
}

func TestDAGValidate_EmptyDAG(t *testing.T) {
	dag := map[string][]string{}
	p := NewPipeline(dag)
	err := p.Validate()
	if err == nil {
		t.Error("empty DAG should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "empty DAG") {
		t.Errorf("error should contain 'empty DAG', got: %v", err)
	}
}

func TestTransitionStage_ValidTransition(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		wantError bool
	}{
		{"pending to running", "pending", "running", false},
		{"running to succeeded", "running", "succeeded", false},
		{"running to failed", "running", "failed", false},
		{"failed to running (retry)", "failed", "running", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dag := map[string][]string{"stage1": {}}
			p := NewPipeline(dag)
			p.Stages["stage1"].Status = tt.from

			err := p.TransitionStage("stage1", tt.to)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && p.Stages["stage1"].Status != tt.to {
				t.Errorf("status should be %q, got %q", tt.to, p.Stages["stage1"].Status)
			}
		})
	}
}

func TestTransitionStage_InvalidTransition(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"running to pending", "running", "pending"},
		{"succeeded to running", "succeeded", "running"},
		{"succeeded to pending", "succeeded", "pending"},
		{"pending to succeeded", "pending", "succeeded"},
		{"pending to failed", "pending", "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dag := map[string][]string{"stage1": {}}
			p := NewPipeline(dag)
			p.Stages["stage1"].Status = tt.from

			err := p.TransitionStage("stage1", tt.to)
			if err == nil {
				t.Error("expected error for invalid transition")
			}
			if err != nil && !strings.Contains(err.Error(), "invalid transition") {
				t.Errorf("error should contain 'invalid transition', got: %v", err)
			}
		})
	}
}

func TestTransitionStage_ConcurrentAccess(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dag := map[string][]string{"stage1": {}}
			p := NewPipeline(dag)
			_ = p.TransitionStage("stage1", "running")
			_ = p.TransitionStage("stage1", "succeeded")
		}()
	}

	wg.Wait()

	dag := map[string][]string{"stage1": {}, "stage2": {}, "stage3": {}}
	p := NewPipeline(dag)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			stages := []string{"stage1", "stage2", "stage3"}
			stage := stages[idx%len(stages)]
			_ = p.TransitionStage(stage, "running")
		}(i)
	}
	wg.Wait()
}

func TestTransitionStage_StageNotFound(t *testing.T) {
	dag := map[string][]string{"stage1": {}}
	p := NewPipeline(dag)

	err := p.TransitionStage("nonexistent", "running")
	if err == nil {
		t.Error("expected error for nonexistent stage")
	}
}

func TestRetryPolicy_ExponentialBackoff(t *testing.T) {
	rp := NewRetryPolicy(100*time.Millisecond, 3, 0.0)

	d1 := rp.NextBackoff(1)
	d2 := rp.NextBackoff(2)
	d3 := rp.NextBackoff(3)

	tolerance := 0.2

	expected1 := 100 * time.Millisecond
	if float64(d1) < float64(expected1)*(1-tolerance) || float64(d1) > float64(expected1)*(1+tolerance) {
		t.Errorf("1st retry delay should be ~%v (±20%%), got %v", expected1, d1)
	}

	expected2 := 200 * time.Millisecond
	if float64(d2) < float64(expected2)*(1-tolerance) || float64(d2) > float64(expected2)*(1+tolerance) {
		t.Errorf("2nd retry delay should be ~%v (±20%%), got %v", expected2, d2)
	}

	expected3 := 400 * time.Millisecond
	if float64(d3) < float64(expected3)*(1-tolerance) || float64(d3) > float64(expected3)*(1+tolerance) {
		t.Errorf("3rd retry delay should be ~%v (±20%%), got %v", expected3, d3)
	}
}

func TestRetryPolicy_MaxRetries(t *testing.T) {
	rp := NewRetryPolicy(100*time.Millisecond, 3, 0.0)

	for i := 1; i <= 3; i++ {
		if err := rp.ShouldRetry(i); err != nil {
			t.Errorf("attempt %d should be allowed, got error: %v", i, err)
		}
	}

	err := rp.ShouldRetry(4)
	if err == nil {
		t.Error("attempt 4 should exceed max retries")
	}
	if err != nil && !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("error should contain 'max retries exceeded', got: %v", err)
	}
}

func TestRetryPolicy_JitterRange(t *testing.T) {
	rp := NewRetryPolicy(100*time.Millisecond, 3, 0.1)

	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		d := rp.NextBackoff(1)
		delays[d] = true
	}

	if len(delays) < 2 {
		t.Error("jitter should produce non-deterministic delays")
	}
}
