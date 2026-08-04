// Package p5 implements a rolling 7-day change-failure-rate (CFR)
// calculator per service, handling deploy outcomes that are discovered
// after the deploy was originally recorded.
package p1

import (
	"container/list"
	"fmt"
	"sync"
)

type Outcome string

const (
	Pending    Outcome = "PENDING"
	Success    Outcome = "SUCCESS"
	Failed     Outcome = "FAILED"
	RolledBack Outcome = "ROLLED_BACK"
)

func isFailure(o Outcome) bool { return o == Failed || o == RolledBack }

const WindowSeconds = 7 * 24 * 60 * 60.0

type DeployEvent struct {
	ID          int
	ServiceName string
	Timestamp   float64
	Outcome     Outcome
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// One flat slice of every event, across every service, forever.
// UpdateOutcome scans it linearly to find the matching ID.
// ChangeFailureRate re-scans and re-filters ALL events on every query.
// ---------------------------------------------------------------------------
type BruteForceCFRCalculator struct {
	mu     sync.Mutex
	events []*DeployEvent
	nextID int
}

func NewBruteForceCFRCalculator() *BruteForceCFRCalculator {
	return &BruteForceCFRCalculator{}
}

func (c *BruteForceCFRCalculator) RecordDeploy(serviceName string, ts float64, outcome Outcome) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	c.events = append(c.events, &DeployEvent{ID: c.nextID, ServiceName: serviceName, Timestamp: ts, Outcome: outcome})
	return c.nextID
}

func (c *BruteForceCFRCalculator) UpdateOutcome(id int, outcome Outcome) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.events { // O(n) scan to find the event
		if e.ID == id {
			e.Outcome = outcome
			return nil
		}
	}
	return fmt.Errorf("no such event id: %d", id)
}

func (c *BruteForceCFRCalculator) ChangeFailureRate(serviceName string, now float64) (rate float64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	windowStart := now - WindowSeconds
	var resolved, failures int
	for _, e := range c.events { // O(n) scan of ALL events, all services
		if e.ServiceName != serviceName || e.Timestamp < windowStart || e.Timestamp > now {
			continue
		}
		if e.Outcome == Pending {
			continue
		}
		resolved++
		if isFailure(e.Outcome) {
			failures++
		}
	}
	if resolved == 0 {
		return 0, false
	}
	return float64(failures) / float64(resolved), true
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Three fixes:
//  1. Partition storage per service: a map of deques, so a query for one
//     service never scans another service's events.
//  2. Maintain running counters (resolvedCount, failureCount) per service,
//     adjusted incrementally as events enter, get corrected, and age out
//     -> O(1) query after amortized-O(1) eviction, instead of O(n).
//  3. Index events by ID (byID map) so UpdateOutcome is O(1) to locate the
//     record, not an O(n) scan.
//
// A PENDING event contributes to neither the numerator nor denominator
// until resolved — a deliberate modeling choice: counting it as an
// implicit success would bias CFR artificially low right after a burst of
// deploys, before outcomes are known.
// ---------------------------------------------------------------------------
type serviceCounters struct {
	events   *list.List // of *DeployEvent, oldest at Front()
	resolved int
	failures int
}

type OptimalCFRCalculator struct {
	mu         sync.Mutex
	byID       map[int]*list.Element
	byService  map[int]string // event ID -> service, so we can find its list
	perService map[string]*serviceCounters
	nextID     int
}

func NewOptimalCFRCalculator() *OptimalCFRCalculator {
	return &OptimalCFRCalculator{
		byID:       make(map[int]*list.Element),
		byService:  make(map[int]string),
		perService: make(map[string]*serviceCounters),
	}
}

func (c *OptimalCFRCalculator) counters(serviceName string) *serviceCounters {
	sc, ok := c.perService[serviceName]
	if !ok {
		sc = &serviceCounters{events: list.New()}
		c.perService[serviceName] = sc
	}
	return sc
}

func (c *OptimalCFRCalculator) RecordDeploy(serviceName string, ts float64, outcome Outcome) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	event := &DeployEvent{ID: c.nextID, ServiceName: serviceName, Timestamp: ts, Outcome: outcome}
	sc := c.counters(serviceName)
	elem := sc.events.PushBack(event)
	c.byID[event.ID] = elem
	c.byService[event.ID] = serviceName
	if outcome != Pending {
		sc.resolved++
		if isFailure(outcome) {
			sc.failures++
		}
	}
	return event.ID
}

func (c *OptimalCFRCalculator) UpdateOutcome(id int, outcome Outcome) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.byID[id]
	if !ok {
		return fmt.Errorf("no such event id: %d", id)
	}
	event := elem.Value.(*DeployEvent)
	sc := c.perService[event.ServiceName]

	wasResolved := event.Outcome != Pending
	wasFailure := isFailure(event.Outcome)
	event.Outcome = outcome
	isResolved := outcome != Pending
	nowFailure := isFailure(outcome)

	if wasResolved && !isResolved {
		sc.resolved--
	} else if !wasResolved && isResolved {
		sc.resolved++
	}
	if wasFailure && !nowFailure {
		sc.failures--
	} else if !wasFailure && nowFailure {
		sc.failures++
	}
	return nil
}

func (c *OptimalCFRCalculator) evictExpired(serviceName string, now float64) {
	sc, ok := c.perService[serviceName]
	if !ok {
		return
	}
	windowStart := now - WindowSeconds
	for e := sc.events.Front(); e != nil; {
		event := e.Value.(*DeployEvent)
		if event.Timestamp >= windowStart {
			break
		}
		next := e.Next()
		sc.events.Remove(e)
		delete(c.byID, event.ID)
		delete(c.byService, event.ID)
		if event.Outcome != Pending {
			sc.resolved--
			if isFailure(event.Outcome) {
				sc.failures--
			}
		}
		e = next
	}
}

func (c *OptimalCFRCalculator) ChangeFailureRate(serviceName string, now float64) (rate float64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictExpired(serviceName, now)
	sc, exists := c.perService[serviceName]
	if !exists || sc.resolved == 0 {
		return 0, false
	}
	return float64(sc.failures) / float64(sc.resolved), true
}
