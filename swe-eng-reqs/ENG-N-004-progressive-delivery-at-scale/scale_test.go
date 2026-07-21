package main

import (
	"sync"
	"testing"
)

var passingMetrics = Metrics{ErrorRate: 0.001, SuccessRate: 0.999}
var failingMetrics = Metrics{ErrorRate: 0.20, SuccessRate: 0.80}
var strictCriteria = Criteria{MaxErrorRate: 0.01, MinSuccess: 0.99}

func TestPlan_Create_RequiresTwoEnvs(t *testing.T) {
	s := NewScaleStore()
	if _, err := s.CreatePlan([]string{"dev"}, strictCriteria); err == nil {
		t.Fatal("CreatePlan with 1 env: want error")
	}
	p, err := s.CreatePlan([]string{"dev", "staging", "prod"}, strictCriteria)
	if err != nil {
		t.Fatalf("CreatePlan with 3 envs: %v", err)
	}
	if len(p.Envs) != 3 {
		t.Fatalf("envs=%d, want 3", len(p.Envs))
	}
	if p.CurrentEnv != "dev" {
		t.Fatalf("current env=%q, want dev", p.CurrentEnv)
	}
}

func TestCriteria_Fail_BlocksPromote(t *testing.T) {
	s := NewScaleStore()
	p, _ := s.CreatePlan([]string{"dev", "staging"}, strictCriteria)
	ev, err := s.Evaluate(p.ID, failingMetrics)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if ev.CriteriaPassed {
		t.Fatal("criteria_passed=true on failing metrics, want false")
	}
	if _, err := s.Promote(p.ID); err == nil {
		t.Fatal("Promote after failed criteria: want error")
	}
	got, _ := s.Get(p.ID)
	if got.CurrentEnv != "dev" {
		t.Fatalf("env advanced to %q despite failed criteria, want dev", got.CurrentEnv)
	}
}

func TestCriteria_Pass_AutoPromote(t *testing.T) {
	s := NewScaleStore()
	p, _ := s.CreatePlan([]string{"dev", "staging", "prod"}, strictCriteria)
	ev, err := s.Evaluate(p.ID, passingMetrics)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !ev.CriteriaPassed {
		t.Fatal("criteria_passed=false on passing metrics, want true")
	}
	promoted, err := s.Promote(p.ID)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if promoted.CurrentEnv != "staging" {
		t.Fatalf("after auto-promote env=%q, want staging", promoted.CurrentEnv)
	}
	if promoted.EnvStatus["dev"] != "promoted" {
		t.Fatalf("dev status=%q, want promoted", promoted.EnvStatus["dev"])
	}
}

func TestPlan_PerEnvStatus(t *testing.T) {
	s := NewScaleStore()
	p, _ := s.CreatePlan([]string{"dev", "staging"}, strictCriteria)
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.EnvStatus) != 2 {
		t.Fatalf("env_status entries=%d, want 2", len(got.EnvStatus))
	}
	if got.EnvStatus["dev"] != "active" {
		t.Fatalf("dev status=%q, want active", got.EnvStatus["dev"])
	}
	if got.EnvStatus["staging"] != "pending" {
		t.Fatalf("staging status=%q, want pending", got.EnvStatus["staging"])
	}
}

func TestScale_ConcurrentEvaluate(t *testing.T) {
	s := NewScaleStore()
	p, _ := s.CreatePlan([]string{"dev", "staging", "prod"}, strictCriteria)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Evaluate(p.ID, passingMetrics)
			_, _ = s.Promote(p.ID)
			_, _ = s.Get(p.ID)
		}()
	}
	wg.Wait()
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	// must never advance past the final environment
	if got.CurrentEnvIndex > len(got.Envs)-1 {
		t.Fatalf("current index=%d out of range for %d envs", got.CurrentEnvIndex, len(got.Envs))
	}
}
