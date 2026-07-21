package main

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func TestRegistry_PutManifest_Get(t *testing.T) {
	r := NewRegistry()
	manifest := []byte(`{"schemaVersion":2,"config":{}}`)
	digest, err := r.PutManifest(manifest)
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	got, ok := r.GetManifest(digest)
	if !ok {
		t.Fatal("GetManifest: not found")
	}
	if !bytes.Equal(got, manifest) {
		t.Fatalf("GetManifest = %q, want %q", got, manifest)
	}
}

func TestRegistry_PutTag_Resolve(t *testing.T) {
	r := NewRegistry()
	digest, err := r.PutManifest([]byte("manifest-v1"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	if err := r.PutTag("demo", "latest", digest); err != nil {
		t.Fatalf("PutTag: %v", err)
	}
	got, ok := r.Resolve("demo", "latest")
	if !ok || got != digest {
		t.Fatalf("Resolve = %q ok=%v, want %q", got, ok, digest)
	}
}

func TestRegistry_TagRetarget_ManifestUnchanged(t *testing.T) {
	r := NewRegistry()
	oldManifest := []byte("old-manifest")
	newManifest := []byte("new-manifest")
	oldDigest, err := r.PutManifest(oldManifest)
	if err != nil {
		t.Fatalf("PutManifest old: %v", err)
	}
	newDigest, err := r.PutManifest(newManifest)
	if err != nil {
		t.Fatalf("PutManifest new: %v", err)
	}
	if err := r.PutTag("demo", "latest", oldDigest); err != nil {
		t.Fatalf("PutTag old: %v", err)
	}
	if err := r.PutTag("demo", "latest", newDigest); err != nil {
		t.Fatalf("PutTag retarget: %v", err)
	}
	resolved, ok := r.Resolve("demo", "latest")
	if !ok || resolved != newDigest {
		t.Fatalf("Resolve after retarget = %q, want %q", resolved, newDigest)
	}
	got, ok := r.GetManifest(oldDigest)
	if !ok || !bytes.Equal(got, oldManifest) {
		t.Fatalf("old manifest mutated: ok=%v got=%q", ok, got)
	}
}

func TestRegistry_DigestImmutable(t *testing.T) {
	r := NewRegistry()
	digest, err := r.PutManifest([]byte("hello"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	err = r.putManifestAt(digest, []byte("world"))
	if err == nil {
		t.Fatal("want conflict on different bytes same digest")
	}
	if !errors.Is(err, ErrManifestConflict) {
		t.Fatalf("error = %v, want ErrManifestConflict", err)
	}
}

func TestRegistry_RejectUnsafeName(t *testing.T) {
	r := NewRegistry()
	digest, err := r.PutManifest([]byte("m"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	if err := r.PutTag("demo/../evil", "latest", digest); err == nil {
		t.Fatal("PutTag accepted name with ..")
	}
	if err := r.PutTag("demo", "../tag", digest); err == nil {
		t.Fatal("PutTag accepted tag with ..")
	}
}

func TestRegistry_ConcurrentTagResolve(t *testing.T) {
	r := NewRegistry()
	digest, err := r.PutManifest([]byte("concurrent-manifest"))
	if err != nil {
		t.Fatalf("PutManifest: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tag := string(rune('a' + (n % 26)))
			if err := r.PutTag("demo", tag, digest); err != nil {
				t.Errorf("PutTag: %v", err)
				return
			}
			got, ok := r.Resolve("demo", tag)
			if !ok || got != digest {
				t.Errorf("Resolve(%q) = %q ok=%v", tag, got, ok)
			}
		}(i)
	}
	wg.Wait()
}
