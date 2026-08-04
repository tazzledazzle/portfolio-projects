// Package p3 implements the canary analysis evaluator: given a stream of
// (timestamp, error_count, request_count) points for a canary and a
// baseline, decide PASS/FAIL/CONTINUE based on error-rate delta over a
// rolling time window, tolerating sparse/missing data.
package p1

import (
	"container/list"
)

type Verdict string

const (
	Pass     Verdict = "PASS"
	Fail     Verdict = "FAIL"
	Continue Verdict = "CONTINUE" // not enough data yet either way
)

type MetricPoint struct {
	TimestampSeconds float64
	ErrorCount       int
	RequestCount     int
}

const (
	WindowSeconds      = 5 * 60.0
	MinSampleSize      = 50   // per side, within the window
	FailThresholdDelta = 0.05 // canary must not exceed baseline error rate by more than this
)

func errorRate(errs, reqs int) (float64, bool) {
	if reqs == 0 {
		return 0, false
	}
	return float64(errs) / float64(reqs), true
}

func decide(cErrs, cReqs, bErrs, bReqs int) Verdict {
	if cReqs < MinSampleSize || bReqs < MinSampleSize {
		return Continue
	}
	cRate, ok1 := errorRate(cErrs, cReqs)
	bRate, ok2 := errorRate(bErrs, bReqs)
	if !ok1 || !ok2 {
		return Continue
	}
	if cRate-bRate > FailThresholdDelta {
		return Fail
	}
	return Pass
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// Keep every point ever received in two slices. Evaluate() re-filters BOTH
// slices down to the window on every call and re-sums from scratch: O(n)
// per evaluation, where n = total points ever recorded, growing without
// bound as the canary runs longer.
// ---------------------------------------------------------------------------
type BruteForceCanaryEvaluator struct {
	canaryPoints   []MetricPoint
	baselinePoints []MetricPoint
}

func NewBruteForceCanaryEvaluator() *BruteForceCanaryEvaluator {
	return &BruteForceCanaryEvaluator{}
}

func (e *BruteForceCanaryEvaluator) RecordCanary(ts float64, errs, reqs int) {
	e.canaryPoints = append(e.canaryPoints, MetricPoint{ts, errs, reqs})
}

func (e *BruteForceCanaryEvaluator) RecordBaseline(ts float64, errs, reqs int) {
	e.baselinePoints = append(e.baselinePoints, MetricPoint{ts, errs, reqs})
}

func sumInWindow(points []MetricPoint, now float64) (errs, reqs int) {
	windowStart := now - WindowSeconds
	for _, p := range points {
		if p.TimestampSeconds >= windowStart && p.TimestampSeconds <= now {
			errs += p.ErrorCount
			reqs += p.RequestCount
		}
	}
	return
}

func (e *BruteForceCanaryEvaluator) Evaluate(now float64) Verdict {
	cErrs, cReqs := sumInWindow(e.canaryPoints, now)
	bErrs, bReqs := sumInWindow(e.baselinePoints, now)
	return decide(cErrs, cReqs, bErrs, bReqs)
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Maintain a deque (container/list.List) plus running sums PER SIDE. On
// insert: push back, add to running sums -> O(1). Before evaluating: evict
// from the front while expired, subtracting each evicted point's
// contribution -> amortized O(1) (each point pushed once, popped at most
// once, ever). Evaluate() becomes O(1) after eviction, instead of O(n).
//
// This assumes points arrive in non-decreasing timestamp order per side —
// true for a live metrics stream — stated explicitly as a scope boundary
// rather than silently assumed.
// ---------------------------------------------------------------------------
type sideWindow struct {
	points     *list.List
	errorSum   int
	requestSum int
}

func newSideWindow() *sideWindow {
	return &sideWindow{points: list.New()}
}

func (w *sideWindow) record(p MetricPoint) {
	w.points.PushBack(p)
	w.errorSum += p.ErrorCount
	w.requestSum += p.RequestCount
}

func (w *sideWindow) evictExpired(now float64) {
	windowStart := now - WindowSeconds
	for e := w.points.Front(); e != nil; {
		p := e.Value.(MetricPoint)
		if p.TimestampSeconds >= windowStart {
			break
		}
		next := e.Next()
		w.points.Remove(e)
		w.errorSum -= p.ErrorCount
		w.requestSum -= p.RequestCount
		e = next
	}
}

type OptimalCanaryEvaluator struct {
	canary   *sideWindow
	baseline *sideWindow
}

func NewOptimalCanaryEvaluator() *OptimalCanaryEvaluator {
	return &OptimalCanaryEvaluator{canary: newSideWindow(), baseline: newSideWindow()}
}

func (e *OptimalCanaryEvaluator) RecordCanary(ts float64, errs, reqs int) {
	e.canary.record(MetricPoint{ts, errs, reqs})
}

func (e *OptimalCanaryEvaluator) RecordBaseline(ts float64, errs, reqs int) {
	e.baseline.record(MetricPoint{ts, errs, reqs})
}

func (e *OptimalCanaryEvaluator) Evaluate(now float64) Verdict {
	e.canary.evictExpired(now)
	e.baseline.evictExpired(now)
	return decide(e.canary.errorSum, e.canary.requestSum, e.baseline.errorSum, e.baseline.requestSum)
}
