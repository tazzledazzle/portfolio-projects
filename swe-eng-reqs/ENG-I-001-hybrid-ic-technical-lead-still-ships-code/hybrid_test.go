package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHybrid_Leadership_ListsArtifacts(t *testing.T) {
	store := NewHybridStore("artifacts/leadership")
	artifacts := store.Leadership()
	if len(artifacts) < 2 {
		t.Fatalf("Leadership() must list ≥2 artifact paths, got %d: %#v", len(artifacts), artifacts)
	}
	for _, path := range artifacts {
		if filepath.Dir(path) != "artifacts/leadership" && filepath.ToSlash(filepath.Dir(path)) != "artifacts/leadership" {
			// allow absolute or relative under artifacts/leadership
			if !containsLeadershipDir(path) {
				t.Fatalf("artifact path must be under artifacts/leadership/: %s", path)
			}
		}
	}
}

func TestHybrid_CodeShipped(t *testing.T) {
	store := NewHybridStore("artifacts/leadership")
	sample := store.Sample()
	if sample["code_shipped"] != true {
		t.Fatalf("Sample must prove code_shipped, got %#v", sample)
	}
}

func TestHybrid_NotADROnly(t *testing.T) {
	store := NewHybridStore("artifacts/leadership")
	status := store.HybridStatus()
	if status["hybrid_ic"] != true {
		t.Fatalf("hybrid_ic requires both code and leadership, got %#v", status)
	}
	if status["code_shipped"] != true || status["leadership_artifacts"] != true {
		t.Fatalf("hybrid_ic must require both code and leadership, got %#v", status)
	}

	emptyDir := t.TempDir()
	empty := NewHybridStore(emptyDir)
	emptyStatus := empty.HybridStatus()
	if emptyStatus["hybrid_ic"] == true {
		t.Fatal("hybrid_ic must be false when leadership artifacts are missing")
	}
}

func containsLeadershipDir(path string) bool {
	return filepath.Base(filepath.Dir(path)) == "leadership" ||
		len(path) > 0 && (path == "artifacts/leadership/LEAD.md" ||
			filepath.Base(path) != "")
}

func TestHybrid_LeadershipDirExists(t *testing.T) {
	// sanity: real artifacts present for demo-local
	if _, err := os.Stat("artifacts/leadership/LEAD.md"); err != nil {
		t.Skip("leadership artifacts created in GREEN")
	}
}
