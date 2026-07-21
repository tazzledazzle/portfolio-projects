package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ADR is a loaded architecture decision record.
type ADR struct {
	ID                string `json:"id"`
	Path              string `json:"path"`
	Title             string `json:"title"`
	Content           string `json:"content"`
	AlternativeCount  int    `json:"alternative_count"`
	DecisionRecorded  bool   `json:"decision_recorded"`
}

// DesignStore indexes ADRs from disk and exposes a thin subsystem skeleton.
type DesignStore struct {
	mu   sync.RWMutex
	dir  string
	adrs []ADR
}

func NewDesignStore(dir string) (*DesignStore, error) {
	s := &DesignStore{dir: dir}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *DesignStore) load() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	var adrs []ADR
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(b)
		id := strings.TrimSuffix(e.Name(), ".md")
		title := firstHeading(content)
		altCount := countAlternatives(content)
		decision := strings.Contains(strings.ToLower(content), "## decision")
		adrs = append(adrs, ADR{
			ID:               id,
			Path:             path,
			Title:            title,
			Content:          content,
			AlternativeCount: altCount,
			DecisionRecorded: decision,
		})
	}
	s.mu.Lock()
	s.adrs = adrs
	s.mu.Unlock()
	return nil
}

func (s *DesignStore) ListADRs() []ADR {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ADR, len(s.adrs))
	copy(out, s.adrs)
	return out
}

func (s *DesignStore) DecisionRecorded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.adrs {
		if a.DecisionRecorded && a.AlternativeCount >= 2 {
			return true
		}
	}
	return false
}

// Skeleton returns a thin runtime reference to the primary ADR.
func (s *DesignStore) Skeleton() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	adrID := ""
	alts := 0
	if len(s.adrs) > 0 {
		adrID = s.adrs[0].ID
		alts = s.adrs[0].AlternativeCount
	}
	return map[string]any{
		"service":            "eng-e-007",
		"requirement_id":     "ENG-E-007",
		"adr_id":             adrID,
		"thin_skeleton":      true,
		"alternative_count":  alts,
		"decision_recorded":  s.decisionRecordedLocked(),
		"does_not_own":       []string{"full_prod_metrics", "hpa_packaging", "mentoring_kits"},
	}
}

func (s *DesignStore) decisionRecordedLocked() bool {
	for _, a := range s.adrs {
		if a.DecisionRecorded && a.AlternativeCount >= 2 {
			return true
		}
	}
	return false
}

func (s *DesignStore) Info() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]any{
		"requirement_id":     "ENG-E-007",
		"service":            "eng-e-007",
		"title":              "Design platform subsystems",
		"adr_count":          len(s.adrs),
		"decision_recorded":  s.decisionRecordedLocked(),
		"owns":               []string{"adr_alternatives", "thin_skeleton"},
		"does_not_own":       []string{"full_prod_metrics", "hpa_packaging", "mentoring_kits"},
	}
}

func firstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func countAlternatives(content string) int {
	lower := strings.ToLower(content)
	count := 0
	// Count "### Alternative" headings (A/B/C style).
	for _, line := range strings.Split(lower, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "### alternative") {
			count++
		}
	}
	if count >= 2 {
		return count
	}
	// Fallback: count standalone "Alternative X" markers.
	count = strings.Count(lower, "### alternative")
	if count >= 2 {
		return count
	}
	return strings.Count(lower, "alternative ")
}
