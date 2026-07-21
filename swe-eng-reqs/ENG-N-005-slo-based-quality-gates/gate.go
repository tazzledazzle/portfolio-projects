package main

import (
	"errors"
	"fmt"
	"sync"
)

// SeriesSample holds error/total counts for short and long windows.
type SeriesSample struct {
	ErrorsShort float64 `json:"errors_short"`
	TotalShort  float64 `json:"total_short"`
	ErrorsLong  float64 `json:"errors_long"`
	TotalLong   float64 `json:"total_long"`
}

// SLO defines objective and burn-rate threshold.
type SLO struct {
	ID        string  `json:"id"`
	Objective float64 `json:"objective"`
	Threshold float64 `json:"threshold"`
}

// GateDecision is allow|deny with server-computed evidence.
type GateDecision struct {
	Decision string         `json:"decision"`
	BurnRate float64        `json:"burn_rate"`
	Evidence map[string]any `json:"evidence"`
}

type sloState struct {
	slo    SLO
	series SeriesSample
	has    bool
}

// GateEngine evaluates PromQL-inspired multi-window burn-rate gates (simulator).
type GateEngine struct {
	mu    sync.Mutex
	slos  map[string]*sloState
}

func NewGateEngine() *GateEngine {
	return &GateEngine{slos: make(map[string]*sloState)}
}

func (e *GateEngine) PutSLO(id string, objective, threshold float64) error {
	if id == "" {
		return errors.New("slo id required")
	}
	if objective <= 0 || objective >= 1 {
		return errors.New("objective must be in (0,1)")
	}
	if threshold <= 0 {
		return errors.New("threshold must be > 0")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	st, ok := e.slos[id]
	if !ok {
		st = &sloState{}
		e.slos[id] = st
	}
	st.slo = SLO{ID: id, Objective: objective, Threshold: threshold}
	return nil
}

func (e *GateEngine) IngestSeries(id string, sample SeriesSample) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	st, ok := e.slos[id]
	if !ok {
		return fmt.Errorf("unknown slo %q", id)
	}
	if sample.TotalShort <= 0 || sample.TotalLong <= 0 {
		return errors.New("total_short and total_long must be > 0")
	}
	st.series = sample
	st.has = true
	return nil
}

func (e *GateEngine) Evaluate(id string) (GateDecision, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	st, ok := e.slos[id]
	if !ok {
		return GateDecision{}, fmt.Errorf("unknown slo %q", id)
	}
	if !st.has {
		return GateDecision{}, errors.New("no series ingested for slo")
	}

	decision, burnShort, burnLong := evaluateBurn(
		st.slo.Objective,
		st.series.ErrorsShort, st.series.TotalShort,
		st.series.ErrorsLong, st.series.TotalLong,
		st.slo.Threshold,
	)
	// Representational burn_rate: max of windows for API consumers
	br := burnShort
	if burnLong > br {
		br = burnLong
	}
	return GateDecision{
		Decision: decision,
		BurnRate: br,
		Evidence: map[string]any{
			"burn_short":   burnShort,
			"burn_long":    burnLong,
			"threshold":    st.slo.Threshold,
			"objective":    st.slo.Objective,
			"errors_short": st.series.ErrorsShort,
			"total_short":  st.series.TotalShort,
			"errors_long":  st.series.ErrorsLong,
			"total_long":   st.series.TotalLong,
			"multi_window": "AND",
		},
	}, nil
}

func (e *GateEngine) Info() map[string]any {
	return map[string]any{
		"requirement_id":  "ENG-N-005",
		"service":         "eng-n-005",
		"title":           "SLO-based quality gates",
		"promql_inspired": true,
		"simulator":       true,
		"note":            "Does NOT connect to Prometheus",
	}
}

// burnRate = error_ratio / (1 - objective)
func burnRate(errorRatio, objective float64) float64 {
	budget := 1 - objective
	if budget <= 0 {
		return 0
	}
	return errorRatio / budget
}

// evaluateBurn implements multi-window AND deny (SRE Workbook pattern).
func evaluateBurn(objective, errShort, totShort, errLong, totLong, threshold float64) (decision string, burnShort, burnLong float64) {
	burnShort = burnRate(errShort/totShort, objective)
	burnLong = burnRate(errLong/totLong, objective)
	if burnShort > threshold && burnLong > threshold {
		return "deny", burnShort, burnLong
	}
	return "allow", burnShort, burnLong
}
