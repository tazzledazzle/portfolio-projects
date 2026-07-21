package main

import (
	"errors"
	"fmt"
	"sync"
)

var defaultWeights = []int{0, 10, 50, 100}

// Canary is a single-target progressive rollout with weight steps.
type Canary struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Weights []int  `json:"weights"`
	Index   int    `json:"index"`
	Weight  int    `json:"weight"`
	Status  string `json:"status"` // running | aborted | promoted
}

// CanaryStore holds canary rollouts under a mutex.
type CanaryStore struct {
	mu    sync.Mutex
	items map[string]*Canary
	seq   int
}

func NewCanaryStore() *CanaryStore {
	return &CanaryStore{items: make(map[string]*Canary)}
}

func (s *CanaryStore) Start(service string, steps []int) (*Canary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if service == "" {
		return nil, errors.New("service required")
	}
	weights := steps
	if len(weights) == 0 {
		weights = append([]int(nil), defaultWeights...)
	}
	s.seq++
	id := fmt.Sprintf("canary-%d", s.seq)
	c := &Canary{
		ID:      id,
		Service: service,
		Weights: weights,
		Index:   0,
		Weight:  weights[0],
		Status:  "running",
	}
	s.items[id] = c
	return copyCanary(c), nil
}

func (s *CanaryStore) Get(id string) (*Canary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.items[id]
	if !ok {
		return nil, errors.New("canary not found")
	}
	return copyCanary(c), nil
}

func (s *CanaryStore) Step(id string) (*Canary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.items[id]
	if !ok {
		return nil, errors.New("canary not found")
	}
	if c.Status != "running" {
		return nil, fmt.Errorf("canary is %s; cannot step", c.Status)
	}
	if c.Index >= len(c.Weights)-1 {
		return nil, errors.New("already at final weight; promote or abort")
	}
	c.Index++
	c.Weight = c.Weights[c.Index]
	return copyCanary(c), nil
}

func (s *CanaryStore) Abort(id string) (*Canary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.items[id]
	if !ok {
		return nil, errors.New("canary not found")
	}
	if c.Status != "running" {
		return nil, fmt.Errorf("canary is %s; cannot abort", c.Status)
	}
	c.Weight = 0
	c.Index = 0
	c.Status = "aborted"
	return copyCanary(c), nil
}

func (s *CanaryStore) Promote(id string) (*Canary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.items[id]
	if !ok {
		return nil, errors.New("canary not found")
	}
	if c.Status != "running" {
		return nil, fmt.Errorf("canary is %s; cannot promote", c.Status)
	}
	c.Weight = 100
	c.Index = len(c.Weights) - 1
	c.Status = "promoted"
	return copyCanary(c), nil
}

func copyCanary(c *Canary) *Canary {
	cp := *c
	cp.Weights = append([]int(nil), c.Weights...)
	return &cp
}
