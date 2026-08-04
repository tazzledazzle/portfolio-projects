package p1

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestBruteForce_ConcurrentDuplicateKeysDeployMoreThanOnce(t *testing.T) {
	var mu sync.Mutex
	var calls int
	svc := NewBruteForceDeployService(func(sn, sha string) error {
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	})

	const n = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start // block until released, so all goroutines race together
			svc.Trigger("same-key", "checkout-service", "abc123")
		}()
	}
	close(start) // release all goroutines at once
	wg.Wait()

	if calls <= 1 {
		t.Fatalf("expected the brute-force version to double-deploy under a race, got %d calls", calls)
	}
}

func newOptimalForTest(t *testing.T, deploy DeployFunc) *OptimalDeployService {
	t.Helper()
	path := fmt.Sprintf("/tmp/p1_test_%d.wal", time.Now().UnixNano())
	svc, err := NewOptimalDeployService(deploy, path, 300*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		svc.Close()
		os.Remove(path)
	})
	return svc
}

func TestOptimal_DuplicateKeyDeploysExactlyOnceUnderConcurrency(t *testing.T) {
	var mu sync.Mutex
	var calls int
	svc := newOptimalForTest(t, func(sn, sha string) error {
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	})

	const n = 50
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			svc.Trigger("same-key", "checkout-service", "abc123")
		}()
	}
	close(start)
	wg.Wait()

	if calls != 1 {
		t.Fatalf("expected exactly 1 deploy call, got %d", calls)
	}
}

func TestOptimal_DifferentKeysDeployIndependently(t *testing.T) {
	svc := newOptimalForTest(t, func(sn, sha string) error { return nil })
	r1 := svc.Trigger("key-a", "svc-a", "sha1")
	r2 := svc.Trigger("key-b", "svc-b", "sha2")
	if r1.IdempotencyKey == r2.IdempotencyKey {
		t.Fatal("expected distinct records for distinct keys")
	}
}

func TestOptimal_UnknownKeyStatusLookupMisses(t *testing.T) {
	svc := newOptimalForTest(t, func(sn, sha string) error { return nil })
	_, ok := svc.GetStatus("never-seen")
	if ok {
		t.Fatal("expected no record for an unseen key")
	}
}

func TestOptimal_FailedDeployIsRecordedAndNotRetried(t *testing.T) {
	var calls int
	svc := newOptimalForTest(t, func(sn, sha string) error {
		calls++
		return fmt.Errorf("deploy target unreachable")
	})
	first := svc.Trigger("bad-key", "svc", "sha")
	if first.Status != StatusFailed {
		t.Fatalf("expected FAILED, got %s", first.Status)
	}
	second := svc.Trigger("bad-key", "svc", "sha")
	if second.Status != StatusFailed {
		t.Fatalf("expected FAILED on duplicate, got %s", second.Status)
	}
	if calls != 1 {
		t.Fatalf("expected deploy fn called exactly once even though it failed, got %d calls", calls)
	}
}

func TestOptimal_CrashMidDeployReconciles(t *testing.T) {
	svc := newOptimalForTest(t, func(sn, sha string) error { return nil })
	// Simulate: a RUNNING record whose last update is already older than
	// the stale timeout, as if the process died before resolving it.
	svc.injectRawRunningRecord("crashed-key", "svc", "sha", time.Now().Add(-time.Second))

	before, _ := svc.GetStatus("crashed-key")
	if before.Status != StatusRunning {
		t.Fatalf("expected RUNNING before reconciliation, got %s", before.Status)
	}

	svc.SimulateCrashAndRestart()

	after, _ := svc.GetStatus("crashed-key")
	if after.Status != StatusUnknown {
		t.Fatalf("expected UNKNOWN after reconciliation of a stale RUNNING record, got %s", after.Status)
	}
}

func TestOptimal_RecentRunningRecordIsNotFalselyReconciled(t *testing.T) {
	svc := newOptimalForTest(t, func(sn, sha string) error { return nil })
	svc.injectRawRunningRecord("in-flight-key", "svc", "sha", time.Now())
	svc.SimulateCrashAndRestart()
	after, _ := svc.GetStatus("in-flight-key")
	if after.Status != StatusRunning {
		t.Fatalf("a freshly-RUNNING record must not be falsely reconciled, got %s", after.Status)
	}
}

func TestOptimal_DurabilityAcrossRestart(t *testing.T) {
	path := fmt.Sprintf("/tmp/p1_test_durability_%d.wal", time.Now().UnixNano())
	defer os.Remove(path)

	svc1, err := NewOptimalDeployService(func(sn, sha string) error { return nil }, path, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc1.Trigger("durable-key", "svc", "sha")
	svc1.Close()

	// "Process restarts": brand-new instance, same WAL file.
	svc2, err := NewOptimalDeployService(func(sn, sha string) error {
		t.Fatal("deploy fn must NOT be called again after a restart for an already-resolved key")
		return nil
	}, path, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer svc2.Close()

	r, ok := svc2.GetStatus("durable-key")
	if !ok {
		t.Fatal("expected the record to survive the restart via WAL replay")
	}
	if r.Status != StatusSucceeded {
		t.Fatalf("expected SUCCEEDED to have survived the restart, got %s", r.Status)
	}

	// Duplicate trigger after restart must not re-deploy.
	r2 := svc2.Trigger("durable-key", "svc", "sha")
	if r2.Status != StatusSucceeded {
		t.Fatalf("expected SUCCEEDED on post-restart duplicate trigger, got %s", r2.Status)
	}
}
