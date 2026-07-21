package main

import (
	"bytes"
	"errors"
	"regexp"
	"sync"
	"testing"
)

var digestRE = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func TestStore_PutGet_Roundtrip(t *testing.T) {
	s := NewBlobStore()
	meta := map[string]string{"content-type": "text/plain"}
	digest, err := s.Put([]byte("hello"), meta)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if !digestRE.MatchString(digest) {
		t.Fatalf("digest %q does not match %s", digest, digestRE)
	}
	got, ok := s.Get(digest)
	if !ok {
		t.Fatal("Get: not found")
	}
	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("Get = %q, want %q", got, "hello")
	}
}

func TestStore_Put_IdempotentSameBytes(t *testing.T) {
	s := NewBlobStore()
	d1, err := s.Put([]byte("hello"), nil)
	if err != nil {
		t.Fatalf("Put1: %v", err)
	}
	d2, err := s.Put([]byte("hello"), nil)
	if err != nil {
		t.Fatalf("Put2: %v", err)
	}
	if d1 != d2 {
		t.Fatalf("digests differ: %q vs %q", d1, d2)
	}
}

func TestStore_Put_RejectOverwrite(t *testing.T) {
	s := NewBlobStore()
	digest, err := s.Put([]byte("hello"), nil)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	err = s.putAt(digest, []byte("different"), nil)
	if err == nil {
		t.Fatal("putAt different bytes: want conflict error")
	}
	if !errors.Is(err, ErrDigestConflict) {
		t.Fatalf("putAt error = %v, want ErrDigestConflict", err)
	}
	got, ok := s.Get(digest)
	if !ok || !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("stored bytes mutated: ok=%v got=%q", ok, got)
	}
}

func TestStore_MetadataPersists(t *testing.T) {
	s := NewBlobStore()
	meta := map[string]string{"content-type": "text/plain"}
	digest, err := s.Put([]byte("hello"), meta)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, ok := s.Head(digest)
	if !ok {
		t.Fatal("Head: not found")
	}
	if got["content-type"] != "text/plain" {
		t.Fatalf("metadata content-type = %q, want text/plain", got["content-type"])
	}
}

func TestStore_ValidDigest_RejectsTraversal(t *testing.T) {
	if ValidDigest("sha256:../evil") {
		t.Fatal("ValidDigest accepted path traversal")
	}
	if ValidDigest("deadbeef") {
		t.Fatal("ValidDigest accepted missing sha256: prefix")
	}
	if ValidDigest("sha256:abcd") {
		t.Fatal("ValidDigest accepted short digest")
	}
	good := DigestSHA256([]byte("x"))
	if !ValidDigest(good) {
		t.Fatalf("ValidDigest rejected good digest %q", good)
	}
}

func TestStore_ConcurrentPutGet(t *testing.T) {
	s := NewBlobStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			payload := []byte{byte(n), byte(n + 1), 'x'}
			d, err := s.Put(payload, map[string]string{"n": string(rune('0' + n%10))})
			if err != nil {
				t.Errorf("Put: %v", err)
				return
			}
			got, ok := s.Get(d)
			if !ok || !bytes.Equal(got, payload) {
				t.Errorf("Get mismatch for %q", d)
			}
		}(i)
	}
	wg.Wait()
}
