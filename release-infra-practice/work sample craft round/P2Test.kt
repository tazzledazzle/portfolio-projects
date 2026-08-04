package p2

import org.junit.Test
import org.junit.Assert.*
import java.util.concurrent.CyclicBarrier
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.ConcurrentLinkedQueue

class P2BruteForceTest {
    @Test
    fun `concurrent requests can exceed the limit`() {
        val audit = AppendOnlyAuditLog()
        val limiter = BruteForceRateLimiter(audit)
        val n = 10
        val barrier = CyclicBarrier(n)
        val accepted = AtomicInteger(0)
        val threads = (1..n).map {
            Thread {
                barrier.await()
                val r = limiter.tryRollback("checkout-service", System.currentTimeMillis())
                if (r.accepted) accepted.incrementAndGet()
            }
        }
        threads.forEach { it.start() }
        threads.forEach { it.join() }

        assertTrue("expected brute force to let more than $MAX_ROLLBACKS_PER_WINDOW through under a race, got ${accepted.get()}",
            accepted.get() > MAX_ROLLBACKS_PER_WINDOW)
    }
}

class P2OptimalTest {
    @Test
    fun `allows up to the limit`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        for (i in 0 until MAX_ROLLBACKS_PER_WINDOW) {
            val r = limiter.tryRollback("svc-a", base + i)
            assertTrue("rollback #${i + 1} should be accepted", r.accepted)
        }
    }

    @Test
    fun `rejects over limit with retry after`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        for (i in 0 until MAX_ROLLBACKS_PER_WINDOW) limiter.tryRollback("svc-a", base + i)
        val r = limiter.tryRollback("svc-a", base + MAX_ROLLBACKS_PER_WINDOW)
        assertFalse(r.accepted)
        assertNotNull(r.retryAfterSeconds)
        assertTrue(r.retryAfterSeconds!! > 0)
    }

    @Test
    fun `window boundary exact expiry`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        limiter.tryRollback("svc-a", base)
        limiter.tryRollback("svc-a", base + 1000)
        limiter.tryRollback("svc-a", base + 2000)
        val r = limiter.tryRollback("svc-a", base + WINDOW_MILLIS)
        assertTrue("oldest entry should have aged out by now", r.accepted)
    }

    @Test
    fun `one second before expiry still rejected`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        limiter.tryRollback("svc-a", base)
        limiter.tryRollback("svc-a", base + 1000)
        limiter.tryRollback("svc-a", base + 2000)
        val r = limiter.tryRollback("svc-a", base + WINDOW_MILLIS - 1000)
        assertFalse(r.accepted)
    }

    @Test
    fun `services have independent limits`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        for (i in 0 until MAX_ROLLBACKS_PER_WINDOW) limiter.tryRollback("svc-a", base + i)
        val r = limiter.tryRollback("svc-b", base)
        assertTrue("svc-b should be unaffected by svc-a's limit", r.accepted)
    }

    @Test
    fun `concurrent requests never exceed the limit`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val n = 20
        val barrier = CyclicBarrier(n)
        val accepted = AtomicInteger(0)
        val threads = (1..n).map {
            Thread {
                barrier.await()
                val r = limiter.tryRollback("svc-concurrent", System.currentTimeMillis())
                if (r.accepted) accepted.incrementAndGet()
            }
        }
        threads.forEach { it.start() }
        threads.forEach { it.join() }

        assertEquals(MAX_ROLLBACKS_PER_WINDOW, accepted.get())
    }

    @Test
    fun `accepted rollbacks are audit logged`() {
        val audit = AppendOnlyAuditLog()
        val limiter = OptimalRateLimiter(audit)
        val base = System.currentTimeMillis()
        for (i in 0 until MAX_ROLLBACKS_PER_WINDOW) limiter.tryRollback("svc-audit", base + i)
        limiter.tryRollback("svc-audit", base + MAX_ROLLBACKS_PER_WINDOW) // rejected

        val count = audit.allEntries().count { it.serviceName == "svc-audit" }
        assertEquals(MAX_ROLLBACKS_PER_WINDOW, count)
    }

    @Test
    fun `retry after shrinks over time`() {
        val limiter = OptimalRateLimiter(AppendOnlyAuditLog())
        val base = System.currentTimeMillis()
        for (i in 0 until MAX_ROLLBACKS_PER_WINDOW) limiter.tryRollback("svc-a", base + i)
        val r1 = limiter.tryRollback("svc-a", base + MAX_ROLLBACKS_PER_WINDOW)
        val r2 = limiter.tryRollback("svc-a", base + MAX_ROLLBACKS_PER_WINDOW + 30_000)
        assertTrue(r1.retryAfterSeconds!! > r2.retryAfterSeconds!!)
    }
}
