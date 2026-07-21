package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSLO_PutGet(t *testing.T) {
	store := NewReliabilityStore("runbooks")
	want, err := store.PutSLO("api-availability", 0.999, "successful_requests / total_requests")
	if err != nil {
		t.Fatalf("PutSLO: %v", err)
	}
	got, ok := store.GetSLO("api-availability")
	if !ok {
		t.Fatal("SLO not found")
	}
	if got != want || got.Objective != 0.999 || got.SLI == "" {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGolden_KeysPresent(t *testing.T) {
	signals := NewReliabilityStore("runbooks").GoldenSignals()
	for _, key := range []string{"latency", "traffic", "errors", "saturation"} {
		if signals[key] == "" {
			t.Errorf("golden signal %q missing: %#v", key, signals)
		}
	}
}

func TestRunbook_IndexNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "error-budget.md"), []byte("# Error budget"), 0o600); err != nil {
		t.Fatalf("write runbook: %v", err)
	}
	runbooks, err := NewReliabilityStore(dir).ListRunbooks()
	if err != nil {
		t.Fatalf("ListRunbooks: %v", err)
	}
	if len(runbooks) < 1 || runbooks[0].Name != "error-budget" {
		t.Fatalf("unexpected runbook index: %#v", runbooks)
	}
}

func TestReliability_ConcurrentPut(t *testing.T) {
	store := NewReliabilityStore("runbooks")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("slo-%d", i)
			if _, err := store.PutSLO(id, 0.99, "good / total"); err != nil {
				t.Errorf("PutSLO(%s): %v", id, err)
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 20; i++ {
		if _, ok := store.GetSLO(fmt.Sprintf("slo-%d", i)); !ok {
			t.Errorf("slo-%d missing", i)
		}
	}
}
