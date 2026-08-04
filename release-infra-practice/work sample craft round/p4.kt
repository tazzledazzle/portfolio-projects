package p4

import java.security.MessageDigest

/**
 * Build-Cache Key Generator (Bazel-style).
 *
 * Each target's cache key must capture its own source hashes AND its
 * entire transitive dependency closure, deterministically (a Merkle
 * tree — exactly what Bazel's action graph does).
 */

data class Target(val name: String, val sourceHashes: List<String>, val deps: List<String> = emptyList())

class CyclicDependencyException(at: String) : Exception("cycle detected at $at")

/**
 * Deterministically combines a list of hash strings into one hash.
 * SORTING is not optional: dependency declaration order must never leak
 * into the cache key, or two functionally identical BUILD files that
 * merely list deps in a different order would produce different keys.
 */
private fun combine(parts: List<String>): String {
    val digest = MessageDigest.getInstance("SHA-256")
    for (part in parts.sorted()) {
        digest.update(part.toByteArray())
        digest.update(0) // separator: prevents ("ab","c") colliding with ("a","bc")
    }
    return digest.digest().joinToString("") { "%02x".format(it) }
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// Plain recursion, no memoization. Fine for a simple chain. Falls over on a
// DIAMOND: two targets sharing a dependency each independently recompute
// the shared target's key from scratch. Worst case exponential in the
// depth of nested diamonds.
// ---------------------------------------------------------------------------
class BruteForceCacheKeyer(private val targets: Map<String, Target>) {
    var calls: Int = 0
        private set

    fun computeKey(name: String, visiting: Set<String> = emptySet()): String {
        calls++
        if (name in visiting) throw CyclicDependencyException(name)
        val target = targets[name] ?: error("unknown target: $name")

        val depKeys = target.deps.map { computeKey(it, visiting + name) }
        return combine(target.sourceHashes + depKeys)
    }
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Memoize: compute each target's key AT MOST ONCE. This is topological-
// order DP over a DAG -> O(V+E) instead of exponential. Cycle detection
// still uses "currently on the recursion stack" (not "ever visited") so a
// real cycle is still caught correctly, not silently memoized wrong.
// ---------------------------------------------------------------------------
class OptimalCacheKeyer(private val targets: Map<String, Target>) {
    private val memo = HashMap<String, String>()
    var calls: Int = 0
        private set

    fun computeKey(name: String, visiting: Set<String> = emptySet()): String {
        memo[name]?.let { return it } // O(1) reuse, no recursion at all

        calls++
        if (name in visiting) throw CyclicDependencyException(name)
        val target = targets[name] ?: error("unknown target: $name")

        val depKeys = target.deps.map { computeKey(it, visiting + name) }
        val key = combine(target.sourceHashes + depKeys)
        memo[name] = key
        return key
    }

    fun computeAll(): Map<String, String> = targets.keys.associateWith { computeKey(it) }
}

/**
 * Builds `depth` stacked diamonds sharing a single base target, to make
 * the brute-force exponential blowup observable:
 *
 *   base <- l0a, l0b <- l1a, l1b <- l2a, l2b <- ...
 */
fun diamondChainGraph(depth: Int): Map<String, Target> {
    val targets = HashMap<String, Target>()
    targets["base"] = Target("base", listOf("src:base"))
    var prevLayer = listOf("base")
    for (level in 0 until depth) {
        val layer = listOf("l${level}a", "l${level}b")
        for (name in layer) {
            targets[name] = Target(name, listOf("src:$name"), prevLayer)
        }
        prevLayer = layer
    }
    return targets
}
