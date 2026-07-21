package main

import (
	"fmt"
	"testing"
)

func TestMultiDC_Register_AtLeast40(t *testing.T) {
	mdc := NewMultiDC()
	for i := 0; i < 42; i++ {
		domain := fmt.Sprintf("fd-%d", i%6)
		if err := mdc.RegisterDC(fmt.Sprintf("dc-%02d", i), domain, true); err != nil {
			t.Fatalf("RegisterDC dc-%02d: %v", i, err)
		}
	}
	if got := mdc.Count(); got < 40 {
		t.Fatalf("expected ≥40 DCs, got %d", got)
	}
}

func TestMultiDC_FailureDomains_Grouped(t *testing.T) {
	mdc := NewMultiDC()
	_ = mdc.RegisterDC("dc-a1", "fd-east", true)
	_ = mdc.RegisterDC("dc-a2", "fd-east", true)
	_ = mdc.RegisterDC("dc-b1", "fd-west", true)

	domains := mdc.Domains()
	if len(domains) < 2 {
		t.Fatalf("expected ≥2 failure domains, got %#v", domains)
	}
	east, ok := domains["fd-east"]
	if !ok || len(east) != 2 {
		t.Fatalf("fd-east should group 2 DCs, got %#v", domains)
	}
	west, ok := domains["fd-west"]
	if !ok || len(west) != 1 {
		t.Fatalf("fd-west should group 1 DC, got %#v", domains)
	}
}

func TestMultiDC_Fanout_PartialOK(t *testing.T) {
	mdc := NewMultiDC()
	_ = mdc.RegisterDC("dc-ok-1", "fd-1", true)
	_ = mdc.RegisterDC("dc-ok-2", "fd-1", true)
	_ = mdc.RegisterDC("dc-down", "fd-2", false)

	result, err := mdc.Fanout(map[string]any{"revision": "cfg-1", "feature": "x"})
	if err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if !result.FanoutOK {
		t.Fatal("expected fanout_ok when healthy DCs receive config")
	}
	if result.Pushed < 2 {
		t.Fatalf("expected ≥2 healthy pushes, got %d", result.Pushed)
	}
	if result.Failed < 1 {
		t.Fatalf("expected ≥1 failed push for unhealthy DC, got %d", result.Failed)
	}
}

func TestMultiDC_Info_Simulator(t *testing.T) {
	mdc := NewMultiDC()
	info := mdc.Info()
	if info["simulator"] != true {
		t.Fatalf("expected simulator=true, got %#v", info)
	}
	if info["multi_dc_simulator"] != true {
		t.Fatalf("expected multi_dc_simulator=true, got %#v", info)
	}
	note, _ := info["note"].(string)
	if note == "" {
		t.Fatal("expected honesty note that this is not physical DCs")
	}
}
