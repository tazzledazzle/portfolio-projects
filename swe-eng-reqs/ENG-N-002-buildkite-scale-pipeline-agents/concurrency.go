package main

import "sync"

// Group tracks a Buildkite concurrency group.
type Group struct {
	Limit   int
	Current int
}

// ConcurrencyManager enforces max parallel jobs per group.
type ConcurrencyManager struct {
	mu     sync.Mutex
	groups map[string]*Group
}

func NewConcurrencyManager() *ConcurrencyManager {
	return &ConcurrencyManager{groups: map[string]*Group{}}
}

func (m *ConcurrencyManager) Acquire(groupName string, limit int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[groupName]
	if !ok {
		g = &Group{Limit: limit}
		m.groups[groupName] = g
	}
	if limit > 0 {
		g.Limit = limit
	}
	if g.Current >= g.Limit {
		return false
	}
	g.Current++
	return true
}

func (m *ConcurrencyManager) Release(groupName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[groupName]
	if !ok {
		return
	}
	if g.Current > 0 {
		g.Current--
	}
}

func (m *ConcurrencyManager) GroupStatus(groupName string) (limit, current int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[groupName]
	if !ok {
		return 0, 0
	}
	return g.Limit, g.Current
}

func (m *ConcurrencyManager) AllStatuses() map[string]map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]map[string]int, len(m.groups))
	for name, g := range m.groups {
		out[name] = map[string]int{"limit": g.Limit, "current": g.Current}
	}
	return out
}

func (m *ConcurrencyManager) GroupCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.groups)
}
