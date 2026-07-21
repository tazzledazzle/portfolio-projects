package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var stages = []string{"dev", "staging", "prod"}

// Version is a named artifact version record with an immutable digest and a stage.
type Version struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Stage  string `json:"stage"`
}

// VersionStore holds version records and a separate mutable tag→digest map.
type VersionStore struct {
	mu       sync.Mutex
	versions map[string]*Version
	tags     map[string]string
	seq      int
}

func NewVersionStore() *VersionStore {
	return &VersionStore{
		versions: make(map[string]*Version),
		tags:     make(map[string]string),
	}
}

func (s *VersionStore) Create(name, digest, stage string) (*Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateName(name); err != nil {
		return nil, err
	}
	norm, err := normalizeDigest(digest)
	if err != nil {
		return nil, err
	}
	if stage == "" {
		stage = "dev"
	}
	if !isKnownStage(stage) {
		return nil, fmt.Errorf("invalid stage %q (allowed: %v)", stage, stages)
	}

	s.seq++
	id := fmt.Sprintf("v-%d", s.seq)
	v := &Version{
		ID:     id,
		Name:   name,
		Digest: norm,
		Stage:  stage,
	}
	s.versions[id] = v
	return copyVersion(v), nil
}

func (s *VersionStore) Get(id string) (*Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.versions[id]
	if !ok {
		return nil, errors.New("version not found")
	}
	return copyVersion(v), nil
}

// Promote advances stage by exactly one step (dev→staging→prod). Digest is never mutated.
func (s *VersionStore) Promote(id string) (*Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.versions[id]
	if !ok {
		return nil, errors.New("version not found")
	}
	idx := stageIndex(v.Stage)
	if idx < 0 {
		return nil, fmt.Errorf("unknown stage %q", v.Stage)
	}
	if idx >= len(stages)-1 {
		return nil, errors.New("already at final stage; cannot promote")
	}
	v.Stage = stages[idx+1]
	return copyVersion(v), nil
}

func (s *VersionStore) SetTag(tag, digest string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateName(tag); err != nil {
		return err
	}
	norm, err := normalizeDigest(digest)
	if err != nil {
		return err
	}
	s.tags[tag] = norm
	return nil
}

func (s *VersionStore) GetTag(tag string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.tags[tag]
	return d, ok
}

func copyVersion(v *Version) *Version {
	cp := *v
	return &cp
}

func isKnownStage(stage string) bool {
	return stageIndex(stage) >= 0
}

func stageIndex(stage string) int {
	for i, s := range stages {
		if s == stage {
			return i
		}
	}
	return -1
}

func validateName(name string) error {
	if name == "" {
		return errors.New("name required")
	}
	if strings.Contains(name, "..") {
		return errors.New("name must not contain '..'")
	}
	return nil
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
		return "", errors.New("digest must be 64 hex characters (optionally prefixed with sha256:)")
	}
	for _, c := range d {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", errors.New("digest must be lowercase hex")
		}
	}
	return "sha256:" + d, nil
}
