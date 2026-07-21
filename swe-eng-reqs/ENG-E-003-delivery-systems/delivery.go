package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var releaseStages = []string{"dev", "staging", "prod"}

// Release is a multistage delivery record (release stages, not CI jobs).
type Release struct {
	ID      string `json:"id"`
	App     string `json:"app"`
	Stage   string `json:"stage"`
	Version string `json:"version"`
}

// AuditEntry is an append-only promote/rollback record.
type AuditEntry struct {
	At     time.Time `json:"at"`
	Action string    `json:"action"`
	From   string    `json:"from"`
	To     string    `json:"to"`
}

// RollbackHook is invoked once per Rollback call.
type RollbackHook func(id string)

// DeliveryStore holds releases, append-only audit, and rollback hooks.
type DeliveryStore struct {
	mu     sync.Mutex
	items  map[string]*Release
	audit  map[string][]AuditEntry
	hooks  []RollbackHook
	seq    int
}

func NewDeliveryStore() *DeliveryStore {
	return &DeliveryStore{
		items: make(map[string]*Release),
		audit: make(map[string][]AuditEntry),
	}
}

func (s *DeliveryStore) RegisterRollbackHook(h RollbackHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hooks = append(s.hooks, h)
}

func (s *DeliveryStore) Create(app, version, stage string) (*Release, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app == "" {
		return nil, errors.New("app required")
	}
	if stage == "" {
		stage = "dev"
	}
	if !isReleaseStage(stage) {
		return nil, fmt.Errorf("invalid stage %q (allowed: %v)", stage, releaseStages)
	}

	s.seq++
	id := fmt.Sprintf("rel-%d", s.seq)
	rel := &Release{
		ID:      id,
		App:     app,
		Stage:   stage,
		Version: version,
	}
	s.items[id] = rel
	s.audit[id] = nil
	return copyRelease(rel), nil
}

func (s *DeliveryStore) Promote(id string) (*Release, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, ok := s.items[id]
	if !ok {
		return nil, errors.New("release not found")
	}
	idx := releaseStageIndex(rel.Stage)
	if idx < 0 {
		return nil, fmt.Errorf("unknown stage %q", rel.Stage)
	}
	if idx >= len(releaseStages)-1 {
		return nil, errors.New("already at final stage; cannot promote")
	}
	from := rel.Stage
	to := releaseStages[idx+1]
	rel.Stage = to
	s.audit[id] = append(s.audit[id], AuditEntry{
		At:     time.Now().UTC(),
		Action: "promote",
		From:   from,
		To:     to,
	})
	return copyRelease(rel), nil
}

func (s *DeliveryStore) Rollback(id string) error {
	s.mu.Lock()
	rel, ok := s.items[id]
	if !ok {
		s.mu.Unlock()
		return errors.New("release not found")
	}
	from := rel.Stage
	hooks := append([]RollbackHook(nil), s.hooks...)
	s.audit[id] = append(s.audit[id], AuditEntry{
		At:     time.Now().UTC(),
		Action: "rollback",
		From:   from,
		To:     from,
	})
	s.mu.Unlock()
	for _, h := range hooks {
		h(id)
	}
	return nil
}

func (s *DeliveryStore) Audit(id string) ([]AuditEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return nil, errors.New("release not found")
	}
	src := s.audit[id]
	out := make([]AuditEntry, len(src))
	copy(out, src)
	return out, nil
}

func copyRelease(r *Release) *Release {
	cp := *r
	return &cp
}

func isReleaseStage(stage string) bool {
	return releaseStageIndex(stage) >= 0
}

func releaseStageIndex(stage string) int {
	for i, s := range releaseStages {
		if s == stage {
			return i
		}
	}
	return -1
}
