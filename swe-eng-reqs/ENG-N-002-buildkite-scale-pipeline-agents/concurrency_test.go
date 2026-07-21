package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrencyManager_Acquire_Available(t *testing.T) {
	m := NewConcurrencyManager()
	if !m.Acquire("deploy", 2) {
		t.Fatal("expected acquire true")
	}
}

func TestConcurrencyManager_Acquire_AtLimit(t *testing.T) {
	m := NewConcurrencyManager()
	_ = m.Acquire("deploy", 2)
	_ = m.Acquire("deploy", 2)
	if m.Acquire("deploy", 2) {
		t.Fatal("expected acquire false at limit")
	}
}

func TestConcurrencyManager_Release(t *testing.T) {
	m := NewConcurrencyManager()
	_ = m.Acquire("deploy", 1)
	m.Release("deploy")
	if !m.Acquire("deploy", 1) {
		t.Fatal("expected acquire after release")
	}
}

func TestConcurrencyManager_GroupStatus(t *testing.T) {
	m := NewConcurrencyManager()
	_ = m.Acquire("deploy", 3)
	limit, current := m.GroupStatus("deploy")
	if limit != 3 || current != 1 {
		t.Fatalf("want limit=3 current=1, got %d %d", limit, current)
	}
}

func TestConcurrencyManager_MultipleGroups(t *testing.T) {
	m := NewConcurrencyManager()
	_ = m.Acquire("deploy", 1)
	if !m.Acquire("test", 1) {
		t.Fatal("groups should be independent")
	}
}

func TestConcurrencyManager_ConcurrentAcquire(t *testing.T) {
	m := NewConcurrencyManager()
	var okCount atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if m.Acquire("deploy", 5) {
				okCount.Add(1)
			}
		}()
	}
	wg.Wait()
	if okCount.Load() != 5 {
		t.Fatalf("want 5 acquires, got %d", okCount.Load())
	}
}
