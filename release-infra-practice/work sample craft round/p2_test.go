package p1

import (
	"sync"
	"testing"
	"time"
)

func TestBruteForce_ConcurrentRequestsCanExceedLimit(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewBruteForceRateLimiter(audit)

	const n = 10
	start := make(chan struct{})
	results := make([]RollbackResult, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			results[i] = limiter.TryRollback("checkout-service", time.Now())
		}()
	}
	close(start)
	wg.Wait()

	accepted := 0
	for _, r := range results {
		if r.Accepted {
			accepted++
		}
	}
	if accepted <= MaxRollbacksPerWindow {
		t.Fatalf("expected brute force to let more than %d through under a race, got %d",
			MaxRollbacksPerWindow, accepted)
	}
}

func TestOptimal_AllowsUpToTheLimit(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	for i := 0; i < MaxRollbacksPerWindow; i++ {
		r := limiter.TryRollback("svc-a", base.Add(time.Duration(i)*time.Second))
		if !r.Accepted {
			t.Fatalf("rollback #%d should be accepted", i+1)
		}
	}
}

func TestOptimal_RejectsOverLimitWithRetryAfter(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	for i := 0; i < MaxRollbacksPerWindow; i++ {
		limiter.TryRollback("svc-a", base.Add(time.Duration(i)*time.Second))
	}
	r := limiter.TryRollback("svc-a", base.Add(MaxRollbacksPerWindow*time.Second))
	if r.Accepted {
		t.Fatal("expected rejection over the limit")
	}
	if r.RetryAfterSeconds <= 0 {
		t.Fatalf("expected positive Retry-After, got %f", r.RetryAfterSeconds)
	}
}

func TestOptimal_WindowBoundaryExactExpiry(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	limiter.TryRollback("svc-a", base)
	limiter.TryRollback("svc-a", base.Add(1*time.Second))
	limiter.TryRollback("svc-a", base.Add(2*time.Second))

	r := limiter.TryRollback("svc-a", base.Add(WindowDuration))
	if !r.Accepted {
		t.Fatal("the oldest entry should have aged out of the window by now")
	}
}

func TestOptimal_OneSecondBeforeExpiryStillRejected(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	limiter.TryRollback("svc-a", base)
	limiter.TryRollback("svc-a", base.Add(1*time.Second))
	limiter.TryRollback("svc-a", base.Add(2*time.Second))

	r := limiter.TryRollback("svc-a", base.Add(WindowDuration-time.Second))
	if r.Accepted {
		t.Fatal("expected rejection one second before the oldest entry expires")
	}
}

func TestOptimal_ServicesHaveIndependentLimits(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	for i := 0; i < MaxRollbacksPerWindow; i++ {
		limiter.TryRollback("svc-a", base.Add(time.Duration(i)*time.Second))
	}
	r := limiter.TryRollback("svc-b", base)
	if !r.Accepted {
		t.Fatal("svc-b should be unaffected by svc-a hitting its limit")
	}
}

func TestOptimal_ConcurrentRequestsNeverExceedLimit(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)

	const n = 20
	start := make(chan struct{})
	results := make([]RollbackResult, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			results[i] = limiter.TryRollback("svc-concurrent", time.Now())
		}()
	}
	close(start)
	wg.Wait()

	accepted := 0
	for _, r := range results {
		if r.Accepted {
			accepted++
		}
	}
	if accepted != MaxRollbacksPerWindow {
		t.Fatalf("expected exactly %d accepted, got %d", MaxRollbacksPerWindow, accepted)
	}
}

func TestOptimal_AcceptedRollbacksAreAuditLogged(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	for i := 0; i < MaxRollbacksPerWindow; i++ {
		limiter.TryRollback("svc-audit", base.Add(time.Duration(i)*time.Second))
	}
	limiter.TryRollback("svc-audit", base.Add(MaxRollbacksPerWindow*time.Second)) // rejected

	count := 0
	for _, e := range audit.AllEntries() {
		if e.ServiceName == "svc-audit" {
			count++
		}
	}
	if count != MaxRollbacksPerWindow {
		t.Fatalf("expected only accepted rollbacks audit-logged, got %d entries", count)
	}
}

func TestOptimal_RetryAfterShrinksOverTime(t *testing.T) {
	audit := NewAppendOnlyAuditLog()
	limiter := NewOptimalRateLimiter(audit)
	base := time.Now()
	for i := 0; i < MaxRollbacksPerWindow; i++ {
		limiter.TryRollback("svc-a", base.Add(time.Duration(i)*time.Second))
	}
	r1 := limiter.TryRollback("svc-a", base.Add(MaxRollbacksPerWindow*time.Second))
	r2 := limiter.TryRollback("svc-a", base.Add((MaxRollbacksPerWindow+30)*time.Second))
	if !(r1.RetryAfterSeconds > r2.RetryAfterSeconds) {
		t.Fatalf("expected retry-after to shrink over time: r1=%f r2=%f", r1.RetryAfterSeconds, r2.RetryAfterSeconds)
	}
}
