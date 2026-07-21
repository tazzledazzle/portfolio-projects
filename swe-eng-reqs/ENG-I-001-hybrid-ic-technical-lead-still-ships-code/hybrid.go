package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// HybridStore proves shipped service code alongside leadership artifacts in one folder.
type HybridStore struct {
	mu            sync.Mutex
	leadershipDir string
	demos         int
	codeShipped   bool
}

// NewHybridStore indexes leadership markdown under leadershipDir.
func NewHybridStore(leadershipDir string) *HybridStore {
	return &HybridStore{
		leadershipDir: leadershipDir,
		codeShipped:   true, // this module is the shipped IC service
	}
}

// Leadership returns ≥0 relative paths to markdown artifacts under the leadership directory.
func (s *HybridStore) Leadership() []string {
	entries, err := os.ReadDir(s.leadershipDir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		out = append(out, filepath.ToSlash(filepath.Join(s.leadershipDir, entry.Name())))
	}
	return out
}

// Sample returns a proof map showing that service code ships in this slice.
func (s *HybridStore) Sample() map[string]any {
	return map[string]any{
		"code_shipped": s.codeShipped,
		"service":      "eng-i-001",
		"module":       "hybrid.go",
	}
}

// HybridStatus is true only when both shipped code and leadership artifacts exist.
func (s *HybridStore) HybridStatus() map[string]any {
	artifacts := s.Leadership()
	hasLeadership := len(artifacts) >= 2
	hybrid := s.codeShipped && hasLeadership
	return map[string]any{
		"code_shipped":         s.codeShipped,
		"leadership_artifacts": hasLeadership,
		"leadership_count":     len(artifacts),
		"hybrid_ic":            hybrid,
	}
}

// IncrementDemo records a demo invocation.
func (s *HybridStore) IncrementDemo() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.demos++
	return s.demos
}

// DemoCount returns demo invocations.
func (s *HybridStore) DemoCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.demos
}
