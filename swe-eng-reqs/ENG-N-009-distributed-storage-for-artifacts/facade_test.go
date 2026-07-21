package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestFacade_PutGet_Quorum(t *testing.T) {
	f := NewFacade(3, 2)
	digest, err := f.Put([]byte("artifact"))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	got, ok := f.Get(digest)
	if !ok || string(got) != "artifact" {
		t.Fatalf("Get() = %q, %v", got, ok)
	}
}

func TestFacade_DurabilityMetadata(t *testing.T) {
	f := NewFacade(3, 2)
	digest, err := f.Put([]byte("durable"))
	if err != nil {
		t.Fatal(err)
	}
	d, ok := f.Durability(digest)
	if !ok {
		t.Fatal("Durability() returned false")
	}
	if d.Replicas < 2 || d.HealthyNodes < 2 {
		t.Fatalf("durability = %+v", d)
	}
	if d.Checksum != digest || !strings.HasPrefix(d.Checksum, "sha256:") {
		t.Fatalf("checksum = %q", d.Checksum)
	}
}

func TestFacade_QuorumFail(t *testing.T) {
	f := NewFacade(3, 3)
	f.SetNodeHealthy(2, false)
	if _, err := f.Put([]byte("cannot reach quorum")); err == nil {
		t.Fatal("Put() error = nil, want quorum error")
	}
}

func TestFacade_DigestFormat(t *testing.T) {
	f := NewFacade(3, 2)
	digest, err := f.Put([]byte("oci digest"))
	if err != nil {
		t.Fatal(err)
	}
	if len(digest) != len("sha256:")+64 || !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("digest = %q", digest)
	}
}

func TestFacade_Concurrent(t *testing.T) {
	f := NewFacade(3, 2)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			want := []byte(fmt.Sprintf("blob-%d", i))
			digest, err := f.Put(want)
			if err != nil {
				t.Errorf("Put() error = %v", err)
				return
			}
			got, ok := f.Get(digest)
			if !ok || string(got) != string(want) {
				t.Errorf("Get() = %q, %v", got, ok)
			}
		}(i)
	}
	wg.Wait()
}
