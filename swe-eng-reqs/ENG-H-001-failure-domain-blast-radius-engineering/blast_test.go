package main

import (
	"testing"
)

func TestBlast_Chaos_Contained(t *testing.T) {
	eng := NewBlastEngine()
	_ = eng.RegisterDomain("fd-east", []string{"tenant-a", "tenant-b"})
	_ = eng.RegisterDomain("fd-west", []string{"tenant-c"})

	result, err := eng.RunChaos("fd-east", "partition")
	if err != nil {
		t.Fatalf("RunChaos: %v", err)
	}
	if !result.Contained {
		t.Fatal("expected contained=true")
	}
	radius := eng.BlastRadius()
	if len(radius.UnaffectedTenants) == 0 {
		t.Fatalf("expected unaffected tenants, got %#v", radius)
	}
	if len(radius.UnaffectedDomains) == 0 {
		t.Fatalf("expected unaffected domains, got %#v", radius)
	}
}

func TestBlast_AffectedDomains_Limited(t *testing.T) {
	eng := NewBlastEngine()
	_ = eng.RegisterDomain("fd-1", []string{"t1"})
	_ = eng.RegisterDomain("fd-2", []string{"t2"})
	_ = eng.RegisterDomain("fd-3", []string{"t3"})

	if _, err := eng.RunChaos("fd-1", "crash"); err != nil {
		t.Fatalf("RunChaos: %v", err)
	}
	radius := eng.BlastRadius()
	if len(radius.AffectedDomains) == 0 {
		t.Fatal("expected at least one affected domain")
	}
	total := len(radius.AffectedDomains) + len(radius.UnaffectedDomains)
	if total < 3 {
		t.Fatalf("expected all domains accounted for, got affected=%v unaffected=%v", radius.AffectedDomains, radius.UnaffectedDomains)
	}
	for _, d := range radius.AffectedDomains {
		if d == "fd-2" || d == "fd-3" {
			t.Fatalf("isolation failed: unaffected domain listed as affected: %s", d)
		}
	}
	if len(radius.AffectedDomains) >= 3 {
		t.Fatalf("affected_domains should not include all domains when isolation works: %#v", radius.AffectedDomains)
	}
}

func TestBlast_UnaffectedTenants(t *testing.T) {
	eng := NewBlastEngine()
	_ = eng.RegisterDomain("fd-hot", []string{"victim"})
	_ = eng.RegisterDomain("fd-cold", []string{"survivor"})

	if _, err := eng.RunChaos("fd-hot", "latency"); err != nil {
		t.Fatalf("RunChaos: %v", err)
	}
	radius := eng.BlastRadius()
	found := false
	for _, tenant := range radius.UnaffectedTenants {
		if tenant == "survivor" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected survivor unaffected, got %#v", radius.UnaffectedTenants)
	}
}
