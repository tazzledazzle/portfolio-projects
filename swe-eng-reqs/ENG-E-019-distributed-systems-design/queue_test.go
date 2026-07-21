package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestQueue_Enqueue_PartitionAssigned(t *testing.T) {
	q := NewQueue(4)
	task, err := q.Enqueue("payload-a", "idem-part-1", 2)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if task == nil {
		t.Fatal("expected task")
	}
	if task.Partition != 2 {
		t.Fatalf("expected partition 2, got %d", task.Partition)
	}
	if task.ID == "" {
		t.Fatal("expected non-empty task id")
	}
	if task.Status != "queued" {
		t.Fatalf("expected status queued, got %s", task.Status)
	}
}

func TestQueue_Idempotency_DuplicateSuppressed(t *testing.T) {
	q := NewQueue(4)
	sideEffects := 0
	t1, err := q.Enqueue("once", "same-key", 0)
	if err != nil {
		t.Fatalf("first Enqueue: %v", err)
	}
	sideEffects++
	t2, err := q.Enqueue("again", "same-key", 0)
	if err != nil {
		t.Fatalf("second Enqueue: %v", err)
	}
	if !t2.DuplicateSuppressed {
		t.Fatal("expected duplicate_suppressed on second enqueue")
	}
	if t2.ID != t1.ID {
		t.Fatalf("expected same task id, got %s vs %s", t2.ID, t1.ID)
	}
	if sideEffects != 1 {
		t.Fatalf("side effect should run once, got %d", sideEffects)
	}
	stats := q.Stats()
	if stats["duplicate_suppressed"].(int) < 1 {
		t.Fatalf("expected duplicate_suppressed count >= 1, got %#v", stats["duplicate_suppressed"])
	}
}

func TestQueue_Retry_IncrementsAttempts(t *testing.T) {
	q := NewQueue(4)
	task, err := q.Enqueue("retry-me", "retry-key", 1)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	claimed := q.Claim()
	if claimed == nil || claimed.ID != task.ID {
		t.Fatalf("expected to claim task %s, got %#v", task.ID, claimed)
	}
	if err := q.Nack(claimed.ID); err != nil {
		t.Fatalf("Nack: %v", err)
	}
	after := q.Get(claimed.ID)
	if after == nil {
		t.Fatal("task missing after nack")
	}
	if after.Attempts < 1 {
		t.Fatalf("expected attempts >= 1 after nack, got %d", after.Attempts)
	}
	if after.Status != "queued" {
		t.Fatalf("expected requeued after nack, got %s", after.Status)
	}
	if err := q.Ack(claimed.ID); err != nil {
		t.Fatalf("Ack after retry: %v", err)
	}
	done := q.Get(claimed.ID)
	if done.Status != "acked" {
		t.Fatalf("expected acked, got %s", done.Status)
	}
}

func TestQueue_ConcurrentEnqueueAck(t *testing.T) {
	q := NewQueue(8)
	var wg sync.WaitGroup
	var acked atomic.Int64
	n := 50
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			key := "ck-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
			_, _ = q.Enqueue("body", key, i%8)
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				if task := q.Claim(); task != nil {
					if err := q.Ack(task.ID); err == nil {
						acked.Add(1)
					}
				}
			}
		}()
	}
	wg.Wait()
	stats := q.Stats()
	if stats["enqueued"].(int) < 1 {
		t.Fatalf("expected enqueues, got %#v", stats)
	}
	_ = acked.Load()
}
