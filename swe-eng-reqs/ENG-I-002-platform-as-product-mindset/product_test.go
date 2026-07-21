package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestProduct_SLA_ReturnsObjectives(t *testing.T) {
	store := NewProductStore("templates")
	sla := store.SLA()
	if sla.Availability == "" || sla.LatencyP99MS <= 0 {
		t.Fatalf("SLA missing objectives: %#v", sla)
	}
}

func TestProduct_Adoption_Counts(t *testing.T) {
	store := NewProductStore("templates")
	store.RecordAdoption("payments")
	store.RecordAdoption("payments")
	store.RecordAdoption("billing")
	a := store.Adoption()
	if a.Teams < 2 || a.Active < 3 {
		t.Fatalf("Adoption counts wrong: %#v", a)
	}
}

func TestProduct_GoldenPath_ArtifactPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "golden-path.md")
	if err := os.WriteFile(path, []byte("# golden path\nself-service onboarding\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewProductStore(dir)
	gp := store.GoldenPath()
	if !strings.Contains(gp.Path, "golden-path.md") {
		t.Fatalf("path should reference golden-path.md: %#v", gp)
	}
	if !strings.Contains(strings.ToLower(gp.Content), "golden") {
		t.Fatalf("content should mention golden: %#v", gp)
	}
}

func TestProduct_ConcurrentAdoption(t *testing.T) {
	store := NewProductStore("templates")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.RecordAdoption(fmt.Sprintf("team-%d", i%5))
		}(i)
	}
	wg.Wait()
	a := store.Adoption()
	if a.Active != 50 || a.Teams != 5 {
		t.Fatalf("concurrent adoption wrong: %#v", a)
	}
}
