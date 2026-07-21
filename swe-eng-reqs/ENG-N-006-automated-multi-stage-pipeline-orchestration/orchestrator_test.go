package main

import (
	"sync"
	"testing"
)

// spyGate counts Evaluate calls and returns a configurable decision.
type spyGate struct {
	mu       sync.Mutex
	decision string
	calls    int
}

func newSpyGate(decision string) *spyGate {
	return &spyGate{decision: decision}
}

func (g *spyGate) set(d string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.decision = d
}

func (g *spyGate) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

func (g *spyGate) Evaluate(sloID string) GateDecision {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls++
	d := g.decision
	if d != "deny" {
		d = "allow"
	}
	return GateDecision{Decision: d, BurnRate: 1.0, Evidence: map[string]any{"slo_id": sloID}}
}

func TestOrch_Tick_AllowAdvancesOneStage(t *testing.T) {
	gate := newSpyGate("allow")
	orch := NewOrchestrator(gate)
	orc, err := orch.Create(nil, "checkout-slo")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if orc.CurrentStage != "dev" {
		t.Fatalf("initial stage=%q, want dev", orc.CurrentStage)
	}
	after, _, err := orch.Tick(orc.ID)
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if after.CurrentStage != "staging" {
		t.Fatalf("after allow Tick stage=%q, want staging", after.CurrentStage)
	}
	if after.StagesAdvanced != 1 {
		t.Fatalf("stages_advanced=%d, want 1", after.StagesAdvanced)
	}
}

func TestOrch_Tick_DenyBlocks(t *testing.T) {
	gate := newSpyGate("deny")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	after, dec, err := orch.Tick(orc.ID)
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if dec.Decision != "deny" {
		t.Fatalf("decision=%q, want deny", dec.Decision)
	}
	if after.CurrentStage != "dev" {
		t.Fatalf("deny advanced to %q, want stay dev", after.CurrentStage)
	}
	if after.StagesAdvanced != 0 {
		t.Fatalf("stages_advanced=%d, want 0 on deny", after.StagesAdvanced)
	}
	if !after.Blocked {
		t.Fatal("expected Blocked=true on deny")
	}
}

func TestOrch_Tick_AllowAfterDeny_Resumes(t *testing.T) {
	gate := newSpyGate("deny")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	if _, _, err := orch.Tick(orc.ID); err != nil {
		t.Fatalf("deny Tick: %v", err)
	}
	gate.set("allow")
	after, _, err := orch.Tick(orc.ID)
	if err != nil {
		t.Fatalf("allow Tick: %v", err)
	}
	if after.CurrentStage != "staging" {
		t.Fatalf("after resume stage=%q, want staging", after.CurrentStage)
	}
	if after.Blocked {
		t.Fatal("expected Blocked=false after resume")
	}
}

func TestOrch_TerminalStage_Stops(t *testing.T) {
	gate := newSpyGate("allow")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	// advance to terminal (dev->staging->prod)
	if _, _, err := orch.Tick(orc.ID); err != nil {
		t.Fatalf("Tick1: %v", err)
	}
	if _, _, err := orch.Tick(orc.ID); err != nil {
		t.Fatalf("Tick2: %v", err)
	}
	after, _, err := orch.Tick(orc.ID)
	if err != nil {
		t.Fatalf("terminal Tick: %v", err)
	}
	if after.CurrentStage != "prod" {
		t.Fatalf("terminal stage=%q, want prod", after.CurrentStage)
	}
	if after.StagesAdvanced != 2 {
		t.Fatalf("stages_advanced=%d, want 2 (dev->staging->prod)", after.StagesAdvanced)
	}
}

func TestOrch_GateRequired(t *testing.T) {
	gate := newSpyGate("allow")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	if _, _, err := orch.Tick(orc.ID); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if gate.count() < 1 {
		t.Fatalf("gate Evaluate calls=%d, want >=1 (gate_required)", gate.count())
	}
}

func TestOrch_NoCIStageNames(t *testing.T) {
	gate := newSpyGate("allow")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	ci := map[string]bool{"lint": true, "unit": true, "build": true, "publish": true}
	for _, s := range orc.Stages {
		if ci[s] {
			t.Fatalf("default stage %q is a CI stage name; N-006 must use release/env stages (D-09)", s)
		}
	}
}

func TestOrch_ConcurrentTick(t *testing.T) {
	gate := newSpyGate("allow")
	orch := NewOrchestrator(gate)
	orc, _ := orch.Create(nil, "slo")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = orch.Tick(orc.ID)
			_, _ = orch.Get(orc.ID)
		}()
	}
	wg.Wait()
	got, err := orch.Get(orc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.CurrentStage != "prod" {
		t.Fatalf("after concurrent allow ticks stage=%q, want prod", got.CurrentStage)
	}
}
