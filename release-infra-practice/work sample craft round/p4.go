// Package p4 implements a Bazel-style build-cache key generator: each
// target's key must capture its own source hashes AND its entire
// transitive dependency closure, deterministically (a Merkle tree).
package p1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type Target struct {
	Name         string
	SourceHashes []string
	Deps         []string // names of directly-declared dependency targets
}

type CyclicDependencyError struct{ At string }

func (e *CyclicDependencyError) Error() string {
	return fmt.Sprintf("cycle detected at %s", e.At)
}

// combine deterministically hashes a list of hash strings. Sorting is not
// optional: dependency declaration order must never leak into the cache
// key, or two functionally identical BUILD files that merely list deps in
// a different order would produce different keys.
func combine(parts []string) string {
	sorted := make([]string, len(parts))
	copy(sorted, parts)
	sort.Strings(sorted)
	h := sha256.New()
	for _, p := range sorted {
		h.Write([]byte(p))
		h.Write([]byte{0}) // separator: prevents ("ab","c") colliding with ("a","bc")
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// Plain recursion, no memoization. Fine for a simple chain. Falls over on a
// DIAMOND: two targets sharing a dependency each independently recompute
// the shared target's key from scratch. Worst case exponential in the
// depth of nested diamonds.
// ---------------------------------------------------------------------------
type BruteForceCacheKeyer struct {
	targets map[string]*Target
	Calls   int // instrumentation, to make the blowup measurable in tests
}

func NewBruteForceCacheKeyer(targets map[string]*Target) *BruteForceCacheKeyer {
	return &BruteForceCacheKeyer{targets: targets}
}

func (k *BruteForceCacheKeyer) ComputeKey(name string) (string, error) {
	return k.computeKey(name, map[string]bool{})
}

func (k *BruteForceCacheKeyer) computeKey(name string, visiting map[string]bool) (string, error) {
	k.Calls++
	if visiting[name] {
		return "", &CyclicDependencyError{At: name}
	}
	target, ok := k.targets[name]
	if !ok {
		return "", fmt.Errorf("unknown target: %s", name)
	}

	visitingNext := make(map[string]bool, len(visiting)+1)
	for kk := range visiting {
		visitingNext[kk] = true
	}
	visitingNext[name] = true

	depKeys := make([]string, 0, len(target.Deps))
	for _, dep := range target.Deps {
		dk, err := k.computeKey(dep, visitingNext)
		if err != nil {
			return "", err
		}
		depKeys = append(depKeys, dk)
	}
	return combine(append(append([]string{}, target.SourceHashes...), depKeys...)), nil
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Memoize: compute each target's key AT MOST ONCE. This is topological-
// order DP over a DAG -> O(V+E) instead of exponential. Cycle detection
// still uses "currently on the recursion stack" (not "ever visited") so a
// real cycle is still caught correctly, not silently memoized wrong.
// ---------------------------------------------------------------------------
type OptimalCacheKeyer struct {
	targets map[string]*Target
	memo    map[string]string
	Calls   int
}

func NewOptimalCacheKeyer(targets map[string]*Target) *OptimalCacheKeyer {
	return &OptimalCacheKeyer{targets: targets, memo: make(map[string]string)}
}

func (k *OptimalCacheKeyer) ComputeKey(name string) (string, error) {
	return k.computeKey(name, map[string]bool{})
}

func (k *OptimalCacheKeyer) computeKey(name string, visiting map[string]bool) (string, error) {
	if key, ok := k.memo[name]; ok {
		return key, nil // O(1) reuse, no recursion at all
	}
	k.Calls++
	if visiting[name] {
		return "", &CyclicDependencyError{At: name}
	}
	target, ok := k.targets[name]
	if !ok {
		return "", fmt.Errorf("unknown target: %s", name)
	}

	visitingNext := make(map[string]bool, len(visiting)+1)
	for kk := range visiting {
		visitingNext[kk] = true
	}
	visitingNext[name] = true

	depKeys := make([]string, 0, len(target.Deps))
	for _, dep := range target.Deps {
		dk, err := k.computeKey(dep, visitingNext)
		if err != nil {
			return "", err
		}
		depKeys = append(depKeys, dk)
	}
	key := combine(append(append([]string{}, target.SourceHashes...), depKeys...))
	k.memo[name] = key
	return key, nil
}

func (k *OptimalCacheKeyer) ComputeAll() (map[string]string, error) {
	out := make(map[string]string, len(k.targets))
	for name := range k.targets {
		key, err := k.ComputeKey(name)
		if err != nil {
			return nil, err
		}
		out[name] = key
	}
	return out, nil
}

// DiamondChainGraph builds `depth` stacked diamonds sharing a single base
// target, for making the brute-force exponential blowup observable:
//
//	base <- l0a, l0b <- l1a, l1b <- l2a, l2b <- ...
func DiamondChainGraph(depth int) map[string]*Target {
	targets := map[string]*Target{
		"base": {Name: "base", SourceHashes: []string{"src:base"}},
	}
	prevLayer := []string{"base"}
	for level := 0; level < depth; level++ {
		layer := []string{fmt.Sprintf("l%da", level), fmt.Sprintf("l%db", level)}
		for _, name := range layer {
			deps := make([]string, len(prevLayer))
			copy(deps, prevLayer)
			targets[name] = &Target{Name: name, SourceHashes: []string{"src:" + name}, Deps: deps}
		}
		prevLayer = layer
	}
	return targets
}
