package p1

import (
	"math/rand"
	"sync"
	"testing"
)

type cfrCalculator interface {
	RecordDeploy(serviceName string, ts float64, outcome Outcome) int
	UpdateOutcome(id int, outcome Outcome) error
	ChangeFailureRate(serviceName string, now float64) (float64, bool)
}

func makeBoth() []cfrCalculator {
	return []cfrCalculator{NewBruteForceCFRCalculator(), NewOptimalCFRCalculator()}
}

func TestContract_NoEventsReturnsNotOK(t *testing.T) {
	for _, c := range makeBoth() {
		if _, ok := c.ChangeFailureRate("svc", 1000); ok {
			t.Errorf("%T: expected not-ok for a service with no events", c)
		}
	}
}

func TestContract_AllSuccessIsZero(t *testing.T) {
	for _, c := range makeBoth() {
		for i := 0; i < 5; i++ {
			c.RecordDeploy("svc", 1000+float64(i), Success)
		}
		rate, ok := c.ChangeFailureRate("svc", 1010)
		if !ok || rate != 0.0 {
			t.Errorf("%T: expected 0.0, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_AllFailedIsOne(t *testing.T) {
	for _, c := range makeBoth() {
		for i := 0; i < 5; i++ {
			c.RecordDeploy("svc", 1000+float64(i), Failed)
		}
		rate, ok := c.ChangeFailureRate("svc", 1010)
		if !ok || rate != 1.0 {
			t.Errorf("%T: expected 1.0, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_MixedOutcomesRate(t *testing.T) {
	for _, c := range makeBoth() {
		c.RecordDeploy("svc", 1000, Success)
		c.RecordDeploy("svc", 1001, Success)
		c.RecordDeploy("svc", 1002, Failed)
		c.RecordDeploy("svc", 1003, RolledBack)
		rate, ok := c.ChangeFailureRate("svc", 1010)
		if !ok || rate != 0.5 {
			t.Errorf("%T: expected 0.5, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_PendingExcludedFromCalculation(t *testing.T) {
	for _, c := range makeBoth() {
		c.RecordDeploy("svc", 1000, Success)
		c.RecordDeploy("svc", 1001, Pending)
		c.RecordDeploy("svc", 1002, Pending)
		rate, ok := c.ChangeFailureRate("svc", 1010)
		if !ok || rate != 0.0 {
			t.Errorf("%T: pending deploys must not dilute/inflate CFR, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_EventsOutsideWindowExcluded(t *testing.T) {
	for _, c := range makeBoth() {
		c.RecordDeploy("svc", 0, Failed) // ancient
		c.RecordDeploy("svc", WindowSeconds+100, Success)
		rate, ok := c.ChangeFailureRate("svc", WindowSeconds+100)
		if !ok || rate != 0.0 {
			t.Errorf("%T: old failure must not count, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_ExactlyWindowSecondsOldStillCounts(t *testing.T) {
	for _, c := range makeBoth() {
		now := 1_000_000.0
		c.RecordDeploy("svc", now-WindowSeconds, Failed)
		rate, ok := c.ChangeFailureRate("svc", now)
		if !ok || rate != 1.0 {
			t.Errorf("%T: an event exactly WindowSeconds old should still count, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_OneSecondOlderThanWindowExcluded(t *testing.T) {
	for _, c := range makeBoth() {
		now := 1_000_000.0
		c.RecordDeploy("svc", now-WindowSeconds-1, Failed)
		c.RecordDeploy("svc", now-WindowSeconds+1, Success)
		rate, ok := c.ChangeFailureRate("svc", now)
		if !ok || rate != 0.0 {
			t.Errorf("%T: an event older than WindowSeconds must not count, got %f (ok=%v)", c, rate, ok)
		}
	}
}

func TestContract_ServicesAreIndependent(t *testing.T) {
	for _, c := range makeBoth() {
		c.RecordDeploy("svc-a", 1000, Failed)
		c.RecordDeploy("svc-b", 1000, Success)
		ra, _ := c.ChangeFailureRate("svc-a", 1010)
		rb, _ := c.ChangeFailureRate("svc-b", 1010)
		if ra != 1.0 || rb != 0.0 {
			t.Errorf("%T: expected independent rates, got svc-a=%f svc-b=%f", c, ra, rb)
		}
	}
}

func TestContract_LateArrivingOutcomeUpdatesRate(t *testing.T) {
	for _, c := range makeBoth() {
		eid := c.RecordDeploy("svc", 1000, Pending)
		c.RecordDeploy("svc", 1001, Success)
		rate, _ := c.ChangeFailureRate("svc", 1010)
		if rate != 0.0 {
			t.Errorf("%T: expected 0.0 before late outcome, got %f", c, rate)
		}
		if err := c.UpdateOutcome(eid, RolledBack); err != nil {
			t.Fatal(err)
		}
		rate, _ = c.ChangeFailureRate("svc", 1010)
		if rate != 0.5 {
			t.Errorf("%T: expected 0.5 after late outcome discovered, got %f", c, rate)
		}
	}
}

func TestContract_OutcomeCorrectedTwice(t *testing.T) {
	for _, c := range makeBoth() {
		eid := c.RecordDeploy("svc", 1000, Pending)
		c.UpdateOutcome(eid, Failed)
		rate, _ := c.ChangeFailureRate("svc", 1010)
		if rate != 1.0 {
			t.Errorf("%T: expected 1.0 after marking FAILED, got %f", c, rate)
		}
		c.UpdateOutcome(eid, Success) // false alarm reversed
		rate, _ = c.ChangeFailureRate("svc", 1010)
		if rate != 0.0 {
			t.Errorf("%T: expected 0.0 after reversal, got %f", c, rate)
		}
	}
}

func TestContract_UpdateUnknownEventIDErrors(t *testing.T) {
	for _, c := range makeBoth() {
		if err := c.UpdateOutcome(99999, Failed); err == nil {
			t.Errorf("%T: expected an error for an unknown event id", c)
		}
	}
}

func TestContract_RollingWindowForgetsOldDeploys(t *testing.T) {
	for _, c := range makeBoth() {
		c.RecordDeploy("svc", 0, Failed)
		rate, _ := c.ChangeFailureRate("svc", 1000)
		if rate != 1.0 {
			t.Errorf("%T: expected 1.0 initially, got %f", c, rate)
		}
		farFuture := WindowSeconds + 1
		c.RecordDeploy("svc", farFuture, Success)
		rate, _ = c.ChangeFailureRate("svc", farFuture)
		if rate != 0.0 {
			t.Errorf("%T: expected the old failure to have aged out, got %f", c, rate)
		}
	}
}

func TestDifferential_OptimalMatchesBruteForceOverRandomStream(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	opt := NewOptimalCFRCalculator()
	brute := NewBruteForceCFRCalculator()
	services := []string{"svc-a", "svc-b", "svc-c"}
	outcomes := []Outcome{Success, Failed, RolledBack, Pending}

	type idPair struct{ optID, bruteID int }
	var recorded []idPair

	now := 0.0
	for i := 0; i < 300; i++ {
		now += 1 + rng.Float64()*(WindowSeconds/20)
		svc := services[rng.Intn(len(services))]
		outcome := outcomes[rng.Intn(len(outcomes))]
		oid := opt.RecordDeploy(svc, now, outcome)
		bid := brute.RecordDeploy(svc, now, outcome)
		recorded = append(recorded, idPair{oid, bid})

		if len(recorded) > 0 && rng.Float64() < 0.2 {
			idx := rng.Intn(len(recorded))
			pair := recorded[idx]
			newOutcome := []Outcome{Success, Failed, RolledBack}[rng.Intn(3)]
			_ = opt.UpdateOutcome(pair.optID, newOutcome)
			_ = brute.UpdateOutcome(pair.bruteID, newOutcome)
		}

		for _, svc := range services {
			or, ook := opt.ChangeFailureRate(svc, now)
			br, bok := brute.ChangeFailureRate(svc, now)
			if ook != bok || (ook && or != br) {
				t.Fatalf("mismatch for %s at t=%f: optimal=(%f,%v) brute=(%f,%v)", svc, now, or, ook, br, bok)
			}
		}
	}
}

func TestConcurrentInsertsDoNotCorruptCounters(t *testing.T) {
	c := NewOptimalCFRCalculator()
	const n = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			outcome := Failed
			if i%2 != 0 {
				outcome = Success
			}
			c.RecordDeploy("svc-concurrent", 1000+float64(i)*0.001, outcome)
		}()
	}
	close(start)
	wg.Wait()

	rate, ok := c.ChangeFailureRate("svc-concurrent", 2000)
	if !ok || rate != 0.5 {
		t.Fatalf("expected 0.5 after 20 concurrent inserts (10 failed/10 success), got %f (ok=%v)", rate, ok)
	}
}
