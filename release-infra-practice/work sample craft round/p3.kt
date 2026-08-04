package p3

import java.util.ArrayDeque

/**
 * Canary Analysis Evaluator.
 *
 * Given a stream of (timestamp, errorCount, requestCount) points for a
 * canary and a baseline, decide PASS/FAIL/CONTINUE based on error-rate
 * delta over a rolling time window, tolerating sparse/missing data.
 */

enum class Verdict { PASS, FAIL, CONTINUE }

data class MetricPoint(val timestampSeconds: Double, val errorCount: Int, val requestCount: Int)

const val WINDOW_SECONDS = 5.0 * 60.0
const val MIN_SAMPLE_SIZE = 50            // per side, within the window
const val FAIL_THRESHOLD_DELTA = 0.05     // canary must not exceed baseline error rate by more than this

private fun errorRate(errors: Int, requests: Int): Double? =
    if (requests == 0) null else errors.toDouble() / requests

private fun decide(cErrors: Int, cRequests: Int, bErrors: Int, bRequests: Int): Verdict {
    if (cRequests < MIN_SAMPLE_SIZE || bRequests < MIN_SAMPLE_SIZE) return Verdict.CONTINUE
    val cRate = errorRate(cErrors, cRequests) ?: return Verdict.CONTINUE
    val bRate = errorRate(bErrors, bRequests) ?: return Verdict.CONTINUE
    return if (cRate - bRate > FAIL_THRESHOLD_DELTA) Verdict.FAIL else Verdict.PASS
}

interface CanaryEvaluator {
    fun recordCanary(timestampSeconds: Double, errorCount: Int, requestCount: Int)
    fun recordBaseline(timestampSeconds: Double, errorCount: Int, requestCount: Int)
    fun evaluate(nowSeconds: Double): Verdict
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// Keep every point ever received in two lists. evaluate() re-filters BOTH
// lists down to the window on every call and re-sums from scratch: O(n)
// per evaluation, where n = total points ever recorded, growing without
// bound as the canary runs longer.
// ---------------------------------------------------------------------------
class BruteForceCanaryEvaluator : CanaryEvaluator {
    private val canaryPoints = mutableListOf<MetricPoint>()
    private val baselinePoints = mutableListOf<MetricPoint>()

    override fun recordCanary(timestampSeconds: Double, errorCount: Int, requestCount: Int) {
        canaryPoints.add(MetricPoint(timestampSeconds, errorCount, requestCount))
    }

    override fun recordBaseline(timestampSeconds: Double, errorCount: Int, requestCount: Int) {
        baselinePoints.add(MetricPoint(timestampSeconds, errorCount, requestCount))
    }

    private fun sumInWindow(points: List<MetricPoint>, now: Double): Pair<Int, Int> {
        val windowStart = now - WINDOW_SECONDS
        var errors = 0
        var requests = 0
        for (p in points) { // O(n) scan + rebuild, every call
            if (p.timestampSeconds in windowStart..now) {
                errors += p.errorCount
                requests += p.requestCount
            }
        }
        return errors to requests
    }

    override fun evaluate(nowSeconds: Double): Verdict {
        val (cErrors, cRequests) = sumInWindow(canaryPoints, nowSeconds)
        val (bErrors, bRequests) = sumInWindow(baselinePoints, nowSeconds)
        return decide(cErrors, cRequests, bErrors, bRequests)
    }
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Maintain an ArrayDeque plus running sums PER SIDE. On insert: push back,
// add to running sums -> O(1). Before evaluating: evict from the front
// while expired, subtracting each evicted point's contribution ->
// amortized O(1) (each point pushed once, popped at most once, ever).
// evaluate() becomes O(1) after eviction, instead of O(n).
//
// Assumes points arrive in non-decreasing timestamp order per side — true
// for a live metrics stream — stated explicitly as a scope boundary rather
// than silently assumed.
// ---------------------------------------------------------------------------
private class SideWindow {
    val points = ArrayDeque<MetricPoint>()
    var errorSum = 0
    var requestSum = 0

    fun record(p: MetricPoint) {
        points.offer(p)
        errorSum += p.errorCount
        requestSum += p.requestCount
    }

    fun evictExpired(now: Double) {
        val windowStart = now - WINDOW_SECONDS
        while (points.isNotEmpty() && points.peek().timestampSeconds < windowStart) {
            val p = points.poll()
            errorSum -= p.errorCount
            requestSum -= p.requestCount
        }
    }
}

class OptimalCanaryEvaluator : CanaryEvaluator {
    private val canary = SideWindow()
    private val baseline = SideWindow()

    override fun recordCanary(timestampSeconds: Double, errorCount: Int, requestCount: Int) {
        canary.record(MetricPoint(timestampSeconds, errorCount, requestCount))
    }

    override fun recordBaseline(timestampSeconds: Double, errorCount: Int, requestCount: Int) {
        baseline.record(MetricPoint(timestampSeconds, errorCount, requestCount))
    }

    override fun evaluate(nowSeconds: Double): Verdict {
        canary.evictExpired(nowSeconds)
        baseline.evictExpired(nowSeconds)
        return decide(canary.errorSum, canary.requestSum, baseline.errorSum, baseline.requestSum)
    }
}
