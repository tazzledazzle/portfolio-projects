package main

import (
	"reflect"
	"sync"
	"testing"
)

func TestCanary_Start_DefaultWeights(t *testing.T) {
	store := NewCanaryStore()
	c, err := store.Start("svc-a", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	want := []int{0, 10, 50, 100}
	if !reflect.DeepEqual(c.Weights, want) {
		t.Fatalf("weights=%v, want %v", c.Weights, want)
	}
	if c.Status != "running" {
		t.Fatalf("status=%q, want running", c.Status)
	}
	if c.Weight != 0 {
		t.Fatalf("weight=%d, want 0", c.Weight)
	}
}

func TestCanary_Step_AdvancesWeight(t *testing.T) {
	store := NewCanaryStore()
	c, err := store.Start("svc-b", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	s1, err := store.Step(c.ID)
	if err != nil {
		t.Fatalf("Step: %v", err)
	}
	if s1.Weight != 10 {
		t.Fatalf("weight=%d, want 10", s1.Weight)
	}
	s2, err := store.Step(c.ID)
	if err != nil {
		t.Fatalf("Step2: %v", err)
	}
	if s2.Weight != 50 {
		t.Fatalf("weight=%d, want 50", s2.Weight)
	}
	s3, err := store.Step(c.ID)
	if err != nil {
		t.Fatalf("Step3: %v", err)
	}
	if s3.Weight != 100 {
		t.Fatalf("weight=%d, want 100", s3.Weight)
	}
	if _, err := store.Step(c.ID); err == nil {
		t.Fatal("Step past end should error until promote")
	}
}

func TestCanary_Abort_FromMidWeight(t *testing.T) {
	store := NewCanaryStore()
	c, err := store.Start("svc-c", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, err := store.Step(c.ID); err != nil {
		t.Fatalf("Step: %v", err)
	}
	aborted, err := store.Abort(c.ID)
	if err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if aborted.Weight != 0 || aborted.Status != "aborted" {
		t.Fatalf("got weight=%d status=%q, want 0/aborted", aborted.Weight, aborted.Status)
	}
	if _, err := store.Step(c.ID); err == nil {
		t.Fatal("Step after abort should error")
	}
}

func TestCanary_Promote_JumpsTo100(t *testing.T) {
	store := NewCanaryStore()
	c, err := store.Start("svc-d", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, err := store.Step(c.ID); err != nil {
		t.Fatalf("Step: %v", err)
	}
	promoted, err := store.Promote(c.ID)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if promoted.Weight != 100 || promoted.Status != "promoted" {
		t.Fatalf("got weight=%d status=%q, want 100/promoted", promoted.Weight, promoted.Status)
	}
	if _, err := store.Step(c.ID); err == nil {
		t.Fatal("Step after promote should error")
	}
}

func TestCanary_TerminalRejectsStep(t *testing.T) {
	store := NewCanaryStore()
	a, _ := store.Start("term-a", nil)
	_, _ = store.Abort(a.ID)
	if _, err := store.Step(a.ID); err == nil {
		t.Fatal("Step after abort: want error")
	}
	b, _ := store.Start("term-b", nil)
	_, _ = store.Promote(b.ID)
	if _, err := store.Step(b.ID); err == nil {
		t.Fatal("Step after promote: want error")
	}
}

func TestCanary_ConcurrentStep(t *testing.T) {
	store := NewCanaryStore()
	c, err := store.Start("race", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.Step(c.ID)
			_, _ = store.Get(c.ID)
		}()
	}
	wg.Wait()
	got, err := store.Get(c.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Weight < 0 || got.Weight > 100 {
		t.Fatalf("unexpected weight %d", got.Weight)
	}
}
