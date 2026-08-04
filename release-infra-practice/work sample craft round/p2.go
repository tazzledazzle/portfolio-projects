// Package p2 implements the rate-limited rollback endpoint problem: at most
// 3 rollbacks per service per 10-minute sliding window, 429 + Retry-After
// over the limit, every ACCEPTED rollback written to an append-only audit
// log.
package p1

import (
	"container/list"
	"sync"
	"time"
)

const (
	WindowDuration        = 10 * time.Minute
	MaxRollbacksPerWindow = 3
)

type RollbackResult struct {
	Accepted          bool
	RetryAfterSeconds float64
	AuditID           string
}

type AuditEntry struct {
	ID          string
	ServiceName string
	RequestedAt time.Time
}

// AppendOnlyAuditLog: durable, append-only. Production backing would be a
// file opened O_APPEND+fsync, or an INSERT-only DB table. The interesting
// design decisions (what gets logged, and when relative to the rate-limit
// decision) are independent of the storage backend, so an in-memory slice
// behind a mutex is enough to demonstrate them.
type AppendOnlyAuditLog struct {
	mu      sync.Mutex
	entries []AuditEntry
	nextID  int
}

func NewAppendOnlyAuditLog() *AppendOnlyAuditLog {
	return &AppendOnlyAuditLog{}
}

func (l *AppendOnlyAuditLog) Append(serviceName string, requestedAt time.Time) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextID++
	id := time.Now().Format("150405.000000000") + "-" + serviceName
	l.entries = append(l.entries, AuditEntry{ID: id, ServiceName: serviceName, RequestedAt: requestedAt})
	return id
}

func (l *AppendOnlyAuditLog) AllEntries() []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]AuditEntry, len(l.entries))
	copy(out, l.entries)
	return out
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// A plain []time.Time per service. Every call rebuilds the "still inside
// the window" slice via a fresh scan (O(n)), then checks length, then
// appends. "Read, decide, write" are three unsynchronized steps -> two
// concurrent requests can both read length 2, both decide "allowed", and
// both append, letting 4 rollbacks through a limit of 3.
// ---------------------------------------------------------------------------
type BruteForceRateLimiter struct {
	mu         sync.Mutex // NOTE: guards the map itself, but NOT the decision below
	timestamps map[string][]time.Time
	audit      *AppendOnlyAuditLog
}

func NewBruteForceRateLimiter(audit *AppendOnlyAuditLog) *BruteForceRateLimiter {
	return &BruteForceRateLimiter{timestamps: make(map[string][]time.Time), audit: audit}
}

func (r *BruteForceRateLimiter) TryRollback(serviceName string, now time.Time) RollbackResult {
	// THE BUG: we grab the current slice, then release nothing — but the
	// *decision* (length check) and the *write* (append) are not atomic
	// with respect to OTHER goroutines also reading the same slice
	// concurrently, because the slice itself is read and rewritten
	// outside a single held lock below.
	r.mu.Lock()
	history := r.timestamps[serviceName]
	r.mu.Unlock()

	windowStart := now.Add(-WindowDuration)
	var recent []time.Time
	for _, t := range history { // O(n) rebuild, every call
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= MaxRollbacksPerWindow {
		oldest := recent[0]
		for _, t := range recent {
			if t.Before(oldest) {
				oldest = t
			}
		}
		retryAfter := oldest.Add(WindowDuration).Sub(now).Seconds()
		if retryAfter < 0 {
			retryAfter = 0
		}
		return RollbackResult{Accepted: false, RetryAfterSeconds: retryAfter}
	}

	time.Sleep(5 * time.Millisecond) // model real work between decision and write
	recent = append(recent, now)

	r.mu.Lock()
	r.timestamps[serviceName] = recent
	r.mu.Unlock()

	id := r.audit.Append(serviceName, now)
	return RollbackResult{Accepted: true, AuditID: id}
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Two fixes:
//  1. Efficiency: a container/list.List (doubly linked list) per service
//     used as a sliding-window log. New timestamps are pushed to the back;
//     expired ones are popped from the front. Since timestamps arrive in
//     non-decreasing order, expired entries are always contiguous at the
//     front -> amortized O(1) eviction (each entry is pushed once, popped
//     at most once, ever) instead of an O(n) rebuild per call.
//  2. Correctness: a per-service sync.Mutex makes "evict expired, check
//     length, append" one atomic critical section. A separate mutex per
//     service (not one global lock) means a flapping service's rollbacks
//     never add latency to an unrelated service's rollback requests.
// ---------------------------------------------------------------------------
type serviceWindow struct {
	mu     sync.Mutex
	events *list.List // of time.Time, oldest at Front()
}

type OptimalRateLimiter struct {
	windowsMu sync.Mutex
	windows   map[string]*serviceWindow
	audit     *AppendOnlyAuditLog
}

func NewOptimalRateLimiter(audit *AppendOnlyAuditLog) *OptimalRateLimiter {
	return &OptimalRateLimiter{windows: make(map[string]*serviceWindow), audit: audit}
}

func (r *OptimalRateLimiter) windowFor(serviceName string) *serviceWindow {
	r.windowsMu.Lock()
	defer r.windowsMu.Unlock()
	w, ok := r.windows[serviceName]
	if !ok {
		w = &serviceWindow{events: list.New()}
		r.windows[serviceName] = w
	}
	return w
}

func (r *OptimalRateLimiter) TryRollback(serviceName string, now time.Time) RollbackResult {
	w := r.windowFor(serviceName)
	w.mu.Lock()
	defer w.mu.Unlock()

	windowStart := now.Add(-WindowDuration)
	for e := w.events.Front(); e != nil; {
		t := e.Value.(time.Time)
		if t.After(windowStart) {
			break // list is time-ordered; nothing further front is expired
		}
		next := e.Next()
		w.events.Remove(e)
		e = next
	}

	if w.events.Len() >= MaxRollbacksPerWindow {
		oldest := w.events.Front().Value.(time.Time)
		retryAfter := oldest.Add(WindowDuration).Sub(now).Seconds()
		if retryAfter < 0 {
			retryAfter = 0
		}
		return RollbackResult{Accepted: false, RetryAfterSeconds: retryAfter}
	}

	w.events.PushBack(now)
	id := r.audit.Append(serviceName, now)
	return RollbackResult{Accepted: true, AuditID: id}
}
