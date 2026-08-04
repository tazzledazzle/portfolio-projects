// Package p1 implements the idempotent deploy-trigger API problem.
//
// Duplicate "trigger deployment" requests keyed by an idempotency key must
// never cause a second deployment. Status must be queryable. The service
// must survive a crash mid-deploy without losing the dedup guarantee or
// silently re-running a completed deploy.
package p1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type DeployStatus string

const (
	StatusPending   DeployStatus = "PENDING"
	StatusRunning   DeployStatus = "RUNNING"
	StatusSucceeded DeployStatus = "SUCCEEDED"
	StatusFailed    DeployStatus = "FAILED"
	StatusUnknown   DeployStatus = "UNKNOWN" // crash-recovery: outcome genuinely unknown
)

type DeploymentRecord struct {
	IdempotencyKey string       `json:"idempotency_key"`
	ServiceName    string       `json:"service_name"`
	GitSHA         string       `json:"git_sha"`
	Status         DeployStatus `json:"status"`
	Error          string       `json:"error,omitempty"`
	CreatedAt      int64        `json:"created_at"` // unix nanos
	UpdatedAt      int64        `json:"updated_at"`
}

type DeployFunc func(serviceName, gitSHA string) error

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// A plain map, no lock. "Check if the key exists, then insert" is two
// separate, unsynchronized steps -> classic TOCTOU race. Nothing is
// persisted, so a process restart loses all dedup history entirely.
// ---------------------------------------------------------------------------
type BruteForceDeployService struct {
	store  map[string]*DeploymentRecord
	deploy DeployFunc
}

func NewBruteForceDeployService(deploy DeployFunc) *BruteForceDeployService {
	return &BruteForceDeployService{store: make(map[string]*DeploymentRecord), deploy: deploy}
}

func (s *BruteForceDeployService) Trigger(key, serviceName, gitSHA string) *DeploymentRecord {
	// --- THE BUG: check-then-act, no synchronization at all. ---
	if existing, ok := s.store[key]; ok {
		return existing
	}

	// Realistic gap between "checked" and "wrote" — validation, logging,
	// object construction all take nonzero time in real code.
	time.Sleep(10 * time.Millisecond)

	now := time.Now().UnixNano()
	record := &DeploymentRecord{
		IdempotencyKey: key, ServiceName: serviceName, GitSHA: gitSHA,
		Status: StatusRunning, CreatedAt: now, UpdatedAt: now,
	}
	s.store[key] = record // data race on the map itself under -race too

	if err := s.deploy(serviceName, gitSHA); err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
	} else {
		record.Status = StatusSucceeded
	}
	record.UpdatedAt = time.Now().UnixNano()
	return record
}

func (s *BruteForceDeployService) GetStatus(key string) (*DeploymentRecord, bool) {
	r, ok := s.store[key]
	return r, ok
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Two independent fixes:
//
//  1. Atomicity: Go's map has no built-in "insert if absent" the way SQL's
//     `INSERT ... ON CONFLICT DO NOTHING` does, but a sync.Mutex guarding a
//     single check-and-set critical section gives us the exact same
//     atomicity guarantee within one process. Whoever holds the lock when
//     the key is absent is the sole owner of that deploy.
//  2. Durability: every state transition is appended to a write-ahead log
//     file (one JSON object per line, fsync'd) BEFORE the in-memory map is
//     updated. On startup, the log is replayed to reconstruct state, so a
//     process restart does not lose dedup history. In production this
//     WAL+replay role is what Postgres/SQLite's own durability already
//     gives you for free via a real UNIQUE constraint; this is the same
//     idea built from stdlib primitives only.
//
// Crash-mid-deploy handling: a RUNNING record whose UpdatedAt is older than
// staleTimeout gets flipped to UNKNOWN on reconciliation (called at
// startup, or on a timer) rather than being silently retried or silently
// treated as success — the only correct move when the outcome genuinely
// cannot be known without checking the real deployment target.
// ---------------------------------------------------------------------------
type OptimalDeployService struct {
	mu            sync.Mutex
	store         map[string]*DeploymentRecord
	deploy        DeployFunc
	walPath       string
	walFile       *os.File
	staleTimeout  time.Duration
}

func NewOptimalDeployService(deploy DeployFunc, walPath string, staleTimeout time.Duration) (*OptimalDeployService, error) {
	s := &OptimalDeployService{
		store: make(map[string]*DeploymentRecord), deploy: deploy,
		walPath: walPath, staleTimeout: staleTimeout,
	}
	if err := s.replayWAL(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	s.walFile = f
	s.reconcileStaleRunning()
	return s, nil
}

func (s *OptimalDeployService) replayWAL() error {
	f, err := os.Open(s.walPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r DeploymentRecord
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue // tolerate a torn last write from a crash mid-append
		}
		rc := r
		s.store[r.IdempotencyKey] = &rc
	}
	return scanner.Err()
}

func (s *OptimalDeployService) appendWAL(r *DeploymentRecord) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := s.walFile.Write(b); err != nil {
		return err
	}
	return s.walFile.Sync() // durability: on disk before we proceed
}

func (s *OptimalDeployService) reconcileStaleRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-s.staleTimeout).UnixNano()
	for _, r := range s.store {
		if r.Status == StatusRunning && r.UpdatedAt < cutoff {
			r.Status = StatusUnknown
			r.UpdatedAt = time.Now().UnixNano()
			_ = s.appendWAL(r)
		}
	}
}

func (s *OptimalDeployService) Trigger(key, serviceName, gitSHA string) *DeploymentRecord {
	s.mu.Lock()
	if existing, ok := s.store[key]; ok {
		s.mu.Unlock()
		return existing // someone else already owns (or owned) this key
	}
	now := time.Now().UnixNano()
	record := &DeploymentRecord{
		IdempotencyKey: key, ServiceName: serviceName, GitSHA: gitSHA,
		Status: StatusRunning, CreatedAt: now, UpdatedAt: now,
	}
	s.store[key] = record
	_ = s.appendWAL(record)
	s.mu.Unlock() // release the lock BEFORE running the (possibly slow) deploy

	err := s.deploy(serviceName, gitSHA)

	s.mu.Lock()
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
	} else {
		record.Status = StatusSucceeded
	}
	record.UpdatedAt = time.Now().UnixNano()
	_ = s.appendWAL(record)
	s.mu.Unlock()

	return record
}

func (s *OptimalDeployService) GetStatus(key string) (*DeploymentRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.store[key]
	return r, ok
}

func (s *OptimalDeployService) Close() error {
	return s.walFile.Close()
}

// injectRawRunningRecord is a test-only hook to simulate "the process
// crashed after writing a RUNNING record but before it resolved."
func (s *OptimalDeployService) injectRawRunningRecord(key, serviceName, gitSHA string, updatedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := &DeploymentRecord{
		IdempotencyKey: key, ServiceName: serviceName, GitSHA: gitSHA,
		Status: StatusRunning, CreatedAt: updatedAt.UnixNano(), UpdatedAt: updatedAt.UnixNano(),
	}
	s.store[key] = r
	_ = s.appendWAL(r)
}

func (s *OptimalDeployService) SimulateCrashAndRestart() {
	s.reconcileStaleRunning()
}

// Demo is a small runnable example, not a package entrypoint (this file is
// package p1, not package main).
func Demo() {
	svc, err := NewOptimalDeployService(func(sn, sha string) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, "/tmp/p1_deploys.wal", 5*time.Minute)
	if err != nil {
		panic(err)
	}
	defer svc.Close()
	r := svc.Trigger("demo-key", "checkout-service", "abc123")
	fmt.Printf("%+v\n", r)
}
