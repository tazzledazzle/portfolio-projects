package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestRunner_ClaimJob_Available(t *testing.T) {
	pool := NewRunnerPool(8)
	pool.Enqueue(&Job{ID: "j1", Status: "queued", Matrix: map[string]string{"os": "ubuntu"}})
	_ = pool.Register("r1")
	job := pool.ClaimJob("r1")
	if job == nil || job.ID != "j1" {
		t.Fatalf("expected j1, got %#v", job)
	}
}

func TestRunner_ClaimJob_Empty(t *testing.T) {
	pool := NewRunnerPool(8)
	_ = pool.Register("r1")
	if job := pool.ClaimJob("r1"); job != nil {
		t.Fatalf("expected nil, got %#v", job)
	}
}

func TestRunner_ReportComplete_Success(t *testing.T) {
	pool := NewRunnerPool(8)
	pool.Enqueue(&Job{ID: "j1", Status: "queued"})
	_ = pool.Register("r1")
	_ = pool.ClaimJob("r1")
	if err := pool.ReportComplete("j1", "success"); err != nil {
		t.Fatal(err)
	}
	if pool.jobs["j1"].Status != "success" {
		t.Fatalf("status=%s", pool.jobs["j1"].Status)
	}
}

func TestRunner_ReportComplete_Failure(t *testing.T) {
	pool := NewRunnerPool(8)
	pool.Enqueue(&Job{ID: "j1", Status: "queued"})
	_ = pool.Register("r1")
	_ = pool.ClaimJob("r1")
	if err := pool.ReportComplete("j1", "failure"); err != nil {
		t.Fatal(err)
	}
	if pool.jobs["j1"].Status != "failure" {
		t.Fatalf("status=%s", pool.jobs["j1"].Status)
	}
}

func TestRunner_InjectSecrets(t *testing.T) {
	env := map[string]string{"PATH": "/usr/bin"}
	masked := InjectSecrets(env, map[string]string{"TOKEN": "supersecret"})
	if masked["TOKEN"] != "***" {
		t.Fatalf("want masked ***, got %q", masked["TOKEN"])
	}
	if env["TOKEN"] != "supersecret" {
		t.Fatalf("env should hold real secret, got %q", env["TOKEN"])
	}
}

func TestRunner_ConcurrentClaim(t *testing.T) {
	pool := NewRunnerPool(64)
	for i := 0; i < 10; i++ {
		pool.Enqueue(&Job{ID: fmt.Sprintf("j%d", i), Status: "queued"})
	}
	var wg sync.WaitGroup
	claimed := make(chan string, 20)
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("r%d", i)
		_ = pool.Register(id)
		wg.Add(1)
		go func(runnerID string) {
			defer wg.Done()
			if j := pool.ClaimJob(runnerID); j != nil {
				claimed <- j.ID
			}
		}(id)
	}
	wg.Wait()
	close(claimed)
	seen := map[string]bool{}
	for id := range claimed {
		if seen[id] {
			t.Fatalf("job %s claimed twice", id)
		}
		seen[id] = true
	}
	if len(seen) != 10 {
		t.Fatalf("want 10 unique claims, got %d", len(seen))
	}
}
