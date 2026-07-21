package main

import (
	"sync"
	"testing"
)

func seedProfiles(t *testing.T, s *ProfileStore) {
	t.Helper()
	if _, err := s.UpsertProfile("k8s-standard", "kubernetes", map[string]string{"gpu": "none"}); err != nil {
		t.Fatalf("Upsert k8s-standard: %v", err)
	}
	if _, err := s.UpsertProfile("k8s-gpu", "kubernetes", map[string]string{"gpu": "nvidia"}); err != nil {
		t.Fatalf("Upsert k8s-gpu: %v", err)
	}
	if _, err := s.UpsertProfile("vm-bake", "vm", map[string]string{"image": "baked"}); err != nil {
		t.Fatalf("Upsert vm-bake: %v", err)
	}
}

func TestProfile_Upsert_ThreeProfiles(t *testing.T) {
	s := NewProfileStore()
	seedProfiles(t, s)
	if got := len(s.Profiles()); got != 3 {
		t.Fatalf("profiles=%d, want 3", got)
	}
}

func TestSchedule_SameWorkload_MapsThreePlacements(t *testing.T) {
	s := NewProfileStore()
	seedProfiles(t, s)
	placements, err := s.Schedule("checkout-api", []string{"k8s-standard", "k8s-gpu", "vm-bake"})
	if err != nil {
		t.Fatalf("Schedule: %v", err)
	}
	if len(placements) < 3 {
		t.Fatalf("placements=%d, want >=3", len(placements))
	}
	for _, p := range placements {
		if p.WorkloadID != "checkout-api" {
			t.Fatalf("placement workload=%q, want checkout-api (same_workload identity)", p.WorkloadID)
		}
	}
	got, err := s.Placements("checkout-api")
	if err != nil {
		t.Fatalf("Placements: %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("stored placements=%d, want >=3", len(got))
	}
}

func TestSchedule_UnknownProfile_Errors(t *testing.T) {
	s := NewProfileStore()
	seedProfiles(t, s)
	if _, err := s.Schedule("checkout-api", []string{"k8s-standard", "does-not-exist"}); err == nil {
		t.Fatal("Schedule with unknown profile: want error")
	}
}

func TestProfile_RejectUnsafeID(t *testing.T) {
	s := NewProfileStore()
	if _, err := s.UpsertProfile("../evil", "vm", nil); err == nil {
		t.Fatal("UpsertProfile with '..': want error")
	}
	if _, err := s.UpsertProfile("a/b", "vm", nil); err == nil {
		t.Fatal("UpsertProfile with '/': want error")
	}
	seedProfiles(t, s)
	if _, err := s.Schedule("../evil", []string{"k8s-standard"}); err == nil {
		t.Fatal("Schedule with unsafe workload id: want error")
	}
}

func TestProfile_ConcurrentSchedule(t *testing.T) {
	s := NewProfileStore()
	seedProfiles(t, s)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Schedule("wl", []string{"k8s-standard", "k8s-gpu", "vm-bake"})
			_, _ = s.Placements("wl")
		}()
	}
	wg.Wait()
	got, err := s.Placements("wl")
	if err != nil {
		t.Fatalf("Placements: %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("placements=%d, want >=3", len(got))
	}
}
