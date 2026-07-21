package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ENG-N-004 owns MULTI-ENVIRONMENT progressive delivery: a plan promotes a
// workload across an ordered list of environments, advancing to the next env
// only when automated promotion criteria pass (D-03). It does NOT own
// single-canary weight internals (ENG-E-024), CI DAGs, or finalizers.

// Criteria are the automated promotion thresholds evaluated per environment.
type Criteria struct {
	MaxErrorRate float64 `json:"max_error_rate"`
	MinSuccess   float64 `json:"min_success"`
}

// Metrics are observed signals for the current environment. Criteria are
// evaluated server-side against these values — never trusting a client-supplied
// pass/fail verdict.
type Metrics struct {
	ErrorRate   float64 `json:"error_rate"`
	SuccessRate float64 `json:"success_rate"`
}

// Plan is a multi-environment progressive delivery plan.
type Plan struct {
	ID              string            `json:"id"`
	Envs            []string          `json:"envs"`
	Criteria        Criteria          `json:"criteria"`
	CriteriaPassed  bool              `json:"criteria_passed"`
	CurrentEnvIndex int               `json:"current_env_index"`
	CurrentEnv      string            `json:"current_env"`
	EnvStatus       map[string]string `json:"env_status"`
}

// ScaleStore holds progressive delivery plans.
type ScaleStore struct {
	mu    sync.Mutex
	plans map[string]*Plan
	seq   int
}

func NewScaleStore() *ScaleStore {
	return &ScaleStore{plans: make(map[string]*Plan)}
}

// CreatePlan requires at least two environments (multi-env is the whole point).
func (s *ScaleStore) CreatePlan(envs []string, crit Criteria) (*Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(envs) < 2 {
		return nil, errors.New("plan requires at least 2 environments")
	}
	for _, e := range envs {
		if !safeEnvID(e) {
			return nil, fmt.Errorf("invalid environment id %q", e)
		}
	}
	if crit.MaxErrorRate <= 0 {
		crit.MaxErrorRate = 0.01
	}
	if crit.MinSuccess <= 0 {
		crit.MinSuccess = 0.99
	}

	s.seq++
	id := fmt.Sprintf("plan-%d", s.seq)
	status := make(map[string]string, len(envs))
	for i, e := range envs {
		if i == 0 {
			status[e] = "active"
		} else {
			status[e] = "pending"
		}
	}
	plan := &Plan{
		ID:              id,
		Envs:            append([]string(nil), envs...),
		Criteria:        crit,
		CurrentEnvIndex: 0,
		CurrentEnv:      envs[0],
		EnvStatus:       status,
	}
	s.plans[id] = plan
	return copyPlan(plan), nil
}

// Evaluate scores observed metrics against the plan criteria and records the
// result. Promotion is gated on this server-computed verdict.
func (s *ScaleStore) Evaluate(id string, m Metrics) (*Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, ok := s.plans[id]
	if !ok {
		return nil, errors.New("plan not found")
	}
	plan.CriteriaPassed = m.ErrorRate <= plan.Criteria.MaxErrorRate &&
		m.SuccessRate >= plan.Criteria.MinSuccess
	return copyPlan(plan), nil
}

// Promote advances to the next environment only when criteria passed. After a
// successful promotion the criteria verdict resets — the next environment must
// be evaluated on its own merits.
func (s *ScaleStore) Promote(id string) (*Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, ok := s.plans[id]
	if !ok {
		return nil, errors.New("plan not found")
	}
	if !plan.CriteriaPassed {
		return nil, errors.New("criteria not passed; promotion blocked")
	}
	if plan.CurrentEnvIndex >= len(plan.Envs)-1 {
		return nil, errors.New("already at final environment; cannot promote")
	}
	prev := plan.Envs[plan.CurrentEnvIndex]
	plan.EnvStatus[prev] = "promoted"
	plan.CurrentEnvIndex++
	plan.CurrentEnv = plan.Envs[plan.CurrentEnvIndex]
	plan.EnvStatus[plan.CurrentEnv] = "active"
	plan.CriteriaPassed = false
	return copyPlan(plan), nil
}

func (s *ScaleStore) Get(id string) (*Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[id]
	if !ok {
		return nil, errors.New("plan not found")
	}
	return copyPlan(plan), nil
}

func copyPlan(p *Plan) *Plan {
	cp := *p
	cp.Envs = append([]string(nil), p.Envs...)
	cp.EnvStatus = make(map[string]string, len(p.EnvStatus))
	for k, v := range p.EnvStatus {
		cp.EnvStatus[k] = v
	}
	return &cp
}

func safeEnvID(id string) bool {
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
