package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// defaultStages are release/environment stages (D-09). ENG-N-006 orchestrates
// promotion through environments — it is NOT a CI DAG and never uses CI job
// names (lint/unit/build/publish belong to ENG-E-001/ENG-E-023).
var defaultStages = []string{"dev", "staging", "prod"}

// GateDecision mirrors the ENG-N-005 GateDecision shape but is defined locally.
// N-006 embeds a GateEvaluator stub and does NOT HTTP-call N-005 (D-05, D-11).
type GateDecision struct {
	Decision string         `json:"decision"` // allow|deny
	BurnRate float64        `json:"burn_rate"`
	Evidence map[string]any `json:"evidence"`
}

// GateEvaluator decides whether a release stage may advance.
type GateEvaluator interface {
	Evaluate(sloID string) GateDecision
}

// StubGate is an in-folder gate stub. It records call counts so the demo can
// prove that every advance path consults the gate (gate_required).
type StubGate struct {
	mu       sync.Mutex
	decision string
	calls    int
}

func NewStubGate(decision string) *StubGate {
	if decision != "deny" {
		decision = "allow"
	}
	return &StubGate{decision: decision}
}

// SetDecision flips the stub between "allow" and "deny".
func (g *StubGate) SetDecision(d string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if d != "deny" {
		d = "allow"
	}
	g.decision = d
}

// Calls returns the number of Evaluate invocations observed.
func (g *StubGate) Calls() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

func (g *StubGate) Evaluate(sloID string) GateDecision {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls++
	decision := g.decision
	burn := 0.4
	if decision == "deny" {
		burn = 2.5
	}
	return GateDecision{
		Decision: decision,
		BurnRate: burn,
		Evidence: map[string]any{"slo_id": sloID, "stub": true},
	}
}

// Orchestration is a single release advancing through environment stages.
type Orchestration struct {
	ID             string   `json:"id"`
	Stages         []string `json:"stages"`
	CurrentIndex   int      `json:"current_index"`
	CurrentStage   string   `json:"current_stage"`
	SLOID          string   `json:"slo_id"`
	LastDecision   string   `json:"last_decision"`
	StagesAdvanced int      `json:"stages_advanced"`
	Blocked        bool     `json:"blocked"`
}

// Orchestrator advances release stages only when the gate returns allow.
type Orchestrator struct {
	mu    sync.Mutex
	items map[string]*Orchestration
	gate  GateEvaluator
	seq   int
}

func NewOrchestrator(gate GateEvaluator) *Orchestrator {
	return &Orchestrator{
		items: make(map[string]*Orchestration),
		gate:  gate,
	}
}

// Create registers a new orchestration. Empty stages default to release/env
// stages; caller-supplied stages are validated as safe identifiers.
func (o *Orchestrator) Create(stages []string, sloID string) (*Orchestration, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(stages) == 0 {
		stages = append([]string(nil), defaultStages...)
	}
	for _, s := range stages {
		if !safeStageID(s) {
			return nil, fmt.Errorf("invalid stage id %q", s)
		}
	}
	if sloID == "" {
		sloID = "default-slo"
	}
	if !safeStageID(sloID) {
		return nil, fmt.Errorf("invalid slo id %q", sloID)
	}

	o.seq++
	id := fmt.Sprintf("orch-%d", o.seq)
	orc := &Orchestration{
		ID:           id,
		Stages:       append([]string(nil), stages...),
		CurrentIndex: 0,
		CurrentStage: stages[0],
		SLOID:        sloID,
	}
	o.items[id] = orc
	return copyOrch(orc), nil
}

// Tick consults the gate, then advances exactly one stage only when the gate
// returns allow. The gate is ALWAYS evaluated (gate_required), including at the
// terminal stage. A deny blocks advancement and marks the orchestration blocked.
func (o *Orchestrator) Tick(id string) (*Orchestration, GateDecision, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	orc, ok := o.items[id]
	if !ok {
		return nil, GateDecision{}, errors.New("orchestration not found")
	}

	decision := o.gate.Evaluate(orc.SLOID)
	orc.LastDecision = decision.Decision

	if orc.CurrentIndex >= len(orc.Stages)-1 {
		orc.Blocked = false
		return copyOrch(orc), decision, nil
	}
	if decision.Decision != "allow" {
		orc.Blocked = true
		return copyOrch(orc), decision, nil
	}

	orc.CurrentIndex++
	orc.CurrentStage = orc.Stages[orc.CurrentIndex]
	orc.StagesAdvanced++
	orc.Blocked = false
	return copyOrch(orc), decision, nil
}

// Get returns a copy of the orchestration state.
func (o *Orchestrator) Get(id string) (*Orchestration, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	orc, ok := o.items[id]
	if !ok {
		return nil, errors.New("orchestration not found")
	}
	return copyOrch(orc), nil
}

func copyOrch(o *Orchestration) *Orchestration {
	cp := *o
	cp.Stages = append([]string(nil), o.Stages...)
	return &cp
}

// safeStageID rejects path traversal and enforces a conservative alphabet.
func safeStageID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
			return false
		}
	}
	return true
}
