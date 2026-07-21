package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// SimResult is the outcome of a load simulation run.
type SimResult struct {
	WorkflowsSimulated int   `json:"workflows_simulated"`
	Accepted           int   `json:"accepted"`
	Rejected           int   `json:"rejected"`
	Delayed            int   `json:"delayed"`
	Backpressure       bool  `json:"backpressure"`
	P99Ms              int64 `json:"p99_ms"`
	QueueDepth         int   `json:"queue_depth"`
}

// ScaleSim simulates ≥1000 in-process workflows with queueing and backpressure (D-03 E-009).
// Does NOT own durable step/signal APIs (H-006), queue idempotency (E-019), or multi-DC (E-010).
type ScaleSim struct {
	mu           sync.Mutex
	capacity     int
	queueDepth   int
	latencies    []int64
	lastResult   *SimResult
	totalSimmed  int
	maxSimulate  int
}

func NewScaleSim(capacity int) *ScaleSim {
	if capacity <= 0 {
		capacity = 256
	}
	return &ScaleSim{
		capacity:    capacity,
		maxSimulate: 50000, // T-5-10 cap
	}
}

// Simulate runs count synthetic workflow admissions against a capped queue.
func (s *ScaleSim) Simulate(count int) (*SimResult, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if count > s.maxSimulate {
		return nil, fmt.Errorf("count exceeds max simulate cap %d", s.maxSimulate)
	}

	accepted := 0
	rejected := 0
	delayed := 0
	latencies := make([]int64, 0, count)
	depth := 0

	for i := 0; i < count; i++ {
		// Synthetic latency grows slightly with depth (SLO signal).
		lat := int64(1 + depth/10)
		if depth >= s.capacity {
			// Saturated: reject half, delay half (backpressure signals).
			if i%2 == 0 {
				rejected++
				latencies = append(latencies, lat+5)
			} else {
				delayed++
				latencies = append(latencies, lat+2)
			}
			continue
		}
		accepted++
		depth++
		latencies = append(latencies, lat)
		// Drain one slot every 4 accepts to keep simulation moving.
		if accepted%4 == 0 && depth > 0 {
			depth--
		}
	}

	p99 := percentile(latencies, 99)
	bp := rejected > 0 || delayed > 0
	res := &SimResult{
		WorkflowsSimulated: count,
		Accepted:           accepted,
		Rejected:           rejected,
		Delayed:            delayed,
		Backpressure:       bp,
		P99Ms:              p99,
		QueueDepth:         depth,
	}
	s.queueDepth = depth
	s.latencies = latencies
	s.lastResult = res
	s.totalSimmed += count
	return res, nil
}

func (s *ScaleSim) Metrics() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	p99 := int64(0)
	if len(s.latencies) > 0 {
		p99 = percentile(s.latencies, 99)
	}
	if s.lastResult != nil {
		p99 = s.lastResult.P99Ms
	}
	return map[string]any{
		"p99_ms":               p99,
		"queue_depth":          s.queueDepth,
		"capacity":             s.capacity,
		"total_simulated":      s.totalSimmed,
		"max_simulate":         s.maxSimulate,
		"workflows_simulated":  s.lastResultSafe(),
	}
}

func (s *ScaleSim) lastResultSafe() int {
	if s.lastResult == nil {
		return 0
	}
	return s.lastResult.WorkflowsSimulated
}

func (s *ScaleSim) Info() map[string]any {
	return map[string]any{
		"requirement_id":   "ENG-E-009",
		"service":          "eng-e-009",
		"title":            "Scale to thousands of application instances and workflows",
		"durable_workflow": false,
		"owns":             []string{"load_simulation", "backpressure", "slo_latency"},
		"does_not_own":     []string{"durable_workflows", "queue_idempotency", "multi_dc"},
		"note":             "In-process load/backpressure simulator; not a durable workflow engine (see ENG-H-006)",
	}
}

func percentile(vals []int64, p int) int64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int64(nil), vals...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	idx := (p * len(cp) / 100) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	_ = time.Now()
	return cp[idx]
}
