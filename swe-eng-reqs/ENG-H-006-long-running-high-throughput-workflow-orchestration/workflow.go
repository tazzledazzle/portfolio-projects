package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrWorkflowNotFound is returned when Signal/Get targets an unknown id.
var ErrWorkflowNotFound = errors.New("workflow not found")

// ErrWorkflowComplete is returned when no further steps remain.
var ErrWorkflowComplete = errors.New("workflow already complete")

// Workflow is a durable multi-step orchestration record (ENG-H-006).
// Distinct from E-009 load-only Simulate and I-009 GPU chunk upload.
type Workflow struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Steps          []string `json:"steps"`
	CurrentStep    int      `json:"current_step"`
	StepsCompleted int      `json:"steps_completed"`
	Status         string   `json:"status"`
	Durable        bool     `json:"durable"`
	ReplaySafe     bool     `json:"replay_safe"`
	AppliedEvents  []string `json:"applied_events,omitempty"`
}

// ThroughputStats reports engine throughput after signal batch work.
type ThroughputStats struct {
	ThroughputPerS float64 `json:"throughput_per_s"`
	SignalsApplied int64   `json:"signals_applied"`
	Workflows      int     `json:"workflows"`
	ElapsedMS      int64   `json:"elapsed_ms"`
}

// WorkflowEngine is an in-memory durable workflow MVP (stdlib mutex store).
// No Temporal SDK — D-02.
type WorkflowEngine struct {
	mu             sync.Mutex
	workflows      map[string]*Workflow
	applied        map[string]struct{} // key: workflowID|eventID
	signalsApplied int64
	startedAt      time.Time
	seq            int
}

// NewWorkflowEngine creates an empty durable workflow engine.
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]*Workflow),
		applied:   make(map[string]struct{}),
		startedAt: time.Now(),
	}
}

// Start creates a durable multi-step workflow. State survives subsequent Signals.
func (e *WorkflowEngine) Start(name string, steps []string) (*Workflow, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	if len(steps) == 0 {
		return nil, errors.New("steps required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seq++
	id := fmt.Sprintf("wf-%d", e.seq)
	wf := &Workflow{
		ID:             id,
		Name:           name,
		Steps:          append([]string(nil), steps...),
		CurrentStep:    0,
		StepsCompleted: 0,
		Status:         "running",
		Durable:        true,
		ReplaySafe:     true,
		AppliedEvents:  nil,
	}
	e.workflows[id] = wf
	return cloneWorkflow(wf), nil
}

// Signal advances the workflow by one step. Replaying the same eventID is a no-op (replay_safe).
func (e *WorkflowEngine) Signal(id, action, eventID string) (*Workflow, error) {
	if id == "" {
		return nil, errors.New("id required")
	}
	if eventID == "" {
		return nil, errors.New("event_id required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	wf, ok := e.workflows[id]
	if !ok {
		return nil, ErrWorkflowNotFound
	}
	key := id + "|" + eventID
	if _, seen := e.applied[key]; seen {
		wf.ReplaySafe = true
		return cloneWorkflow(wf), nil
	}
	if action == "" {
		action = "advance"
	}
	if wf.StepsCompleted >= len(wf.Steps) {
		wf.Status = "completed"
		return nil, ErrWorkflowComplete
	}
	wf.StepsCompleted++
	wf.CurrentStep = wf.StepsCompleted
	e.applied[key] = struct{}{}
	e.signalsApplied++
	wf.AppliedEvents = append(wf.AppliedEvents, eventID)
	wf.ReplaySafe = true
	if wf.StepsCompleted >= len(wf.Steps) {
		wf.Status = "completed"
	}
	return cloneWorkflow(wf), nil
}

// Get returns a durable workflow snapshot.
func (e *WorkflowEngine) Get(id string) (*Workflow, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	wf, ok := e.workflows[id]
	if !ok {
		return nil, false
	}
	return cloneWorkflow(wf), true
}

// Throughput returns signals/sec since engine creation (positive after batch work).
func (e *WorkflowEngine) Throughput() ThroughputStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	elapsed := time.Since(e.startedAt)
	ms := elapsed.Milliseconds()
	if ms < 1 {
		ms = 1
	}
	perS := float64(e.signalsApplied) / (float64(ms) / 1000.0)
	if e.signalsApplied > 0 && perS <= 0 {
		perS = float64(e.signalsApplied) // pathological clock floor
	}
	return ThroughputStats{
		ThroughputPerS: perS,
		SignalsApplied: e.signalsApplied,
		Workflows:      len(e.workflows),
		ElapsedMS:      ms,
	}
}

func cloneWorkflow(wf *Workflow) *Workflow {
	cp := *wf
	cp.Steps = append([]string(nil), wf.Steps...)
	cp.AppliedEvents = append([]string(nil), wf.AppliedEvents...)
	return &cp
}
