package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SLO defines an objective and the service-level indicator used to measure it.
type SLO struct {
	ID        string  `json:"id"`
	Objective float64 `json:"objective"`
	SLI       string  `json:"sli"`
}

// Runbook identifies a local, reviewed operational procedure.
type Runbook struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ReliabilityStore owns SLO definitions and indexes local operability artifacts.
type ReliabilityStore struct {
	mu         sync.RWMutex
	slos       map[string]SLO
	runbookDir string
}

// NewReliabilityStore returns an empty store backed by the given runbook directory.
func NewReliabilityStore(runbookDir string) *ReliabilityStore {
	return &ReliabilityStore{
		slos:       make(map[string]SLO),
		runbookDir: runbookDir,
	}
}

// PutSLO validates and stores an SLO definition.
func (s *ReliabilityStore) PutSLO(id string, objective float64, sli string) (SLO, error) {
	if !safeSLOID(id) || objective <= 0 || objective >= 1 || strings.TrimSpace(sli) == "" {
		return SLO{}, errors.New("invalid SLO")
	}
	slo := SLO{ID: id, Objective: objective, SLI: strings.TrimSpace(sli)}
	s.mu.Lock()
	s.slos[id] = slo
	s.mu.Unlock()
	return slo, nil
}

// GetSLO returns an SLO by identifier.
func (s *ReliabilityStore) GetSLO(id string) (SLO, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slo, ok := s.slos[id]
	return slo, ok
}

// GoldenSignals returns the canonical four operability signal names.
func (s *ReliabilityStore) GoldenSignals() map[string]string {
	return map[string]string{
		"latency":    "request_duration_seconds",
		"traffic":    "requests_total",
		"errors":     "request_errors_total",
		"saturation": "resource_saturation_ratio",
	}
}

// ListRunbooks indexes Markdown runbooks without exposing their contents.
func (s *ReliabilityStore) ListRunbooks() ([]Runbook, error) {
	entries, err := os.ReadDir(s.runbookDir)
	if err != nil {
		return nil, err
	}
	runbooks := make([]Runbook, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		runbooks = append(runbooks, Runbook{
			Name: strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Path: filepath.ToSlash(filepath.Join(s.runbookDir, entry.Name())),
		})
	}
	return runbooks, nil
}

func safeSLOID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}
