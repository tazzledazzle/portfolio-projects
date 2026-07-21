package main

import (
	"fmt"
	"sync"
)

// Task is a queued unit of work with partition and idempotency metadata.
type Task struct {
	ID                  string `json:"id"`
	Payload             string `json:"payload"`
	IdempotencyKey      string `json:"idempotency_key"`
	Partition           int    `json:"partition"`
	Status              string `json:"status"`
	Attempts            int    `json:"attempts"`
	DuplicateSuppressed bool   `json:"duplicate_suppressed,omitempty"`
}

// Queue is an in-memory partitioned task queue with idempotency keys (D-03 E-019).
// Does NOT own event-bus DLQ (E-020), durable workflows (H-006), or load SLO sim (E-009).
type Queue struct {
	mu           sync.Mutex
	partitions   int
	tasks        map[string]*Task
	byIdem       map[string]string // idempotency key → task id
	queues       [][]*Task         // per-partition FIFO of task ids (via Task pointers)
	counter      int
	dupSuppressed int
	enqueued     int
	acked        int
	nacked       int
}

func NewQueue(partitions int) *Queue {
	if partitions <= 0 {
		partitions = 4
	}
	return &Queue{
		partitions: partitions,
		tasks:      map[string]*Task{},
		byIdem:     map[string]string{},
		queues:     make([][]*Task, partitions),
	}
}

// Enqueue adds a task. Same idempotencyKey returns the existing task with DuplicateSuppressed.
func (q *Queue) Enqueue(payload, idempotencyKey string, partitionHint int) (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if idempotencyKey != "" {
		if id, ok := q.byIdem[idempotencyKey]; ok {
			existing := q.tasks[id]
			dup := *existing
			dup.DuplicateSuppressed = true
			q.dupSuppressed++
			return &dup, nil
		}
	}

	part := partitionHint
	if part < 0 {
		part = 0
	}
	part = part % q.partitions

	q.counter++
	id := fmt.Sprintf("task-%d", q.counter)
	t := &Task{
		ID:             id,
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
		Partition:      part,
		Status:         "queued",
		Attempts:       0,
	}
	q.tasks[id] = t
	if idempotencyKey != "" {
		q.byIdem[idempotencyKey] = id
	}
	q.queues[part] = append(q.queues[part], t)
	q.enqueued++
	out := *t
	return &out, nil
}

// Claim dequeues the next queued task across partitions (round-robin by index).
func (q *Queue) Claim() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := 0; i < q.partitions; i++ {
		if len(q.queues[i]) == 0 {
			continue
		}
		t := q.queues[i][0]
		q.queues[i] = q.queues[i][1:]
		t.Status = "in_flight"
		out := *t
		return &out
	}
	return nil
}

// Nack marks failure, increments attempts, and requeues the task.
func (q *Queue) Nack(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	t.Attempts++
	t.Status = "queued"
	q.queues[t.Partition] = append(q.queues[t.Partition], t)
	q.nacked++
	return nil
}

// Ack marks the task successfully processed.
func (q *Queue) Ack(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	t.Status = "acked"
	q.acked++
	return nil
}

// Get returns a copy of the task or nil.
func (q *Queue) Get(id string) *Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.tasks[id]
	if !ok {
		return nil
	}
	out := *t
	return &out
}

// Stats returns queue counters for demos and metrics.
func (q *Queue) Stats() map[string]any {
	q.mu.Lock()
	defer q.mu.Unlock()
	depth := 0
	for _, qq := range q.queues {
		depth += len(qq)
	}
	return map[string]any{
		"enqueued":             q.enqueued,
		"acked":                q.acked,
		"nacked":               q.nacked,
		"duplicate_suppressed": q.dupSuppressed,
		"partitions":           q.partitions,
		"queue_depth":          depth,
		"tasks":                len(q.tasks),
	}
}

// Info describes ownership for /v1/info.
func (q *Queue) Info() map[string]any {
	return map[string]any{
		"requirement_id": "ENG-E-019",
		"service":        "eng-e-019",
		"title":          "Distributed systems design — task queue",
		"owns":           []string{"retries", "idempotency_keys", "partitions"},
		"does_not_own":   []string{"event_bus_dlq", "durable_workflows", "load_slo_sim"},
	}
}
