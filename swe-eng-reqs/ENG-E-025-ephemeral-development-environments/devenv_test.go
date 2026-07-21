package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDevEnv_Create_Ready(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	c := NewController()
	env, err := c.Create("dev-1", 60, now)
	if err != nil {
		t.Fatal(err)
	}
	if !conditionTrue(env, "Ready") || conditionTrue(env, "Expired") || env.Reclaimed {
		t.Fatalf("created env = %+v", env)
	}
}

func TestDevEnv_TickPastTTL_Expired(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	c := NewController()
	_, _ = c.Create("dev-1", 60, now)
	env, err := c.Tick("dev-1", now.Add(61*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if conditionTrue(env, "Ready") || !conditionTrue(env, "Expired") || !env.Reclaimed {
		t.Fatalf("expired env = %+v", env)
	}
}

func TestDevEnv_BeforeTTL_StillReady(t *testing.T) {
	now := time.Now().UTC()
	c := NewController()
	_, _ = c.Create("dev-1", 60, now)
	env, err := c.Tick("dev-1", now.Add(59*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if !conditionTrue(env, "Ready") || conditionTrue(env, "Expired") || env.Reclaimed {
		t.Fatalf("env before TTL = %+v", env)
	}
}

func TestDevEnv_ReconcileAll(t *testing.T) {
	now := time.Now().UTC()
	c := NewController()
	_, _ = c.Create("short", 10, now)
	_, _ = c.Create("long", 100, now)
	if reclaimed := c.Reconcile(now.Add(11 * time.Second)); reclaimed != 1 {
		t.Fatalf("Reconcile() = %d, want 1", reclaimed)
	}
	short, _ := c.Get("short")
	long, _ := c.Get("long")
	if !short.Reclaimed || long.Reclaimed {
		t.Fatalf("short=%+v long=%+v", short, long)
	}
}

func TestDevEnv_ConcurrentReconcile(t *testing.T) {
	now := time.Now().UTC()
	c := NewController()
	for i := 0; i < 32; i++ {
		_, _ = c.Create(fmt.Sprintf("dev-%d", i), 1, now)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Reconcile(now.Add(2 * time.Second))
		}()
	}
	wg.Wait()
	for i := 0; i < 32; i++ {
		env, _ := c.Get(fmt.Sprintf("dev-%d", i))
		if !env.Reclaimed {
			t.Fatalf("%s not reclaimed", env.ID)
		}
	}
}
