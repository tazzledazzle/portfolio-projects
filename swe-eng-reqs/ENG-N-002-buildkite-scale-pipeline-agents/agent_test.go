package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestAgent_Poll_NoJobs(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	if job := pool.Poll("a1"); job != nil {
		t.Fatalf("expected nil, got %#v", job)
	}
}

func TestAgent_Poll_HasJob(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	pool.Enqueue(&BKJob{ID: "j1", Pipeline: "p", Stage: "build", Status: "waiting"})
	pool.Enqueue(&BKJob{ID: "j2", Pipeline: "p", Stage: "test", Status: "waiting"})
	job := pool.Poll("a1")
	if job == nil || job.ID != "j1" {
		t.Fatalf("want oldest j1, got %#v", job)
	}
}

func TestAgent_Claim_Success(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	pool.Enqueue(&BKJob{ID: "j1", Status: "waiting"})
	_ = pool.Poll("a1")
	if err := pool.Claim("a1", "j1"); err != nil {
		t.Fatal(err)
	}
	if pool.jobs["j1"].Status != "running" {
		t.Fatalf("status=%s", pool.jobs["j1"].Status)
	}
}

func TestAgent_Claim_AlreadyClaimed(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	_ = pool.Register("a2", nil)
	pool.Enqueue(&BKJob{ID: "j1", Status: "waiting"})
	_ = pool.Poll("a1")
	_ = pool.Claim("a1", "j1")
	if err := pool.Claim("a2", "j1"); err == nil {
		t.Fatal("expected already claimed error")
	}
}

func TestAgent_Complete_Success(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	pool.Enqueue(&BKJob{ID: "j1", Status: "waiting"})
	_ = pool.Poll("a1")
	_ = pool.Claim("a1", "j1")
	if err := pool.Complete("j1", 0); err != nil {
		t.Fatal(err)
	}
	if pool.jobs["j1"].Status != "succeeded" {
		t.Fatalf("status=%s", pool.jobs["j1"].Status)
	}
}

func TestAgent_Complete_Failure(t *testing.T) {
	pool := NewAgentPool()
	_ = pool.Register("a1", nil)
	pool.Enqueue(&BKJob{ID: "j1", Status: "waiting"})
	_ = pool.Poll("a1")
	_ = pool.Claim("a1", "j1")
	if err := pool.Complete("j1", 1); err != nil {
		t.Fatal(err)
	}
	if pool.jobs["j1"].Status != "failed" {
		t.Fatalf("status=%s", pool.jobs["j1"].Status)
	}
}

func TestAgent_DynamicPipeline(t *testing.T) {
	pool := NewAgentPool()
	yaml := "steps:\n  - label: lint\n  - label: test\n  - label: build\n"
	if err := pool.UploadPipeline("pipe-1", yaml); err != nil {
		t.Fatal(err)
	}
	if len(pool.queue) != 3 {
		t.Fatalf("want 3 stages queued, got %d", len(pool.queue))
	}
}

func TestAgent_ConcurrentPoll(t *testing.T) {
	pool := NewAgentPool()
	for i := 0; i < 10; i++ {
		pool.Enqueue(&BKJob{ID: fmt.Sprintf("j%d", i), Status: "waiting"})
		_ = pool.Register(fmt.Sprintf("a%d", i), nil)
	}
	var wg sync.WaitGroup
	got := make(chan string, 20)
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("a%d", i)
		wg.Add(1)
		go func(agentID string) {
			defer wg.Done()
			if j := pool.Poll(agentID); j != nil {
				_ = pool.Claim(agentID, j.ID)
				got <- j.ID
			}
		}(id)
	}
	wg.Wait()
	close(got)
	seen := map[string]bool{}
	for id := range got {
		if seen[id] {
			t.Fatalf("duplicate claim %s", id)
		}
		seen[id] = true
	}
	if len(seen) != 10 {
		t.Fatalf("want 10 jobs, got %d", len(seen))
	}
}
