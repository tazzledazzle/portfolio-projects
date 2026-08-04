package p1

import java.io.File
import java.io.RandomAccessFile
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

/**
 * Idempotent Deploy-Trigger API.
 *
 * Duplicate "trigger deployment" requests keyed by (idempotencyKey,
 * serviceName, gitSha) must never cause a second deployment. Status must be
 * queryable. The service must survive a crash mid-deploy without losing the
 * dedup guarantee or silently re-running a completed deploy.
 */

enum class DeployStatus { PENDING, RUNNING, SUCCEEDED, FAILED, UNKNOWN }

data class DeploymentRecord(
    val idempotencyKey: String,
    val serviceName: String,
    val gitSha: String,
    var status: DeployStatus,
    var error: String? = null,
    var createdAtMillis: Long,
    var updatedAtMillis: Long
) {
    // Minimal hand-rolled serialization (no JSON library dependency): one
    // record per line, pipe-delimited, with '~' escaping any literal pipe.
    fun toLine(): String {
        fun esc(s: String) = s.replace("|", "~PIPE~")
        return listOf(
            esc(idempotencyKey), esc(serviceName), esc(gitSha), status.name,
            esc(error ?: ""), createdAtMillis.toString(), updatedAtMillis.toString()
        ).joinToString("|")
    }

    companion object {
        fun fromLine(line: String): DeploymentRecord {
            fun unesc(s: String) = s.replace("~PIPE~", "|")
            val parts = line.split("|")
            return DeploymentRecord(
                idempotencyKey = unesc(parts[0]),
                serviceName = unesc(parts[1]),
                gitSha = unesc(parts[2]),
                status = DeployStatus.valueOf(parts[3]),
                error = unesc(parts[4]).ifEmpty { null },
                createdAtMillis = parts[5].toLong(),
                updatedAtMillis = parts[6].toLong()
            )
        }
    }
}

typealias DeployFn = (serviceName: String, gitSha: String) -> Unit

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// A plain HashMap, no synchronization at all. "Check if the key exists,
// then insert" is two unsynchronized steps -> classic check-then-act race:
// two concurrent callers can both see the key absent before either has
// written it. Nothing survives a restart, since it's all in one process's
// heap.
// ---------------------------------------------------------------------------
class BruteForceDeployService(private val deploy: DeployFn) {
    private val store = HashMap<String, DeploymentRecord>()

    fun trigger(key: String, serviceName: String, gitSha: String): DeploymentRecord {
        // --- THE BUG: check-then-act, no lock. ---
        store[key]?.let { return it }

        // A realistic gap between "checked" and "wrote" — validation,
        // logging, and object construction all take nonzero time.
        Thread.sleep(10)

        val now = System.currentTimeMillis()
        val record = DeploymentRecord(
            key, serviceName, gitSha, DeployStatus.RUNNING,
            createdAtMillis = now, updatedAtMillis = now
        )
        store[key] = record // unsynchronized write to a plain HashMap

        try {
            deploy(serviceName, gitSha)
            record.status = DeployStatus.SUCCEEDED
        } catch (e: Exception) {
            record.status = DeployStatus.FAILED
            record.error = e.message
        }
        record.updatedAtMillis = System.currentTimeMillis()
        return record
    }

    fun getStatus(key: String): DeploymentRecord? = store[key]
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Two independent fixes:
//   1. Atomicity: ConcurrentHashMap.putIfAbsent() is a single atomic
//      "insert if absent" operation — exactly the JVM-native equivalent of
//      SQL's INSERT ... ON CONFLICT DO NOTHING. Whoever's putIfAbsent
//      actually installs a NEW value (returns null) is the sole owner of
//      that deploy; everyone else gets back the existing record and must
//      not deploy again.
//   2. Durability: every state transition is appended to a write-ahead
//      log file BEFORE (or as part of) the in-memory update, fsync'd via
//      FileDescriptor.sync(). On startup, the log is replayed to
//      reconstruct state, so a process restart does not lose dedup
//      history. In production this role is normally played by a real
//      database's own durability (Postgres/SQLite + a UNIQUE constraint);
//      this models the same durability property from JDK stdlib only.
//
// Crash-mid-deploy handling: a RUNNING record whose updatedAtMillis is
// older than staleTimeoutMillis gets flipped to UNKNOWN on reconciliation
// (called at startup) rather than being silently retried (double-deploy
// risk) or silently treated as success (worse) — the only correct move
// when the outcome genuinely cannot be known without checking the real
// deployment target.
// ---------------------------------------------------------------------------
class OptimalDeployService(
    private val deploy: DeployFn,
    private val walPath: String,
    private val staleTimeoutMillis: Long
) {
    private val store = ConcurrentHashMap<String, DeploymentRecord>()
    private val perKeyLocks = ConcurrentHashMap<String, ReentrantLock>()
    private var walFile: RandomAccessFile

    init {
        replayWal()
        walFile = RandomAccessFile(walPath, "rw")
        walFile.seek(walFile.length()) // append mode
        reconcileStaleRunning()
    }

    private fun replayWal() {
        val f = File(walPath)
        if (!f.exists()) return
        f.forEachLine { line ->
            if (line.isNotBlank()) {
                try {
                    val record = DeploymentRecord.fromLine(line)
                    store[record.idempotencyKey] = record // last write for a key wins on replay
                } catch (e: Exception) {
                    // tolerate a torn last write from a crash mid-append
                }
            }
        }
    }

    private fun appendWal(record: DeploymentRecord) {
        synchronized(walFile) {
            walFile.write((record.toLine() + "\n").toByteArray())
            walFile.fd.sync() // durability: on disk before we proceed
        }
    }

    private fun lockFor(key: String): ReentrantLock =
        perKeyLocks.computeIfAbsent(key) { ReentrantLock() }

    private fun reconcileStaleRunning() {
        val cutoff = System.currentTimeMillis() - staleTimeoutMillis
        for (record in store.values) {
            if (record.status == DeployStatus.RUNNING && record.updatedAtMillis < cutoff) {
                lockFor(record.idempotencyKey).withLock {
                    if (record.status == DeployStatus.RUNNING && record.updatedAtMillis < cutoff) {
                        record.status = DeployStatus.UNKNOWN
                        record.updatedAtMillis = System.currentTimeMillis()
                        appendWal(record)
                    }
                }
            }
        }
    }

    fun trigger(key: String, serviceName: String, gitSha: String): DeploymentRecord {
        val now = System.currentTimeMillis()
        val candidate = DeploymentRecord(
            key, serviceName, gitSha, DeployStatus.RUNNING,
            createdAtMillis = now, updatedAtMillis = now
        )
        // Atomic insert-if-absent: exactly one caller's putIfAbsent
        // returns null (meaning ITS candidate was installed); everyone
        // else gets back the winner's record and must not deploy again.
        val previous = store.putIfAbsent(key, candidate)
        if (previous != null) {
            return previous
        }
        appendWal(candidate)

        lockFor(key).withLock {
            try {
                deploy(serviceName, gitSha)
                candidate.status = DeployStatus.SUCCEEDED
            } catch (e: Exception) {
                candidate.status = DeployStatus.FAILED
                candidate.error = e.message
            }
            candidate.updatedAtMillis = System.currentTimeMillis()
            appendWal(candidate)
        }
        return candidate
    }

    fun getStatus(key: String): DeploymentRecord? = store[key]

    fun close() {
        walFile.close()
    }

    // Test-only hook: simulate "the process crashed after writing a
    // RUNNING record but before it resolved."
    internal fun injectRawRunningRecord(key: String, serviceName: String, gitSha: String, updatedAtMillis: Long) {
        val record = DeploymentRecord(key, serviceName, gitSha, DeployStatus.RUNNING,
            createdAtMillis = updatedAtMillis, updatedAtMillis = updatedAtMillis)
        store[key] = record
        appendWal(record)
    }

    fun simulateCrashAndRestart() {
        reconcileStaleRunning()
    }
}
