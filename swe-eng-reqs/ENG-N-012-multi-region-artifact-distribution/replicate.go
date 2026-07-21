package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ErrUnknownRegion = errors.New("unknown or unsafe region")

type ReplicationStatus struct {
	Regions []string `json:"regions"`
	Pending int64    `json:"pending"`
	LagMs   int64    `json:"lag_ms"`
}

type Replicator struct {
	mu      sync.RWMutex
	regions map[string]map[string][]byte
	delay   time.Duration
	pending atomic.Int64
	lagMs   atomic.Int64
	wg      sync.WaitGroup
}

func NewReplicator(regions []string, delay time.Duration) *Replicator {
	unique := make(map[string]map[string][]byte)
	for _, region := range regions {
		if safeRegion(region) {
			unique[region] = make(map[string][]byte)
		}
	}
	if len(unique) == 0 {
		unique["us-east"] = make(map[string][]byte)
	}
	if len(unique) == 1 {
		fallback := "eu-west"
		if _, exists := unique[fallback]; exists {
			fallback = "us-east"
		}
		unique[fallback] = make(map[string][]byte)
	}
	if delay <= 0 {
		delay = time.Millisecond
	}
	return &Replicator{regions: unique, delay: delay}
}

func (r *Replicator) Put(region string, blob []byte) (string, error) {
	if !safeRegion(region) {
		return "", ErrUnknownRegion
	}
	digest := regionalDigest(blob)
	r.mu.Lock()
	bucket, ok := r.regions[region]
	if !ok {
		r.mu.Unlock()
		return "", ErrUnknownRegion
	}
	bucket[digest] = append([]byte(nil), blob...)
	targets := make([]string, 0, len(r.regions)-1)
	for target := range r.regions {
		if target != region {
			targets = append(targets, target)
		}
	}
	r.mu.Unlock()

	r.pending.Add(1)
	lag := r.delay.Milliseconds()
	if lag < 1 {
		lag = 1
	}
	r.lagMs.Store(lag)
	r.wg.Add(1)
	go r.replicate(targets, digest, blob)
	return digest, nil
}

func (r *Replicator) replicate(targets []string, digest string, blob []byte) {
	defer r.wg.Done()
	time.Sleep(r.delay)
	r.mu.Lock()
	for _, target := range targets {
		r.regions[target][digest] = append([]byte(nil), blob...)
	}
	r.mu.Unlock()
	if r.pending.Add(-1) == 0 {
		r.lagMs.Store(0)
	}
}

func (r *Replicator) Get(region, digest string) ([]byte, bool) {
	if !safeRegion(region) || !validRegionalDigest(digest) {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	bucket, ok := r.regions[region]
	if !ok {
		return nil, false
	}
	blob, ok := bucket[digest]
	return append([]byte(nil), blob...), ok
}

func (r *Replicator) Status() ReplicationStatus {
	r.mu.RLock()
	regions := make([]string, 0, len(r.regions))
	for region := range r.regions {
		regions = append(regions, region)
	}
	r.mu.RUnlock()
	sort.Strings(regions)
	return ReplicationStatus{Regions: regions, Pending: r.pending.Load(), LagMs: r.lagMs.Load()}
}

func (r *Replicator) LagMs() int64 {
	return r.lagMs.Load()
}

func (r *Replicator) Wait() {
	r.wg.Wait()
}

func regionalDigest(blob []byte) string {
	sum := sha256.Sum256(blob)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validRegionalDigest(digest string) bool {
	if len(digest) != 71 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(digest, "sha256:"))
	return err == nil
}

func safeRegion(region string) bool {
	return region != "" && !strings.Contains(region, "..") && !strings.ContainsAny(region, `/\`)
}
