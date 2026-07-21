package main

import (
	"sync"
	"testing"
)

func TestTenant_Quota_Enforced(t *testing.T) {
	s := NewTenantScheduler()
	if err := s.SetQuota("tenant-a", 2); err != nil {
		t.Fatalf("SetQuota: %v", err)
	}
	if _, err := s.Schedule("tenant-a", 1); err != nil {
		t.Fatalf("first schedule: %v", err)
	}
	if _, err := s.Schedule("tenant-a", 1); err != nil {
		t.Fatalf("second schedule: %v", err)
	}
	_, err := s.Schedule("tenant-a", 1)
	if err == nil {
		t.Fatal("expected quota denial on third schedule")
	}
	proof := s.Proof()
	if proof["quota_enforced"] != true {
		t.Fatalf("expected quota_enforced, got %#v", proof)
	}
}

func TestTenant_NoisyNeighbor_Limited(t *testing.T) {
	s := NewTenantScheduler()
	_ = s.SetQuota("noisy", 100)
	_ = s.SetQuota("quiet", 10)
	s.SetRateLimit("noisy", 2) // max concurrent/scheduled units per window

	// Burn noisy tenant rate limit
	if _, err := s.Schedule("noisy", 1); err != nil {
		t.Fatalf("noisy 1: %v", err)
	}
	if _, err := s.Schedule("noisy", 1); err != nil {
		t.Fatalf("noisy 2: %v", err)
	}
	if _, err := s.Schedule("noisy", 1); err == nil {
		t.Fatal("expected noisy neighbor limit denial")
	}

	// Quiet tenant must still schedule
	if _, err := s.Schedule("quiet", 1); err != nil {
		t.Fatalf("quiet tenant blocked by noisy neighbor: %v", err)
	}
	proof := s.Proof()
	if proof["noisy_neighbor_limited"] != true {
		t.Fatalf("expected noisy_neighbor_limited, got %#v", proof)
	}
}

func TestTenant_RequiresTenantID(t *testing.T) {
	s := NewTenantScheduler()
	if err := s.SetQuota("", 5); err == nil {
		t.Fatal("expected reject empty tenant on SetQuota")
	}
	if _, err := s.Schedule("", 1); err == nil {
		t.Fatal("expected reject missing tenant ID on Schedule")
	}
	if _, err := s.Schedule("  ", 1); err == nil {
		t.Fatal("expected reject blank tenant ID")
	}
}

func TestTenant_ConcurrentSchedule(t *testing.T) {
	s := NewTenantScheduler()
	_ = s.SetQuota("t1", 50)
	_ = s.SetQuota("t2", 50)

	var wg sync.WaitGroup
	errs := make(chan error, 100)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "t1"
			if i%2 == 0 {
				id = "t2"
			}
			if _, err := s.Schedule(id, 1); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent schedule: %v", err)
	}
}
