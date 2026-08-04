package p4

import org.junit.Test
import org.junit.Assert.*

private fun simpleGraph(): Map<String, Target> = mapOf(
    "leaf" to Target("leaf", listOf("h_leaf")),
    "mid" to Target("mid", listOf("h_mid"), listOf("leaf")),
    "top" to Target("top", listOf("h_top"), listOf("mid"))
)

class P4Test {
    @Test
    fun `single node no deps`() {
        val g = mapOf("solo" to Target("solo", listOf("h1")))
        val bfKey = BruteForceCacheKeyer(g).computeKey("solo")
        val optKey = OptimalCacheKeyer(g).computeKey("solo")
        assertEquals(64, bfKey.length) // sha256 hex digest length
        assertEquals(64, optKey.length)
    }

    @Test
    fun `both implementations agree on a chain`() {
        val g = simpleGraph()
        val bf = BruteForceCacheKeyer(g).computeKey("top")
        val opt = OptimalCacheKeyer(g).computeKey("top")
        assertEquals(bf, opt)
    }

    @Test
    fun `diamond dependency is counted once not twice`() {
        val g = mapOf(
            "shared" to Target("shared", listOf("h_shared")),
            "d1" to Target("d1", listOf("h_d1"), listOf("shared")),
            "d2" to Target("d2", listOf("h_d2"), listOf("shared")),
            "top" to Target("top", listOf("h_top"), listOf("d1", "d2"))
        )
        val keyer = OptimalCacheKeyer(g)
        keyer.computeKey("top")
        assertEquals(4, keyer.calls) // shared, d1, d2, top — each visited once
    }

    @Test
    fun `dependency order does not change the key`() {
        val g1 = mapOf(
            "a" to Target("a", listOf("ha")),
            "b" to Target("b", listOf("hb")),
            "top" to Target("top", listOf("ht"), listOf("a", "b"))
        )
        val g2 = mapOf(
            "a" to Target("a", listOf("ha")),
            "b" to Target("b", listOf("hb")),
            "top" to Target("top", listOf("ht"), listOf("b", "a")) // swapped
        )
        val k1 = OptimalCacheKeyer(g1).computeKey("top")
        val k2 = OptimalCacheKeyer(g2).computeKey("top")
        assertEquals("dep declaration order must not affect the cache key", k1, k2)
    }

    @Test
    fun `changing a leaf source changes every ancestor key`() {
        val g = simpleGraph().toMutableMap()
        val before = OptimalCacheKeyer(g).computeKey("top")
        g["leaf"] = Target("leaf", listOf("h_leaf_CHANGED"))
        val after = OptimalCacheKeyer(g).computeKey("top")
        assertNotEquals(before, after)
    }

    @Test
    fun `unchanged graph produces identical key across runs`() {
        val g = simpleGraph()
        val k1 = OptimalCacheKeyer(g).computeKey("top")
        val k2 = OptimalCacheKeyer(g).computeKey("top")
        assertEquals(k1, k2)
    }

    @Test
    fun `changing an unrelated sibling does not change this targets key`() {
        val g = mapOf(
            "leaf" to Target("leaf", listOf("h_leaf")),
            "mid" to Target("mid", listOf("h_mid"), listOf("leaf")),
            "unrelated" to Target("unrelated", listOf("h_unrelated")),
            "top" to Target("top", listOf("h_top"), listOf("mid"))
        ).toMutableMap()
        val before = OptimalCacheKeyer(g).computeKey("top")
        g["unrelated"] = Target("unrelated", listOf("CHANGED"))
        val after = OptimalCacheKeyer(g).computeKey("top")
        assertEquals(before, after)
    }

    @Test(expected = CyclicDependencyException::class)
    fun `cyclic dependency is detected by optimal`() {
        val g = mapOf(
            "a" to Target("a", listOf("ha"), listOf("b")),
            "b" to Target("b", listOf("hb"), listOf("a"))
        )
        OptimalCacheKeyer(g).computeKey("a")
    }

    @Test(expected = CyclicDependencyException::class)
    fun `cyclic dependency is detected by brute force`() {
        val g = mapOf(
            "a" to Target("a", listOf("ha"), listOf("b")),
            "b" to Target("b", listOf("hb"), listOf("a"))
        )
        BruteForceCacheKeyer(g).computeKey("a")
    }

    @Test
    fun `compute all returns a key for every target`() {
        val g = simpleGraph()
        val keys = OptimalCacheKeyer(g).computeAll()
        assertEquals(setOf("leaf", "mid", "top"), keys.keys)
    }

    @Test
    fun `memoized version does asymptotically less work`() {
        val depth = 16
        val g = diamondChainGraph(depth)
        val topName = "l${depth - 1}a"

        val bf = BruteForceCacheKeyer(g)
        bf.computeKey(topName)
        val opt = OptimalCacheKeyer(g)
        opt.computeKey(topName)

        println("depth=$depth brute_force_calls=${bf.calls} optimal_calls=${opt.calls}")
        assertTrue("expected optimal call count to stay small, got ${opt.calls}", opt.calls < 100)
        assertTrue("expected brute-force call count to blow up, got ${bf.calls}", bf.calls > 10_000)
    }

    @Test
    fun `both versions agree despite perf gap`() {
        val depth = 8
        val g = diamondChainGraph(depth)
        val topName = "l${depth - 1}a"
        val bfKey = BruteForceCacheKeyer(g).computeKey(topName)
        val optKey = OptimalCacheKeyer(g).computeKey(topName)
        assertEquals(bfKey, optKey)
    }
}
