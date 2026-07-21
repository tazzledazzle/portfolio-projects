package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ENG-N-008 owns the PROFILE abstraction and scheduling the SAME workload
// across ≥3 heterogeneous deployment profiles (D-03, D-11): k8s-standard,
// k8s-gpu, vm-bake. It does NOT own canary weights, burn-rate gates, or
// reconcile finalizers.

// Profile describes a heterogeneous deployment target.
type Profile struct {
	Name        string            `json:"name"`
	Runtime     string            `json:"runtime"`
	Constraints map[string]string `json:"constraints"`
}

// Placement binds one workload to one profile. The workload identity is
// preserved across every placement (same_workload).
type Placement struct {
	WorkloadID string `json:"workload_id"`
	Profile    string `json:"profile"`
	Runtime    string `json:"runtime"`
}

// ProfileStore is a registry of profiles plus per-workload placements.
type ProfileStore struct {
	mu         sync.Mutex
	profiles   map[string]*Profile
	placements map[string][]Placement
}

func NewProfileStore() *ProfileStore {
	return &ProfileStore{
		profiles:   make(map[string]*Profile),
		placements: make(map[string][]Placement),
	}
}

// UpsertProfile creates or replaces a profile. Names must be safe identifiers.
func (s *ProfileStore) UpsertProfile(name, runtime string, constraints map[string]string) (*Profile, error) {
	if !safeProfileID(name) {
		return nil, fmt.Errorf("invalid profile name %q", name)
	}
	if runtime == "" {
		runtime = "kubernetes"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make(map[string]string, len(constraints))
	for k, v := range constraints {
		cp[k] = v
	}
	p := &Profile{Name: name, Runtime: runtime, Constraints: cp}
	s.profiles[name] = p
	return copyProfile(p), nil
}

// Profiles returns a snapshot list of registered profiles, name-sorted.
func (s *ProfileStore) Profiles() []Profile {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Profile, 0, len(s.profiles))
	for _, p := range s.profiles {
		out = append(out, *copyProfile(p))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Schedule maps one workload onto the named profiles, preserving workload
// identity across every placement. All named profiles must be registered.
// Re-scheduling a workload replaces its prior placements (idempotent).
func (s *ProfileStore) Schedule(workloadID string, profileNames []string) ([]Placement, error) {
	if !safeProfileID(workloadID) {
		return nil, fmt.Errorf("invalid workload id %q", workloadID)
	}
	if len(profileNames) == 0 {
		return nil, errors.New("at least one profile required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	placements := make([]Placement, 0, len(profileNames))
	for _, name := range profileNames {
		prof, ok := s.profiles[name]
		if !ok {
			return nil, fmt.Errorf("unknown profile %q", name)
		}
		placements = append(placements, Placement{
			WorkloadID: workloadID,
			Profile:    prof.Name,
			Runtime:    prof.Runtime,
		})
	}
	s.placements[workloadID] = placements
	return copyPlacements(placements), nil
}

// Placements returns the placements recorded for a workload.
func (s *ProfileStore) Placements(workloadID string) ([]Placement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.placements[workloadID]
	if !ok {
		return nil, errors.New("workload not scheduled")
	}
	return copyPlacements(p), nil
}

func copyProfile(p *Profile) *Profile {
	cp := *p
	cp.Constraints = make(map[string]string, len(p.Constraints))
	for k, v := range p.Constraints {
		cp.Constraints[k] = v
	}
	return &cp
}

func copyPlacements(src []Placement) []Placement {
	out := make([]Placement, len(src))
	copy(out, src)
	return out
}

// safeProfileID enforces the ^[a-z0-9][a-z0-9-]*$ alphabet and rejects path
// traversal (analog of ENG-E-025 safeDevEnvID).
func safeProfileID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return false
	}
	for i, r := range id {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if i == 0 && !(isLower || isDigit) {
			return false
		}
		if !isLower && !isDigit && r != '-' {
			return false
		}
	}
	return true
}
