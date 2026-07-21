package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestSelfService_Submit_RemovesTicket(t *testing.T) {
	store := NewSelfServiceStore(10)
	req, err := store.Submit("env-provision", "need staging")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if !req.TicketRemoved {
		t.Fatalf("expected ticket_removed path: %#v", req)
	}
	m := store.Metrics()
	if !m.TicketRemovedPath {
		t.Fatalf("metrics missing ticket_removed path: %#v", m)
	}
}

func TestSelfService_BeforeAfterMetrics(t *testing.T) {
	store := NewSelfServiceStore(5)
	before := store.Metrics()
	if before.TicketsBefore != 5 {
		t.Fatalf("expected tickets_before=5, got %#v", before)
	}
	_, _ = store.Submit("access", "grant")
	_, _ = store.Submit("access", "grant-2")
	after := store.Metrics()
	if after.TicketsAfter >= after.TicketsBefore {
		t.Fatalf("tickets_after should drop: %#v", after)
	}
	if after.AutomationMetrics["tickets_before"] == nil || after.AutomationMetrics["tickets_after"] == nil {
		t.Fatalf("automation_metrics incomplete: %#v", after.AutomationMetrics)
	}
}

func TestSelfService_ConcurrentSubmit(t *testing.T) {
	store := NewSelfServiceStore(100)
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := store.Submit("kind", fmt.Sprintf("req-%d", i)); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent: %v", err)
	}
	m := store.Metrics()
	if m.Requests != 40 {
		t.Fatalf("expected 40 requests, got %#v", m)
	}
}
