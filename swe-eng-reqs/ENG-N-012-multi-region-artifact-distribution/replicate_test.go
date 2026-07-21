package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestReplicator_PutPrimary_GetSecondaryMissing(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, 100*time.Millisecond)
	digest, err := r.Put("us-east", []byte("artifact"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Get("eu-west", digest); ok {
		t.Fatal("secondary GET succeeded before replication")
	}
	r.Wait()
}

func TestReplicator_LagPositiveThenZero(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, 75*time.Millisecond)
	digest, err := r.Put("us-east", []byte("lag proof"))
	if err != nil {
		t.Fatal(err)
	}
	if got := r.LagMs(); got <= 0 {
		t.Fatalf("LagMs() = %d, want > 0", got)
	}
	r.Wait()
	if got := r.LagMs(); got != 0 {
		t.Fatalf("LagMs() after Wait = %d", got)
	}
	if _, ok := r.Get("eu-west", digest); !ok {
		t.Fatal("secondary GET missing after replication")
	}
}

func TestReplicator_Status(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, 100*time.Millisecond)
	if _, err := r.Put("us-east", []byte("status")); err != nil {
		t.Fatal(err)
	}
	status := r.Status()
	if len(status.Regions) != 2 || status.Pending != 1 || status.LagMs <= 0 {
		t.Fatalf("Status() = %+v", status)
	}
	r.Wait()
}

func TestReplicator_RegionsAtLeastTwo(t *testing.T) {
	r := NewReplicator([]string{"only"}, time.Millisecond)
	if len(r.Status().Regions) < 2 {
		t.Fatalf("regions = %v, want at least two", r.Status().Regions)
	}
}

func TestReplicator_ConcurrentPuts(t *testing.T) {
	r := NewReplicator([]string{"us-east", "eu-west"}, time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := r.Put("us-east", []byte(fmt.Sprintf("blob-%d", i))); err != nil {
				t.Errorf("Put() error = %v", err)
			}
		}(i)
	}
	wg.Wait()
	r.Wait()
	if r.Status().Pending != 0 {
		t.Fatalf("pending = %d", r.Status().Pending)
	}
}
