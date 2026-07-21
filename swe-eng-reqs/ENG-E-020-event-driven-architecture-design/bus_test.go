package main

import (
	"testing"
	"time"
)

func TestBus_Publish_EnvelopeSchema(t *testing.T) {
	b := NewBus()
	env, err := b.Publish("orders.created", "orders.v1", map[string]any{"order_id": "o-1"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if env.ID == "" {
		t.Fatal("expected envelope id")
	}
	if env.Subject != "orders.created" {
		t.Fatalf("subject: got %s", env.Subject)
	}
	if env.Schema != "orders.v1" {
		t.Fatalf("schema: got %s", env.Schema)
	}
	if env.Payload["order_id"] != "o-1" {
		t.Fatalf("payload: %#v", env.Payload)
	}
	if env.TS.IsZero() {
		t.Fatal("expected non-zero TS")
	}
}

func TestBus_Consume_Delivers(t *testing.T) {
	b := NewBus()
	_, _ = b.Publish("ping", "ping.v1", map[string]any{"n": 1})
	got, err := b.Consume("c1")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if got == nil {
		t.Fatal("expected message")
	}
	if got.Subject != "ping" {
		t.Fatalf("subject: %s", got.Subject)
	}
}

func TestBus_DLQ_OnHandlerFailure(t *testing.T) {
	b := NewBus()
	_, _ = b.Publish("fail.me", "fail.v1", map[string]any{"x": true})
	env, err := b.Consume("c-dlq")
	if err != nil || env == nil {
		t.Fatalf("Consume: %v %#v", err, env)
	}
	if err := b.Fail(env.ID, "handler boom"); err != nil {
		t.Fatalf("Fail: %v", err)
	}
	dlq := b.DLQ()
	if len(dlq) < 1 {
		t.Fatal("expected DLQ entry after handler failure")
	}
	if dlq[0].Reason == "" {
		t.Fatal("expected DLQ reason")
	}
}

func TestBus_Replay_FromOffset(t *testing.T) {
	b := NewBus()
	e1, _ := b.Publish("a", "s.v1", map[string]any{"i": 1})
	e2, _ := b.Publish("b", "s.v1", map[string]any{"i": 2})
	_ = e1
	replayed, err := b.Replay(0)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(replayed) < 2 {
		t.Fatalf("expected >=2 replayed, got %d", len(replayed))
	}
	if replayed[1].ID != e2.ID {
		t.Fatalf("expected second id %s, got %s", e2.ID, replayed[1].ID)
	}
	// Replay must not claim live NATS — Info honesty checked separately
	_ = time.Now()
}

func TestBus_Info_NATSInspired(t *testing.T) {
	b := NewBus()
	info := b.Info()
	if info["nats_inspired"] != true {
		t.Fatalf("expected nats_inspired=true, got %#v", info["nats_inspired"])
	}
	if info["simulator"] != true {
		t.Fatalf("expected simulator=true, got %#v", info["simulator"])
	}
	if info["bus"] == "nats" {
		t.Fatal("must not claim bus=nats without disclaimer")
	}
	if info["nats_connected"] == true {
		t.Fatal("must not claim live NATS connectivity")
	}
}
