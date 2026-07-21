package main

import (
	"sync"
	"testing"
)

func TestGate_HighBurn_Deny(t *testing.T) {
	eng := NewGateEngine()
	if err := eng.PutSLO("slo-bad", 0.999, 14.4); err != nil {
		t.Fatalf("PutSLO: %v", err)
	}
	// High error ratio both windows → burn >> threshold
	if err := eng.IngestSeries("slo-bad", SeriesSample{
		ErrorsShort: 50, TotalShort: 100,
		ErrorsLong:  500, TotalLong: 1000,
	}); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	dec, err := eng.Evaluate("slo-bad")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if dec.Decision != "deny" {
		t.Fatalf("decision=%q, want deny", dec.Decision)
	}
	if _, ok := dec.Evidence["burn_short"]; !ok {
		t.Fatal("evidence missing burn_short")
	}
	if _, ok := dec.Evidence["burn_long"]; !ok {
		t.Fatal("evidence missing burn_long")
	}
}

func TestGate_LowBurn_Allow(t *testing.T) {
	eng := NewGateEngine()
	if err := eng.PutSLO("slo-ok", 0.999, 14.4); err != nil {
		t.Fatalf("PutSLO: %v", err)
	}
	if err := eng.IngestSeries("slo-ok", SeriesSample{
		ErrorsShort: 0, TotalShort: 1000,
		ErrorsLong:  1, TotalLong: 100000,
	}); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	dec, err := eng.Evaluate("slo-ok")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if dec.Decision != "allow" {
		t.Fatalf("decision=%q, want allow", dec.Decision)
	}
}

func TestGate_MultiWindowAND(t *testing.T) {
	eng := NewGateEngine()
	if err := eng.PutSLO("slo-and", 0.999, 14.4); err != nil {
		t.Fatalf("PutSLO: %v", err)
	}
	// Only short window burning hard; long healthy → allow (AND)
	if err := eng.IngestSeries("slo-and", SeriesSample{
		ErrorsShort: 50, TotalShort: 100,
		ErrorsLong:  1, TotalLong: 100000,
	}); err != nil {
		t.Fatalf("Ingest short-only: %v", err)
	}
	dec, err := eng.Evaluate("slo-and")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if dec.Decision != "allow" {
		t.Fatalf("one-window burn: decision=%q, want allow", dec.Decision)
	}

	// Both windows over → deny
	if err := eng.IngestSeries("slo-and", SeriesSample{
		ErrorsShort: 50, TotalShort: 100,
		ErrorsLong:  500, TotalLong: 1000,
	}); err != nil {
		t.Fatalf("Ingest both: %v", err)
	}
	dec, err = eng.Evaluate("slo-and")
	if err != nil {
		t.Fatalf("Evaluate both: %v", err)
	}
	if dec.Decision != "deny" {
		t.Fatalf("both-window burn: decision=%q, want deny", dec.Decision)
	}
}

func TestGate_EvidenceServerSide(t *testing.T) {
	eng := NewGateEngine()
	if err := eng.PutSLO("slo-ev", 0.999, 14.4); err != nil {
		t.Fatalf("PutSLO: %v", err)
	}
	if err := eng.IngestSeries("slo-ev", SeriesSample{
		ErrorsShort: 50, TotalShort: 100,
		ErrorsLong:  500, TotalLong: 1000,
	}); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	dec, err := eng.Evaluate("slo-ev")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	// Evidence must be computed from series — client burn cannot override via Evaluate
	if dec.Decision != "deny" {
		t.Fatalf("server evidence should deny high burn, got %q", dec.Decision)
	}
	bs, _ := dec.Evidence["burn_short"].(float64)
	if bs <= 14.4 {
		t.Fatalf("burn_short=%v should exceed threshold from ingested series", bs)
	}
}

func TestGate_UnknownSLO_Errors(t *testing.T) {
	eng := NewGateEngine()
	if _, err := eng.Evaluate("missing"); err == nil {
		t.Fatal("Evaluate missing slo: want error")
	}
}

func TestGate_ConcurrentIngestEvaluate(t *testing.T) {
	eng := NewGateEngine()
	_ = eng.PutSLO("slo-race", 0.999, 14.4)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = eng.IngestSeries("slo-race", SeriesSample{
				ErrorsShort: 1, TotalShort: 1000,
				ErrorsLong:  10, TotalLong: 10000,
			})
			_, _ = eng.Evaluate("slo-race")
		}()
	}
	wg.Wait()
}

func TestBurn_Formula(t *testing.T) {
	// burn = error_ratio / (1 - objective)
	objective := 0.999
	errRatio := 0.5
	want := errRatio / (1 - objective)
	got := burnRate(errRatio, objective)
	if got != want {
		t.Fatalf("burnRate=%v, want %v", got, want)
	}
}
