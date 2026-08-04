package p1

import (
	"math/rand"
	"testing"
)

type evaluator interface {
	RecordCanary(ts float64, errs, reqs int)
	RecordBaseline(ts float64, errs, reqs int)
	Evaluate(now float64) Verdict
}

func makeBoth() []evaluator {
	return []evaluator{NewBruteForceCanaryEvaluator(), NewOptimalCanaryEvaluator()}
}

func TestContract_InsufficientSamplesIsContinue(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 1, 2)
		ev.RecordBaseline(0, 0, 2)
		if got := ev.Evaluate(0); got != Continue {
			t.Errorf("%T: expected CONTINUE, got %s", ev, got)
		}
	}
}

func TestContract_ZeroTrafficIsContinueNotPass(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 0, 0)
		ev.RecordBaseline(0, 0, 100)
		if got := ev.Evaluate(0); got != Continue {
			t.Errorf("%T: zero canary traffic must be CONTINUE, not a false PASS, got %s", ev, got)
		}
	}
}

func TestContract_EqualRatesPass(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 5, 100)
		ev.RecordBaseline(0, 5, 100)
		if got := ev.Evaluate(0); got != Pass {
			t.Errorf("%T: expected PASS, got %s", ev, got)
		}
	}
}

func TestContract_ExactlyAtThresholdIsPass(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 10, 100)  // 10%
		ev.RecordBaseline(0, 5, 100) // 5% -> delta exactly 0.05
		if got := ev.Evaluate(0); got != Pass {
			t.Errorf("%T: exactly-at-threshold should be PASS (strict >), got %s", ev, got)
		}
	}
}

func TestContract_JustOverThresholdIsFail(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 11, 100)  // 11%
		ev.RecordBaseline(0, 5, 100) // 5% -> delta 0.06
		if got := ev.Evaluate(0); got != Fail {
			t.Errorf("%T: expected FAIL, got %s", ev, got)
		}
	}
}

func TestContract_CanaryBetterThanBaselinePasses(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 1, 100)
		ev.RecordBaseline(0, 10, 100)
		if got := ev.Evaluate(0); got != Pass {
			t.Errorf("%T: expected PASS, got %s", ev, got)
		}
	}
}

func TestContract_PointsOutsideWindowAreIgnored(t *testing.T) {
	for _, ev := range makeBoth() {
		ev.RecordCanary(0, 1000, 1000) // ancient spike, long before the window
		for ts := 1000.0; ts < 1060; ts++ {
			ev.RecordCanary(ts, 0, 2)
			ev.RecordBaseline(ts, 0, 2)
		}
		if got := ev.Evaluate(1060); got != Pass {
			t.Errorf("%T: expected PASS (old spike outside window ignored), got %s", ev, got)
		}
	}
}

func TestContract_SparseDataStillAccumulatesToAVerdict(t *testing.T) {
	for _, ev := range makeBoth() {
		for _, ts := range []float64{0, 90, 180, 270, 300, 330} {
			ev.RecordCanary(ts, 1, 20)
			ev.RecordBaseline(ts, 1, 20)
		}
		got := ev.Evaluate(330)
		if got != Pass && got != Fail {
			t.Errorf("%T: expected a real verdict once enough sparse samples accumulate, got %s", ev, got)
		}
	}
}

func TestDifferential_OptimalMatchesBruteForceOverRandomStream(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	opt := NewOptimalCanaryEvaluator()
	brute := NewBruteForceCanaryEvaluator()

	now := 0.0
	for i := 0; i < 500; i++ {
		now += 0.1 + rng.Float64()*2.9
		errs, reqs := rng.Intn(4), 1+rng.Intn(20)
		opt.RecordCanary(now, errs, reqs)
		brute.RecordCanary(now, errs, reqs)

		errs2, reqs2 := rng.Intn(4), 1+rng.Intn(20)
		opt.RecordBaseline(now, errs2, reqs2)
		brute.RecordBaseline(now, errs2, reqs2)

		want := brute.Evaluate(now)
		got := opt.Evaluate(now)
		if got != want {
			t.Fatalf("mismatch at t=%f: brute=%s optimal=%s", now, want, got)
		}
	}
}
