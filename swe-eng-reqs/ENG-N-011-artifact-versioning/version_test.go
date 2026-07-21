package main

import (
	"strings"
	"sync"
	"testing"
)

const (
	digestA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	digestB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	bareHex = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestVersion_Create(t *testing.T) {
	store := NewVersionStore()
	v, err := store.Create("app", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if v.ID == "" {
		t.Fatal("Create: expected non-empty ID")
	}
	if v.Stage != "dev" {
		t.Fatalf("Create stage=%q, want dev", v.Stage)
	}
	if v.Digest != digestA {
		t.Fatalf("Create digest=%q, want %q", v.Digest, digestA)
	}
	if !strings.HasPrefix(v.Digest, "sha256:") {
		t.Fatalf("digest must be sha256: form, got %q", v.Digest)
	}
}

func TestPromote_DevToStaging(t *testing.T) {
	store := NewVersionStore()
	v, err := store.Create("svc", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	promoted, err := store.Promote(v.ID)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if promoted.Stage != "staging" {
		t.Fatalf("stage=%q, want staging", promoted.Stage)
	}
	if promoted.Digest != digestA {
		t.Fatalf("digest mutated: got %q want %q", promoted.Digest, digestA)
	}
}

func TestPromote_StagingToProd(t *testing.T) {
	store := NewVersionStore()
	v, err := store.Create("svc", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Promote(v.ID); err != nil {
		t.Fatalf("Promote to staging: %v", err)
	}
	promoted, err := store.Promote(v.ID)
	if err != nil {
		t.Fatalf("Promote to prod: %v", err)
	}
	if promoted.Stage != "prod" {
		t.Fatalf("stage=%q, want prod", promoted.Stage)
	}
	if promoted.Digest != digestA {
		t.Fatalf("digest mutated: got %q want %q", promoted.Digest, digestA)
	}
}

func TestPromote_InvalidSkip(t *testing.T) {
	store := NewVersionStore()
	// Creating at staging is allowed; skipping staging via promote from a fake jump is rejected
	// by linear-only promote: from prod cannot promote further, and Create with invalid stage fails.
	_, err := store.Create("svc", digestA, "prod")
	if err == nil {
		// if create at prod is allowed for seeding, promote further must fail
		v, _ := store.Create("svc2", digestB, "dev")
		// simulate illegal skip: try to set stage somehow — Promote only advances one step
		_ = v
	}
	v, err := store.Create("linear", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// After one promote we are staging; a second promote goes to prod (linear).
	// Invalid skip means we never jump dev→prod in a single Promote call.
	promoted, err := store.Promote(v.ID)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if promoted.Stage == "prod" {
		t.Fatal("Promote from dev must not skip to prod (linear stages only)")
	}
	if promoted.Stage != "staging" {
		t.Fatalf("stage=%q, want staging after first promote", promoted.Stage)
	}
	// At prod, further promote must error
	if _, err := store.Promote(v.ID); err != nil {
		t.Fatalf("second promote: %v", err)
	}
	if _, err := store.Promote(v.ID); err == nil {
		t.Fatal("Promote from prod: expected error")
	}
}

func TestTag_Mutable(t *testing.T) {
	store := NewVersionStore()
	v, err := store.Create("tagged", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.SetTag("release", digestA); err != nil {
		t.Fatalf("SetTag d1: %v", err)
	}
	if err := store.SetTag("release", digestB); err != nil {
		t.Fatalf("SetTag d2: %v", err)
	}
	got, ok := store.GetTag("release")
	if !ok || got != digestB {
		t.Fatalf("GetTag=%q ok=%v, want %q", got, ok, digestB)
	}
	cur, err := store.Get(v.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cur.Digest != digestA {
		t.Fatalf("version digest mutated by tag: got %q", cur.Digest)
	}
}

func TestVersion_RejectBadDigest(t *testing.T) {
	store := NewVersionStore()
	if _, err := store.Create("bad", "..", "dev"); err == nil {
		t.Fatal("Create with '..' digest: expected error")
	}
	if _, err := store.Create("badname..", digestA, "dev"); err == nil {
		t.Fatal("Create with '..' in name: expected error")
	}
	v, err := store.Create("bare", bareHex, "dev")
	if err != nil {
		t.Fatalf("Create bare hex: %v", err)
	}
	if v.Digest != "sha256:"+bareHex {
		t.Fatalf("bare hex not normalized: got %q", v.Digest)
	}
	if _, err := store.Create("short", "abc", "dev"); err == nil {
		t.Fatal("Create with short digest: expected error")
	}
}

func TestVersion_ConcurrentPromote(t *testing.T) {
	store := NewVersionStore()
	v, err := store.Create("race", digestA, "dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Promote(v.ID)
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	cur, err := store.Get(v.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cur.Digest != digestA {
		t.Fatalf("concurrent promote mutated digest: %q", cur.Digest)
	}
	if cur.Stage != "staging" && cur.Stage != "prod" {
		t.Fatalf("unexpected stage after concurrent promote: %q", cur.Stage)
	}
}
