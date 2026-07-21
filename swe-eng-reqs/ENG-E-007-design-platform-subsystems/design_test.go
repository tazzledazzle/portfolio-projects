package main

import (
	"strings"
	"testing"
)

func TestADR_List_CountAtLeastOne(t *testing.T) {
	store, err := NewDesignStore("adr")
	if err != nil {
		t.Fatalf("NewDesignStore: %v", err)
	}
	adrs := store.ListADRs()
	if len(adrs) < 1 {
		t.Fatalf("expected ≥1 ADR, got %d", len(adrs))
	}
	if adrs[0].ID == "" {
		t.Fatal("expected ADR with non-empty ID")
	}
}

func TestADR_Alternatives_AtLeastTwo(t *testing.T) {
	store, err := NewDesignStore("adr")
	if err != nil {
		t.Fatalf("NewDesignStore: %v", err)
	}
	adrs := store.ListADRs()
	if len(adrs) < 1 {
		t.Fatal("need at least one ADR")
	}
	found := false
	for _, a := range adrs {
		if a.AlternativeCount >= 2 {
			found = true
			break
		}
		// Also accept content markers.
		if strings.Count(strings.ToLower(a.Content), "alternative") >= 2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ≥2 alternatives in ADR metadata/content, got %#v", adrs)
	}
}

func TestDesign_Skeleton_ReferencesADR(t *testing.T) {
	store, err := NewDesignStore("adr")
	if err != nil {
		t.Fatalf("NewDesignStore: %v", err)
	}
	skel := store.Skeleton()
	if skel == nil {
		t.Fatal("expected skeleton")
	}
	adrID, _ := skel["adr_id"].(string)
	if adrID == "" {
		t.Fatalf("skeleton missing adr_id: %#v", skel)
	}
	adrs := store.ListADRs()
	match := false
	for _, a := range adrs {
		if a.ID == adrID {
			match = true
			break
		}
	}
	if !match {
		t.Fatalf("skeleton adr_id %q not in ListADRs", adrID)
	}
}

func TestDesign_DecisionRecorded(t *testing.T) {
	store, err := NewDesignStore("adr")
	if err != nil {
		t.Fatalf("NewDesignStore: %v", err)
	}
	if !store.DecisionRecorded() {
		t.Fatal("expected decision_recorded true after load")
	}
}
