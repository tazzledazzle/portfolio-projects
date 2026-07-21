package main

import (
	"strings"
	"testing"
)

func TestExpandMatrix_SingleDimension(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{
		Dimensions: map[string][]string{"os": {"ubuntu", "macos"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("want 2 jobs, got %d", len(jobs))
	}
}

func TestExpandMatrix_TwoDimensions(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{
		Dimensions: map[string][]string{
			"os":   {"ubuntu"},
			"node": {"14", "16"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("want 2 jobs, got %d", len(jobs))
	}
}

func TestExpandMatrix_ThreeDimensions(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{
		Dimensions: map[string][]string{
			"os":   {"ubuntu", "macos"},
			"node": {"14", "16"},
			"arch": {"x64", "arm64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 8 {
		t.Fatalf("want 8 jobs, got %d", len(jobs))
	}
}

func TestExpandMatrix_WithInclude(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{
		Dimensions: map[string][]string{
			"os":   {"ubuntu"},
			"node": {"14"},
		},
		Include: []map[string]string{{"os": "windows", "node": "18"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("want 2 jobs (1 base + 1 include), got %d", len(jobs))
	}
}

func TestExpandMatrix_WithExclude(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{
		Dimensions: map[string][]string{
			"os":   {"ubuntu", "macos"},
			"node": {"14", "16"},
		},
		Exclude: []map[string]string{{"os": "ubuntu", "node": "14"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 3 {
		t.Fatalf("want 3 jobs after exclude, got %d", len(jobs))
	}
}

func TestExpandMatrix_EmptyMatrix(t *testing.T) {
	jobs, err := ExpandMatrix(MatrixConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("want 1 default job, got %d", len(jobs))
	}
}

func TestExpandMatrix_MaxCombinations(t *testing.T) {
	dims := map[string][]string{
		"a": make([]string, 17),
		"b": make([]string, 17),
	}
	for i := 0; i < 17; i++ {
		dims["a"][i] = string(rune('a' + i))
		dims["b"][i] = string(rune('A' + i))
	}
	_, err := ExpandMatrix(MatrixConfig{Dimensions: dims})
	if err == nil || !strings.Contains(err.Error(), "matrix too large") {
		t.Fatalf("want matrix too large error, got %v", err)
	}
}
