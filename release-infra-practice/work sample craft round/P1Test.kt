package p1

import org.junit.Test
import org.junit.Assert.*
import java.io.File
import java.util.concurrent.CyclicBarrier
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.CountDownLatch

class P1BruteForceTest {
    @Test
    fun `concurrent duplicate keys deploy more than once`() {
        val calls = AtomicInteger(0)
        val svc = BruteForceDeployService { _, _ ->
            Thread.sleep(5)
            calls.incrementAndGet()
        }

        val n = 20
        val barrier = CyclicBarrier(n)
        val threads = (1..n).map {
            Thread {
                barrier.await()
                svc.trigger("same-key", "checkout-service", "abc123")
            }
        }
        threads.forEach { it.start() }
        threads.forEach { it.join() }

        assertTrue("expected the brute-force version to double-deploy under a race, got ${calls.get()}",
            calls.get() > 1)
    }
}

class P1OptimalTest {
    private fun newService(deploy: DeployFn = { _, _ -> }): OptimalDeployService {
        val path = "/tmp/p1_kt_test_${System.nanoTime()}.wal"
        File(path).deleteOnExit()
        return OptimalDeployService(deploy, path, staleTimeoutMillis = 300)
    }

    @Test
    fun `duplicate key deploys exactly once under concurrency`() {
        val calls = AtomicInteger(0)
        val svc = newService { _, _ -> Thread.sleep(5); calls.incrementAndGet() }

        val n = 50
        val barrier = CyclicBarrier(n)
        val threads = (1..n).map {
            Thread {
                barrier.await()
                svc.trigger("same-key", "checkout-service", "abc123")
            }
        }
        threads.forEach { it.start() }
        threads.forEach { it.join() }

        assertEquals(1, calls.get())
        svc.close()
    }

    @Test
    fun `different keys deploy independently`() {
        val svc = newService()
        val r1 = svc.trigger("key-a", "svc-a", "sha1")
        val r2 = svc.trigger("key-b", "svc-b", "sha2")
        assertNotEquals(r1.idempotencyKey, r2.idempotencyKey)
        svc.close()
    }

    @Test
    fun `status query for unknown key returns null`() {
        val svc = newService()
        assertNull(svc.getStatus("never-seen"))
        svc.close()
    }

    @Test
    fun `failed deploy is recorded and not retried on duplicate request`() {
        val calls = AtomicInteger(0)
        val svc = newService { _, _ -> calls.incrementAndGet(); throw RuntimeException("deploy target unreachable") }

        val first = svc.trigger("bad-key", "svc", "sha")
        assertEquals(DeployStatus.FAILED, first.status)
        assertTrue(first.error!!.contains("unreachable"))

        val second = svc.trigger("bad-key", "svc", "sha")
        assertEquals(DeployStatus.FAILED, second.status)
        assertEquals(1, calls.get())
        svc.close()
    }

    @Test
    fun `crash mid deploy leaves status running then reconciles to unknown`() {
        val svc = newService()
        svc.injectRawRunningRecord("crashed-key", "svc", "sha", System.currentTimeMillis() - 1000)

        assertEquals(DeployStatus.RUNNING, svc.getStatus("crashed-key")!!.status)

        svc.simulateCrashAndRestart()

        assertEquals(DeployStatus.UNKNOWN, svc.getStatus("crashed-key")!!.status)
        svc.close()
    }

    @Test
    fun `recent running record is not falsely reconciled`() {
        val svc = newService()
        svc.injectRawRunningRecord("in-flight-key", "svc", "sha", System.currentTimeMillis())
        svc.simulateCrashAndRestart()
        assertEquals(DeployStatus.RUNNING, svc.getStatus("in-flight-key")!!.status)
        svc.close()
    }

    @Test
    fun `durability across restart`() {
        val path = "/tmp/p1_kt_durability_${System.nanoTime()}.wal"
        File(path).deleteOnExit()

        val svc1 = OptimalDeployService({ _, _ -> }, path, staleTimeoutMillis = 3_600_000)
        svc1.trigger("durable-key", "svc", "sha")
        svc1.close()

        // "Process restarts": brand-new instance, same WAL file.
        val svc2 = OptimalDeployService(
            { _, _ -> fail("deploy fn must NOT be called again after a restart for an already-resolved key") },
            path, staleTimeoutMillis = 3_600_000
        )
        val record = svc2.getStatus("durable-key")
        assertNotNull("expected the record to survive the restart via WAL replay", record)
        assertEquals(DeployStatus.SUCCEEDED, record!!.status)

        val r2 = svc2.trigger("durable-key", "svc", "sha")
        assertEquals(DeployStatus.SUCCEEDED, r2.status)
        svc2.close()
    }
}
