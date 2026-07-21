package main

import (
	"sync"
	"testing"
)

func TestDeliver_Create_PromoteAdvancesStage(t *testing.T) {
	store := NewDeliveryStore()
	rel, err := store.Create("demo-app", "v1.0.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rel.Stage != "dev" {
		t.Fatalf("stage=%q, want dev", rel.Stage)
	}
	p1, err := store.Promote(rel.ID)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if p1.Stage != "staging" {
		t.Fatalf("after first promote stage=%q, want staging", p1.Stage)
	}
	p2, err := store.Promote(rel.ID)
	if err != nil {
		t.Fatalf("second Promote: %v", err)
	}
	if p2.Stage != "prod" {
		t.Fatalf("after second promote stage=%q, want prod", p2.Stage)
	}
}

func TestDeliver_Promote_AppendsAudit(t *testing.T) {
	store := NewDeliveryStore()
	rel, err := store.Create("app", "1.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	before, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if _, err := store.Promote(rel.ID); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	after, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit after: %v", err)
	}
	if len(after) <= len(before) {
		t.Fatalf("audit length did not grow: before=%d after=%d", len(before), len(after))
	}
	last := after[len(after)-1]
	if last.Action != "promote" {
		t.Fatalf("audit action=%q, want promote", last.Action)
	}
	if last.From != "dev" || last.To != "staging" {
		t.Fatalf("audit From/To=%q→%q, want dev→staging", last.From, last.To)
	}
}

func TestDeliver_Rollback_InvokesHookOnce(t *testing.T) {
	store := NewDeliveryStore()
	rel, err := store.Create("app", "1.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Promote(rel.ID); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	calls := 0
	store.RegisterRollbackHook(func(id string) {
		calls++
	})
	if err := store.Rollback(rel.ID); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if calls != 1 {
		t.Fatalf("hook calls=%d, want 1", calls)
	}
	entries, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Action == "rollback" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected audit entry with action rollback")
	}
}

func TestDeliver_UnknownID_Errors(t *testing.T) {
	store := NewDeliveryStore()
	if _, err := store.Promote("missing"); err == nil {
		t.Fatal("Promote missing id: want error")
	}
	if err := store.Rollback("missing"); err == nil {
		t.Fatal("Rollback missing id: want error")
	}
	if _, err := store.Audit("missing"); err == nil {
		t.Fatal("Audit missing id: want error")
	}
}

func TestDeliver_AuditAppendOnly(t *testing.T) {
	store := NewDeliveryStore()
	rel, err := store.Create("app", "1.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Promote(rel.ID); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	first, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(first) == 0 {
		t.Fatal("expected at least one audit entry after promote")
	}
	snapshot := first[0]
	if _, err := store.Promote(rel.ID); err != nil {
		t.Fatalf("second Promote: %v", err)
	}
	second, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit after second: %v", err)
	}
	if len(second) < 2 {
		t.Fatalf("audit len=%d, want ≥2", len(second))
	}
	if second[0].Action != snapshot.Action || second[0].From != snapshot.From || second[0].To != snapshot.To {
		t.Fatalf("prior audit mutated: got %+v want %+v", second[0], snapshot)
	}
}

func TestDeliver_ConcurrentPromote(t *testing.T) {
	store := NewDeliveryStore()
	rel, err := store.Create("race-app", "1.0", "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.Promote(rel.ID)
			_, _ = store.Audit(rel.ID)
		}()
	}
	wg.Wait()
	entries, err := store.Audit(rel.ID)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected audit entries after concurrent promote")
	}
}
