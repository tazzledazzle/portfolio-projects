package p5

import org.junit.Test
import org.junit.Assert.*
import java.util.concurrent.CyclicBarrier
import kotlin.random.Random

private fun bothCalculators(): List<CFRCalculator> =
    listOf(BruteForceCFRCalculator(), OptimalCFRCalculator())

class P5ContractTest {
    @Test
    fun `no events returns null`() {
        for (c in bothCalculators()) {
            assertNull(c.javaClass.simpleName, c.changeFailureRate("svc", 1000.0))
        }
    }

    @Test
    fun `all success is zero`() {
        for (c in bothCalculators()) {
            for (i in 0 until 5) c.recordDeploy("svc", 1000.0 + i, Outcome.SUCCESS)
            assertEquals(c.javaClass.simpleName, 0.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `all failed is one`() {
        for (c in bothCalculators()) {
            for (i in 0 until 5) c.recordDeploy("svc", 1000.0 + i, Outcome.FAILED)
            assertEquals(c.javaClass.simpleName, 1.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `mixed outcomes rate`() {
        for (c in bothCalculators()) {
            c.recordDeploy("svc", 1000.0, Outcome.SUCCESS)
            c.recordDeploy("svc", 1001.0, Outcome.SUCCESS)
            c.recordDeploy("svc", 1002.0, Outcome.FAILED)
            c.recordDeploy("svc", 1003.0, Outcome.ROLLED_BACK)
            assertEquals(c.javaClass.simpleName, 0.5, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `pending excluded from calculation`() {
        for (c in bothCalculators()) {
            c.recordDeploy("svc", 1000.0, Outcome.SUCCESS)
            c.recordDeploy("svc", 1001.0, Outcome.PENDING)
            c.recordDeploy("svc", 1002.0, Outcome.PENDING)
            assertEquals(c.javaClass.simpleName, 0.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `events outside window excluded`() {
        for (c in bothCalculators()) {
            c.recordDeploy("svc", 0.0, Outcome.FAILED)
            c.recordDeploy("svc", WINDOW_SECONDS + 100, Outcome.SUCCESS)
            assertEquals(c.javaClass.simpleName, 0.0, c.changeFailureRate("svc", WINDOW_SECONDS + 100)!!, 0.0)
        }
    }

    @Test
    fun `exactly window seconds old still counts`() {
        for (c in bothCalculators()) {
            val now = 1_000_000.0
            c.recordDeploy("svc", now - WINDOW_SECONDS, Outcome.FAILED)
            assertEquals(c.javaClass.simpleName, 1.0, c.changeFailureRate("svc", now)!!, 0.0)
        }
    }

    @Test
    fun `one second older than window excluded`() {
        for (c in bothCalculators()) {
            val now = 1_000_000.0
            c.recordDeploy("svc", now - WINDOW_SECONDS - 1, Outcome.FAILED)
            c.recordDeploy("svc", now - WINDOW_SECONDS + 1, Outcome.SUCCESS)
            assertEquals(c.javaClass.simpleName, 0.0, c.changeFailureRate("svc", now)!!, 0.0)
        }
    }

    @Test
    fun `services are independent`() {
        for (c in bothCalculators()) {
            c.recordDeploy("svc-a", 1000.0, Outcome.FAILED)
            c.recordDeploy("svc-b", 1000.0, Outcome.SUCCESS)
            assertEquals(1.0, c.changeFailureRate("svc-a", 1010.0)!!, 0.0)
            assertEquals(0.0, c.changeFailureRate("svc-b", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `late arriving outcome updates the rate`() {
        for (c in bothCalculators()) {
            val eid = c.recordDeploy("svc", 1000.0, Outcome.PENDING)
            c.recordDeploy("svc", 1001.0, Outcome.SUCCESS)
            assertEquals(c.javaClass.simpleName, 0.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)

            c.updateOutcome(eid, Outcome.ROLLED_BACK)
            assertEquals(c.javaClass.simpleName, 0.5, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `outcome can be corrected twice`() {
        for (c in bothCalculators()) {
            val eid = c.recordDeploy("svc", 1000.0, Outcome.PENDING)
            c.updateOutcome(eid, Outcome.FAILED)
            assertEquals(1.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)
            c.updateOutcome(eid, Outcome.SUCCESS) // false alarm reversed
            assertEquals(0.0, c.changeFailureRate("svc", 1010.0)!!, 0.0)
        }
    }

    @Test
    fun `update unknown event id throws`() {
        for (c in bothCalculators()) {
            try {
                c.updateOutcome(99999, Outcome.FAILED)
                fail("${c.javaClass.simpleName}: expected an exception for an unknown event id")
            } catch (e: NoSuchElementException) {
                // expected
            }
        }
    }

    @Test
    fun `rolling window forgets old deploys`() {
        for (c in bothCalculators()) {
            c.recordDeploy("svc", 0.0, Outcome.FAILED)
            assertEquals(1.0, c.changeFailureRate("svc", 1000.0)!!, 0.0)
            val farFuture = WINDOW_SECONDS + 1
            c.recordDeploy("svc", farFuture, Outcome.SUCCESS)
            assertEquals(0.0, c.changeFailureRate("svc", farFuture)!!, 0.0)
        }
    }
}

class P5DifferentialTest {
    @Test
    fun `optimal matches brute force over random stream with late corrections`() {
        val rng = Random(7)
        val opt = OptimalCFRCalculator()
        val brute = BruteForceCFRCalculator()
        val services = listOf("svc-a", "svc-b", "svc-c")
        val outcomes = listOf(Outcome.SUCCESS, Outcome.FAILED, Outcome.ROLLED_BACK, Outcome.PENDING)

        data class IdPair(val optId: Int, val bruteId: Int)
        val recorded = mutableListOf<IdPair>()

        var now = 0.0
        repeat(300) {
            now += 1 + rng.nextDouble() * (WINDOW_SECONDS / 20)
            val svc = services[rng.nextInt(services.size)]
            val outcome = outcomes[rng.nextInt(outcomes.size)]
            val oid = opt.recordDeploy(svc, now, outcome)
            val bid = brute.recordDeploy(svc, now, outcome)
            recorded.add(IdPair(oid, bid))

            if (recorded.isNotEmpty() && rng.nextDouble() < 0.2) {
                val pair = recorded[rng.nextInt(recorded.size)]
                val newOutcome = listOf(Outcome.SUCCESS, Outcome.FAILED, Outcome.ROLLED_BACK)[rng.nextInt(3)]
                try {
                    opt.updateOutcome(pair.optId, newOutcome)
                    brute.updateOutcome(pair.bruteId, newOutcome)
                } catch (e: NoSuchElementException) {
                    // fine if already evicted
                }
            }

            for (svc in services) {
                val orv = opt.changeFailureRate(svc, now)
                val brv = brute.changeFailureRate(svc, now)
                assertEquals("mismatch for $svc at t=$now", brv, orv)
            }
        }
    }

    @Test
    fun `concurrent inserts do not corrupt counters`() {
        val c = OptimalCFRCalculator()
        val n = 20
        val barrier = CyclicBarrier(n)
        val threads = (0 until n).map { i ->
            Thread {
                barrier.await()
                val outcome = if (i % 2 == 0) Outcome.FAILED else Outcome.SUCCESS
                c.recordDeploy("svc-concurrent", 1000.0 + i * 0.001, outcome)
            }
        }
        threads.forEach { it.start() }
        threads.forEach { it.join() }

        assertEquals(0.5, c.changeFailureRate("svc-concurrent", 2000.0)!!, 0.0)
    }
}
