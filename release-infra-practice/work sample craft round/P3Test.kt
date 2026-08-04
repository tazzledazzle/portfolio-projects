package p3

import org.junit.Test
import org.junit.Assert.*
import kotlin.random.Random

private fun bothEvaluators(): List<CanaryEvaluator> =
    listOf(BruteForceCanaryEvaluator(), OptimalCanaryEvaluator())

class P3ContractTest {
    @Test
    fun `insufficient samples is continue`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 1, 2)
            ev.recordBaseline(0.0, 0, 2)
            assertEquals(ev.javaClass.simpleName, Verdict.CONTINUE, ev.evaluate(0.0))
        }
    }

    @Test
    fun `zero traffic is continue not pass`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 0, 0)
            ev.recordBaseline(0.0, 0, 100)
            assertEquals(ev.javaClass.simpleName, Verdict.CONTINUE, ev.evaluate(0.0))
        }
    }

    @Test
    fun `equal rates pass`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 5, 100)
            ev.recordBaseline(0.0, 5, 100)
            assertEquals(ev.javaClass.simpleName, Verdict.PASS, ev.evaluate(0.0))
        }
    }

    @Test
    fun `exactly at threshold is pass`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 10, 100)   // 10%
            ev.recordBaseline(0.0, 5, 100)  // 5% -> delta exactly 0.05
            assertEquals(ev.javaClass.simpleName, Verdict.PASS, ev.evaluate(0.0))
        }
    }

    @Test
    fun `just over threshold is fail`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 11, 100)   // 11%
            ev.recordBaseline(0.0, 5, 100)  // 5% -> delta 0.06
            assertEquals(ev.javaClass.simpleName, Verdict.FAIL, ev.evaluate(0.0))
        }
    }

    @Test
    fun `canary better than baseline passes`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 1, 100)
            ev.recordBaseline(0.0, 10, 100)
            assertEquals(ev.javaClass.simpleName, Verdict.PASS, ev.evaluate(0.0))
        }
    }

    @Test
    fun `points outside window are ignored`() {
        for (ev in bothEvaluators()) {
            ev.recordCanary(0.0, 1000, 1000) // ancient spike
            var ts = 1000.0
            while (ts < 1060) {
                ev.recordCanary(ts, 0, 2)
                ev.recordBaseline(ts, 0, 2)
                ts += 1.0
            }
            assertEquals(ev.javaClass.simpleName, Verdict.PASS, ev.evaluate(1060.0))
        }
    }

    @Test
    fun `sparse data still accumulates to a verdict`() {
        for (ev in bothEvaluators()) {
            for (ts in listOf(0.0, 90.0, 180.0, 270.0, 300.0, 330.0)) {
                ev.recordCanary(ts, 1, 20)
                ev.recordBaseline(ts, 1, 20)
            }
            val v = ev.evaluate(330.0)
            assertTrue(ev.javaClass.simpleName, v == Verdict.PASS || v == Verdict.FAIL)
        }
    }
}

class P3DifferentialTest {
    @Test
    fun `optimal matches brute force over random stream`() {
        val rng = Random(42)
        val opt = OptimalCanaryEvaluator()
        val brute = BruteForceCanaryEvaluator()

        var now = 0.0
        repeat(500) {
            now += 0.1 + rng.nextDouble() * 2.9
            val errs = rng.nextInt(4)
            val reqs = 1 + rng.nextInt(20)
            opt.recordCanary(now, errs, reqs)
            brute.recordCanary(now, errs, reqs)

            val errs2 = rng.nextInt(4)
            val reqs2 = 1 + rng.nextInt(20)
            opt.recordBaseline(now, errs2, reqs2)
            brute.recordBaseline(now, errs2, reqs2)

            val want = brute.evaluate(now)
            val got = opt.evaluate(now)
            assertEquals("mismatch at t=$now", want, got)
        }
    }
}
