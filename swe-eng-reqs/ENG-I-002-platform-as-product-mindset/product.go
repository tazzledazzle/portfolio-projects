package main

import (
	"os"
	"path/filepath"
	"sync"
)

type SLAObjectives struct {
	Availability string `json:"availability"`
	LatencyP99MS int    `json:"latency_p99_ms"`
	SupportHours string `json:"support_hours"`
}

type AdoptionMetrics struct {
	Teams  int `json:"teams"`
	Active int `json:"active"`
}

type GoldenPathArtifact struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ProductStore struct {
	mu           sync.Mutex
	templateDir  string
	teamActive   map[string]int
	activeEvents int
}

func NewProductStore(templateDir string) *ProductStore {
	return &ProductStore{
		templateDir: templateDir,
		teamActive:  make(map[string]int),
	}
}

func (s *ProductStore) SLA() SLAObjectives {
	return SLAObjectives{
		Availability: "99.9%",
		LatencyP99MS: 200,
		SupportHours: "business-hours",
	}
}

func (s *ProductStore) RecordAdoption(team string) {
	if team == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teamActive[team]++
	s.activeEvents++
}

func (s *ProductStore) Adoption() AdoptionMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AdoptionMetrics{Teams: len(s.teamActive), Active: s.activeEvents}
}

func (s *ProductStore) GoldenPath() GoldenPathArtifact {
	path := filepath.Join(s.templateDir, "golden-path.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return GoldenPathArtifact{Path: path, Content: ""}
	}
	return GoldenPathArtifact{Path: path, Content: string(content)}
}
