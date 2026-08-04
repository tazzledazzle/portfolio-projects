"""
Problem 4: Build-Cache Key Generator (Bazel-style)
==========================================
Given a build graph — targets, each with its own source file hashes and a
list of declared dependency targets — compute a deterministic cache key
per target that captures its ENTIRE transitive input: its own sources
plus every dependency's cache key, recursively (a Merkle tree / content
hash, exactly what Bazel's action graph does).

Run:
    python3 -m unittest p4_build_cache_key -v
"""
from __future__ import annotations

import hashlib
from dataclasses import dataclass, field


@dataclass
class Target:
    name: str
    source_hashes: list[str]      # hashes of this target's own source files
    deps: list[str]                # names of directly-declared dependency targets


class CyclicDependencyError(Exception):
    pass


def _combine(parts: list[str]) -> str:
    """Deterministically combine a list of hash strings into one hash.
    SORTING is not optional: dict/list iteration order for a target's
    `deps` must never leak into the cache key, or two functionally
    identical BUILD files that merely list deps in a different order
    would produce different cache keys — a correctness-neutral but
    cache-effectiveness-destroying bug, and the mirror-image bug (two
    targets that HAPPEN to sort to the same digest by coincidence) is
    what cryptographic hashing (not just any hash) protects against.
    """
    h = hashlib.sha256()
    for part in sorted(parts):
        h.update(part.encode())
        h.update(b"\x00")  # separator, prevents ("ab","c") colliding with ("a","bc")
    return h.hexdigest()


# ---------------------------------------------------------------------------
# STEP 1 — BRUTE FORCE
# ---------------------------------------------------------------------------
# Reasoning:
#   The direct translation of "a target's key depends on its deps' keys":
#   a plain recursive function with NO memoization. For a simple chain
#   (A -> B -> C) this is fine — each target is visited once. It falls
#   over the instant the graph has a DIAMOND: e.g. targets D and E both
#   depend on shared target F. Computing key(D) recomputes key(F) from
#   scratch, and separately computing key(E) recomputes key(F) again.
#   In a monorepo-scale build graph, a small number of "leaf" utility
#   targets are depended on by huge fan-out — the classic case being a
#   shared logging/proto/base library depended on by hundreds of
#   services. Without memoization, computing cache keys for the whole
#   graph is worst-case EXPONENTIAL in the depth of the diamond chain
#   (each additional layer of diamonds multiplies redundant recomputation
#   of the shared base), even though the actual DAG has only V nodes and
#   E edges. This shows up as "why does computing cache keys take longer
#   than the actual build" on a real diamond-heavy dependency graph.
# ---------------------------------------------------------------------------
class BruteForceCacheKeyer:
    def __init__(self, targets: dict[str, Target]):
        self._targets = targets
        self.calls = 0  # instrumentation, so the test can prove the blowup

    def compute_key(self, target_name: str, _visiting: frozenset[str] = frozenset()) -> str:
        self.calls += 1
        if target_name in _visiting:
            raise CyclicDependencyError(f"cycle detected at {target_name}")
        target = self._targets[target_name]

        dep_keys = [
            self.compute_key(dep, _visiting | {target_name})
            for dep in target.deps
        ]
        return _combine(target.source_hashes + dep_keys)


# ---------------------------------------------------------------------------
# STEP 2 — OPTIMAL
# ---------------------------------------------------------------------------
# Reasoning:
#   Memoize: compute each target's key AT MOST ONCE, cache it, and reuse
#   the cached value for every subsequent reference to that target as a
#   dependency. This is exactly topological-order DP over a DAG — every
#   target's key is computed only after all of its dependencies' keys are
#   already known, and each target contributes O(1) work beyond its own
#   deps list, giving O(V + E) total instead of exponential blowup.
#   Cycle detection uses the same "currently on the recursion stack" set
#   as the brute-force version (correctness must not regress when we
#   optimize for speed) — a cycle in a build graph is a real user error
#   (target A depends on target B which depends back on A) and must be
#   reported clearly, not silently infinite-loop or silently produce a
#   wrong/stale key.
# ---------------------------------------------------------------------------
class OptimalCacheKeyer:
    def __init__(self, targets: dict[str, Target]):
        self._targets = targets
        self._memo: dict[str, str] = {}
        self.calls = 0  # instrumentation

    def compute_key(self, target_name: str, _visiting: frozenset[str] = frozenset()) -> str:
        if target_name in self._memo:
            return self._memo[target_name]  # O(1) reuse, no recursion at all

        self.calls += 1
        if target_name in _visiting:
            raise CyclicDependencyError(f"cycle detected at {target_name}")
        target = self._targets[target_name]

        dep_keys = [
            self.compute_key(dep, _visiting | {target_name})
            for dep in target.deps
        ]
        key = _combine(target.source_hashes + dep_keys)
        self._memo[target_name] = key
        return key

    def compute_all(self) -> dict[str, str]:
        return {name: self.compute_key(name) for name in self._targets}


# ---------------------------------------------------------------------------
# Correctness-of-invalidation note (asked for explicitly in the prompt):
#   The single most common real-world bug in a build-cache key scheme is
#   NOT in this recursive structure — it's forgetting to include something
#   that actually affects the build output. Concretely, all of the
#   following must be part of the key or you get FALSE CACHE HITS (the
#   dangerous failure mode — silently serving a stale/wrong artifact,
#   not just a slow rebuild):
#     - the full transitive dependency closure (this file's whole point)
#     - the toolchain/compiler version used to build the target
#     - relevant build flags / configuration (e.g. debug vs release)
#     - anything the build reads that isn't a declared source file
#       (an undeclared read of an env var or a file outside the sandbox
#       is exactly the class of bug Bazel's hermeticity sandboxing exists
#       to catch)
#   Conversely, including something in the key that does NOT actually
#   affect output (e.g. a comment-only source change if you're hashing
#   post-parse ASTs) only costs you cache EFFICIENCY, not correctness —
#   a much cheaper failure mode to have, which is why "when in doubt,
#   include it in the key" is the right default.
# ---------------------------------------------------------------------------


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
import unittest  # noqa: E402


def _diamond_chain_graph(depth: int) -> dict[str, Target]:
    """Builds a graph shaped like `depth` stacked diamonds sharing a
    single base target, to make the brute-force exponential blowup
    observable in a small, fast test:

        base
        ^  ^
        |  |
       l1a l1b       <- both depend on base
        ^  ^
        |  |
       l2a l2b       <- both depend on {l1a, l1b}
        ...
    """
    targets: dict[str, Target] = {"base": Target("base", ["src:base"], [])}
    prev_layer = ["base"]
    for level in range(depth):
        layer = [f"l{level}a", f"l{level}b"]
        for name in layer:
            targets[name] = Target(name, [f"src:{name}"], list(prev_layer))
        prev_layer = layer
    return targets


class TestCacheKeyerContract(unittest.TestCase):
    def _make_simple_graph(self) -> dict[str, Target]:
        return {
            "leaf": Target("leaf", ["h_leaf"], []),
            "mid": Target("mid", ["h_mid"], ["leaf"]),
            "top": Target("top", ["h_top"], ["mid"]),
        }

    def test_single_node_no_deps(self):
        for impl_cls in (BruteForceCacheKeyer, OptimalCacheKeyer):
            impl = impl_cls({"solo": Target("solo", ["h1"], [])})
            key = impl.compute_key("solo")
            self.assertEqual(len(key), 64)  # sha256 hex digest length

    def test_both_implementations_agree_on_a_chain(self):
        graph = self._make_simple_graph()
        bf = BruteForceCacheKeyer(graph).compute_key("top")
        opt = OptimalCacheKeyer(graph).compute_key("top")
        self.assertEqual(bf, opt)

    def test_diamond_dependency_is_counted_once_not_twice(self):
        graph = {
            "shared": Target("shared", ["h_shared"], []),
            "d1": Target("d1", ["h_d1"], ["shared"]),
            "d2": Target("d2", ["h_d2"], ["shared"]),
            "top": Target("top", ["h_top"], ["d1", "d2"]),
        }
        keyer = OptimalCacheKeyer(graph)
        keyer.compute_key("top")
        # "shared" should have been visited exactly once, not once per
        # path that reaches it.
        self.assertEqual(keyer.calls, 4)  # shared, d1, d2, top

    def test_dependency_order_in_the_source_list_does_not_change_the_key(self):
        g1 = {
            "a": Target("a", ["ha"], []),
            "b": Target("b", ["hb"], []),
            "top": Target("top", ["ht"], ["a", "b"]),
        }
        g2 = {
            "a": Target("a", ["ha"], []),
            "b": Target("b", ["hb"], []),
            "top": Target("top", ["ht"], ["b", "a"]),  # order swapped
        }
        k1 = OptimalCacheKeyer(g1).compute_key("top")
        k2 = OptimalCacheKeyer(g2).compute_key("top")
        self.assertEqual(k1, k2, "dep declaration order must not affect the cache key")

    def test_changing_a_leaf_source_changes_every_ancestor_key(self):
        graph = self._make_simple_graph()
        before = OptimalCacheKeyer(graph).compute_key("top")

        graph["leaf"].source_hashes = ["h_leaf_CHANGED"]
        after = OptimalCacheKeyer(graph).compute_key("top")

        self.assertNotEqual(before, after)

    def test_unchanged_graph_produces_identical_key_across_runs(self):
        graph = self._make_simple_graph()
        k1 = OptimalCacheKeyer(graph).compute_key("top")
        k2 = OptimalCacheKeyer(graph).compute_key("top")
        self.assertEqual(k1, k2)

    def test_changing_an_unrelated_sibling_does_not_change_this_targets_key(self):
        graph = {
            "leaf": Target("leaf", ["h_leaf"], []),
            "mid": Target("mid", ["h_mid"], ["leaf"]),
            "unrelated": Target("unrelated", ["h_unrelated"], []),
            "top": Target("top", ["h_top"], ["mid"]),  # note: does NOT depend on 'unrelated'
        }
        before = OptimalCacheKeyer(graph).compute_key("top")
        graph["unrelated"].source_hashes = ["CHANGED"]
        after = OptimalCacheKeyer(graph).compute_key("top")
        self.assertEqual(before, after)

    def test_cyclic_dependency_is_detected_not_infinite_looped(self):
        graph = {
            "a": Target("a", ["ha"], ["b"]),
            "b": Target("b", ["hb"], ["a"]),  # cycle: a -> b -> a
        }
        with self.assertRaises(CyclicDependencyError):
            OptimalCacheKeyer(graph).compute_key("a")
        with self.assertRaises(CyclicDependencyError):
            BruteForceCacheKeyer(graph).compute_key("a")

    def test_compute_all_returns_a_key_for_every_target(self):
        graph = self._make_simple_graph()
        keys = OptimalCacheKeyer(graph).compute_all()
        self.assertEqual(set(keys.keys()), {"leaf", "mid", "top"})


class TestBruteForceExponentialBlowup(unittest.TestCase):
    def test_memoized_version_does_asymptotically_less_work_on_diamond_chain(self):
        # Depth chosen to be fast to run (<1s) but large enough that the
        # exponential-vs-linear gap is dramatic and unmistakable.
        depth = 16
        graph = _diamond_chain_graph(depth)

        bf = BruteForceCacheKeyer(graph)
        bf.compute_key(f"l{depth - 1}a")

        opt = OptimalCacheKeyer(graph)
        opt.compute_key(f"l{depth - 1}a")

        # Optimal: exactly one visit per node actually on the reachable
        # subgraph -> O(V). Brute force: roughly doubles per diamond
        # layer -> ~2^depth. At depth=16 that's a >1000x gap.
        self.assertLess(opt.calls, 100)
        self.assertGreater(bf.calls, 10_000,
                            "expected the unmemoized version's call count to blow up "
                            "on a diamond-shaped dependency graph")

    def test_both_versions_still_agree_on_the_final_key_despite_the_perf_gap(self):
        depth = 8
        graph = _diamond_chain_graph(depth)
        bf_key = BruteForceCacheKeyer(graph).compute_key(f"l{depth - 1}a")
        opt_key = OptimalCacheKeyer(graph).compute_key(f"l{depth - 1}a")
        self.assertEqual(bf_key, opt_key)


if __name__ == "__main__":
    unittest.main()
