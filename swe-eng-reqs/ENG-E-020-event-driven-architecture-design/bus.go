package main

import (
	"fmt"
	"sync"
	"time"
)

// Envelope is a NATS-inspired message schema (in-process simulator).
type Envelope struct {
	ID      string         `json:"id"`
	Subject string         `json:"subject"`
	Schema  string         `json:"schema"`
	Payload map[string]any `json:"payload"`
	TS      time.Time      `json:"ts"`
	Offset  int            `json:"offset"`
}

// DLQEntry records a failed consumer handling.
type DLQEntry struct {
	Envelope Envelope `json:"envelope"`
	Reason   string   `json:"reason"`
	FailedAt time.Time `json:"failed_at"`
}

// Bus is an in-memory NATS-inspired event bus (D-10, D-11).
// Stdlib only — no nats.go. Label nats_inspired + simulator; never claim live NATS.
type Bus struct {
	mu       sync.Mutex
	log      []Envelope
	pending  []*Envelope // undelivered to consumers
	dlq      []DLQEntry
	counter  int
	consumed map[string]bool // envelope id consumed once from pending
}

func NewBus() *Bus {
	return &Bus{
		log:      []Envelope{},
		pending:  []*Envelope{},
		dlq:      []DLQEntry{},
		consumed: map[string]bool{},
	}
}

func (b *Bus) Publish(subject, schema string, payload map[string]any) (*Envelope, error) {
	if subject == "" {
		return nil, fmt.Errorf("subject required")
	}
	if schema == "" {
		return nil, fmt.Errorf("schema required")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counter++
	env := Envelope{
		ID:      fmt.Sprintf("msg-%d", b.counter),
		Subject: subject,
		Schema:  schema,
		Payload: payload,
		TS:      time.Now().UTC(),
		Offset:  len(b.log),
	}
	b.log = append(b.log, env)
	cp := env
	b.pending = append(b.pending, &cp)
	out := env
	return &out, nil
}

func (b *Bus) Consume(consumerID string) (*Envelope, error) {
	_ = consumerID
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.pending) == 0 {
		return nil, nil
	}
	env := b.pending[0]
	b.pending = b.pending[1:]
	b.consumed[env.ID] = true
	out := *env
	return &out, nil
}

// Fail moves a previously consumed/published message into the DLQ (handler failure).
func (b *Bus) Fail(envelopeID, reason string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	var found *Envelope
	for i := range b.log {
		if b.log[i].ID == envelopeID {
			e := b.log[i]
			found = &e
			break
		}
	}
	if found == nil {
		return fmt.Errorf("envelope not found")
	}
	if reason == "" {
		reason = "handler failure"
	}
	b.dlq = append(b.dlq, DLQEntry{
		Envelope: *found,
		Reason:   reason,
		FailedAt: time.Now().UTC(),
	})
	return nil
}

func (b *Bus) DLQ() []DLQEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]DLQEntry, len(b.dlq))
	copy(out, b.dlq)
	return out
}

// Replay redelivers from the durable-style in-memory log offset (not live NATS).
func (b *Bus) Replay(fromOffset int) ([]Envelope, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if fromOffset < 0 {
		fromOffset = 0
	}
	if fromOffset > len(b.log) {
		return nil, fmt.Errorf("offset out of range")
	}
	out := make([]Envelope, 0, len(b.log)-fromOffset)
	for i := fromOffset; i < len(b.log); i++ {
		out = append(out, b.log[i])
	}
	return out, nil
}

func (b *Bus) Info() map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return map[string]any{
		"requirement_id":  "ENG-E-020",
		"service":         "eng-e-020",
		"title":           "Event-driven architecture — NATS-inspired bus",
		"nats_inspired":   true,
		"simulator":       true,
		"nats_connected":  false,
		"broker":          "none",
		"log_length":      len(b.log),
		"dlq_depth":       len(b.dlq),
		"pending":         len(b.pending),
		"note":            "In-process stdlib NATS-inspired simulator; does NOT connect to a NATS server or JetStream",
		"owns":            []string{"envelopes", "consumers", "dlq", "replay"},
		"does_not_own":    []string{"queue_idempotency", "durable_workflow_engine", "rest_openapi"},
	}
}
