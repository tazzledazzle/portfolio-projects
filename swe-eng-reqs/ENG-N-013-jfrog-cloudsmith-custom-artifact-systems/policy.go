package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const requiredWriteScope = "artifacts:write"

// Artifact is an in-memory registry entry (simulator only).
type Artifact struct {
	Name      string    `json:"name"`
	Digest    string    `json:"digest"`
	CreatedAt time.Time `json:"created_at"`
}

// Finding is a fixture vulnerability-scan result (no network scanner).
type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Package  string `json:"package"`
	Title    string `json:"title"`
}

// PolicyEngine simulates custom-registry scopes, retention, and scan hooks.
type PolicyEngine struct {
	mu        sync.Mutex
	artifacts []Artifact
	seq       int
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		artifacts: make([]Artifact, 0),
	}
}

func (e *PolicyEngine) Authorize(scopes []string) (bool, error) {
	for _, s := range scopes {
		if s == requiredWriteScope {
			return true, nil
		}
	}
	return false, errors.New("missing required scope: artifacts:write")
}

func (e *PolicyEngine) PutArtifact(name, digest string, scopes []string) error {
	ok, err := e.Authorize(scopes)
	if !ok {
		return err
	}
	if name == "" || strings.Contains(name, "..") {
		return errors.New("invalid artifact name")
	}
	norm, err := normalizeDigest(digest)
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seq++
	e.artifacts = append(e.artifacts, Artifact{
		Name:      name,
		Digest:    norm,
		CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (e *PolicyEngine) ListArtifacts() []Artifact {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Artifact, len(e.artifacts))
	copy(out, e.artifacts)
	return out
}

// RunRetention keeps the newest keepCount artifacts; deletes older ones.
// Digests of remaining artifacts are never rewritten.
func (e *PolicyEngine) RunRetention(keepCount int) (int, error) {
	if keepCount < 0 {
		return 0, errors.New("keepCount must be >= 0")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	n := len(e.artifacts)
	if n <= keepCount {
		return 0, nil
	}
	deleted := n - keepCount
	// Keep newest = last keepCount entries (insertion order)
	e.artifacts = append([]Artifact(nil), e.artifacts[n-keepCount:]...)
	return deleted, nil
}

// Scan returns fixture findings only — never calls a network scanner.
func (e *PolicyEngine) Scan() []Finding {
	return []Finding{
		{
			ID:       "SIM-CVE-0001",
			Severity: "medium",
			Package:  "demo-lib",
			Title:    "Fixture finding (simulator stub — not a real scanner)",
		},
		{
			ID:       "SIM-CVE-0002",
			Severity: "low",
			Package:  "demo-util",
			Title:    "Fixture finding (simulator stub — not a real scanner)",
		},
	}
}

func (e *PolicyEngine) Info() map[string]any {
	return map[string]any{
		"requirement_id": "ENG-N-013",
		"service":        "eng-n-013",
		"title":          "Custom registry simulator",
		"simulator":      true,
		"vendor_model":   "custom-registry",
		"note":           "Does NOT connect to JFrog Artifactory or Cloudsmith",
	}
}

func normalizeDigest(digest string) (string, error) {
	if digest == "" || strings.Contains(digest, "..") {
		return "", errors.New("invalid digest")
	}
	d := digest
	if strings.HasPrefix(d, "sha256:") {
		d = strings.TrimPrefix(d, "sha256:")
	}
	if len(d) != 64 {
		return "", fmt.Errorf("digest must be 64 hex characters")
	}
	for _, c := range d {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", errors.New("digest must be lowercase hex")
		}
	}
	return "sha256:" + d, nil
}
