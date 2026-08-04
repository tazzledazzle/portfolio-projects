package p1

import (
	"errors"
	"testing"
)

func simpleGraph() map[string]*Target {
	return map[string]*Target{
		"leaf": {Name: "leaf", SourceHashes: []string{"h_leaf"}},
		"mid":  {Name: "mid", SourceHashes: []string{"h_mid"}, Deps: []string{"leaf"}},
		"top":  {Name: "top", SourceHashes: []string{"h_top"}, Deps: []string{"mid"}},
	}
}

func TestSingleNodeNoDeps(t *testing.T) {
	g := map[string]*Target{"solo": {Name: "solo", SourceHashes: []string{"h1"}}}
	for _, keyer := range []interface {
		ComputeKey(string) (string, error)
	}{NewBruteForceCacheKeyer(g), NewOptimalCacheKeyer(g)} {
		key, err := keyer.ComputeKey("solo")
		if err != nil {
			t.Fatal(err)
		}
		if len(key) != 64 { // sha256 hex digest length
			t.Fatalf("expected 64-char hex digest, got %d chars", len(key))
		}
	}
}

func TestBothImplementationsAgreeOnAChain(t *testing.T) {
	g := simpleGraph()
	bf, err := NewBruteForceCacheKeyer(g).ComputeKey("top")
	if err != nil {
		t.Fatal(err)
	}
	opt, err := NewOptimalCacheKeyer(g).ComputeKey("top")
	if err != nil {
		t.Fatal(err)
	}
	if bf != opt {
		t.Fatalf("brute force and optimal disagree: %s vs %s", bf, opt)
	}
}

func TestDiamondDependencyCountedOnce(t *testing.T) {
	g := map[string]*Target{
		"shared": {Name: "shared", SourceHashes: []string{"h_shared"}},
		"d1":     {Name: "d1", SourceHashes: []string{"h_d1"}, Deps: []string{"shared"}},
		"d2":     {Name: "d2", SourceHashes: []string{"h_d2"}, Deps: []string{"shared"}},
		"top":    {Name: "top", SourceHashes: []string{"h_top"}, Deps: []string{"d1", "d2"}},
	}
	keyer := NewOptimalCacheKeyer(g)
	if _, err := keyer.ComputeKey("top"); err != nil {
		t.Fatal(err)
	}
	if keyer.Calls != 4 { // shared, d1, d2, top — each visited exactly once
		t.Fatalf("expected 4 calls (each target visited once), got %d", keyer.Calls)
	}
}

func TestDependencyOrderDoesNotAffectKey(t *testing.T) {
	g1 := map[string]*Target{
		"a":   {Name: "a", SourceHashes: []string{"ha"}},
		"b":   {Name: "b", SourceHashes: []string{"hb"}},
		"top": {Name: "top", SourceHashes: []string{"ht"}, Deps: []string{"a", "b"}},
	}
	g2 := map[string]*Target{
		"a":   {Name: "a", SourceHashes: []string{"ha"}},
		"b":   {Name: "b", SourceHashes: []string{"hb"}},
		"top": {Name: "top", SourceHashes: []string{"ht"}, Deps: []string{"b", "a"}}, // swapped
	}
	k1, err := NewOptimalCacheKeyer(g1).ComputeKey("top")
	if err != nil {
		t.Fatal(err)
	}
	k2, err := NewOptimalCacheKeyer(g2).ComputeKey("top")
	if err != nil {
		t.Fatal(err)
	}
	if k1 != k2 {
		t.Fatal("dep declaration order must not affect the cache key")
	}
}

func TestChangingLeafSourceChangesAncestorKey(t *testing.T) {
	g := simpleGraph()
	before, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	g["leaf"].SourceHashes = []string{"h_leaf_CHANGED"}
	after, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	if before == after {
		t.Fatal("expected key to change when a leaf source changes")
	}
}

func TestUnchangedGraphProducesIdenticalKey(t *testing.T) {
	g := simpleGraph()
	k1, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	k2, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	if k1 != k2 {
		t.Fatal("expected identical key across repeated runs on an unchanged graph")
	}
}

func TestUnrelatedSiblingChangeDoesNotAffectKey(t *testing.T) {
	g := map[string]*Target{
		"leaf":      {Name: "leaf", SourceHashes: []string{"h_leaf"}},
		"mid":       {Name: "mid", SourceHashes: []string{"h_mid"}, Deps: []string{"leaf"}},
		"unrelated": {Name: "unrelated", SourceHashes: []string{"h_unrelated"}},
		"top":       {Name: "top", SourceHashes: []string{"h_top"}, Deps: []string{"mid"}},
	}
	before, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	g["unrelated"].SourceHashes = []string{"CHANGED"}
	after, _ := NewOptimalCacheKeyer(g).ComputeKey("top")
	if before != after {
		t.Fatal("an unrelated sibling's change must not affect this target's key")
	}
}

func TestCyclicDependencyDetected(t *testing.T) {
	g := map[string]*Target{
		"a": {Name: "a", SourceHashes: []string{"ha"}, Deps: []string{"b"}},
		"b": {Name: "b", SourceHashes: []string{"hb"}, Deps: []string{"a"}},
	}
	var cycleErr *CyclicDependencyError

	_, err := NewOptimalCacheKeyer(g).ComputeKey("a")
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected CyclicDependencyError from optimal, got %v", err)
	}

	_, err = NewBruteForceCacheKeyer(g).ComputeKey("a")
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected CyclicDependencyError from brute force, got %v", err)
	}
}

func TestComputeAllReturnsKeyForEveryTarget(t *testing.T) {
	g := simpleGraph()
	keys, err := NewOptimalCacheKeyer(g).ComputeAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"leaf", "mid", "top"} {
		if _, ok := keys[name]; !ok {
			t.Fatalf("missing key for target %s", name)
		}
	}
}

func TestMemoizedVersionDoesAsymptoticallyLessWork(t *testing.T) {
	depth := 16
	g := DiamondChainGraph(depth)
	topName := "l15a"

	bf := NewBruteForceCacheKeyer(g)
	if _, err := bf.ComputeKey(topName); err != nil {
		t.Fatal(err)
	}
	opt := NewOptimalCacheKeyer(g)
	if _, err := opt.ComputeKey(topName); err != nil {
		t.Fatal(err)
	}

	t.Logf("depth=%d brute_force_calls=%d optimal_calls=%d", depth, bf.Calls, opt.Calls)
	if opt.Calls >= 100 {
		t.Fatalf("expected optimal call count to stay small (linear), got %d", opt.Calls)
	}
	if bf.Calls <= 10000 {
		t.Fatalf("expected brute-force call count to blow up on a diamond chain, got %d", bf.Calls)
	}
}

func TestBothVersionsAgreeDespitePerfGap(t *testing.T) {
	depth := 8
	g := DiamondChainGraph(depth)
	topName := "l7a"
	bfKey, err := NewBruteForceCacheKeyer(g).ComputeKey(topName)
	if err != nil {
		t.Fatal(err)
	}
	optKey, err := NewOptimalCacheKeyer(g).ComputeKey(topName)
	if err != nil {
		t.Fatal(err)
	}
	if bfKey != optKey {
		t.Fatal("brute force and optimal must produce the identical final key")
	}
}
