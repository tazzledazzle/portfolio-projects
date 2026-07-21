package main

import (
	"errors"
	"fmt"
	"sync"
)

var ErrEmptyRequest = errors.New("kind and summary required")

type ServiceRequest struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Summary       string `json:"summary"`
	TicketRemoved bool   `json:"ticket_removed"`
	Path          string `json:"path"`
}

type AutomationMetrics struct {
	TicketsBefore      int            `json:"tickets_before"`
	TicketsAfter       int            `json:"tickets_after"`
	Requests           int            `json:"requests"`
	TicketRemovedPath  bool           `json:"ticket_removed_path"`
	AutomationMetrics  map[string]any `json:"automation_metrics"`
}

type SelfServiceStore struct {
	mu            sync.Mutex
	ticketsBefore int
	ticketsAfter  int
	requests      int
	counter       int
	items         map[string]*ServiceRequest
}

func NewSelfServiceStore(baselineTickets int) *SelfServiceStore {
	if baselineTickets < 0 {
		baselineTickets = 0
	}
	return &SelfServiceStore{
		ticketsBefore: baselineTickets,
		ticketsAfter:  baselineTickets,
		items:         make(map[string]*ServiceRequest),
	}
}

func (s *SelfServiceStore) Submit(kind, summary string) (*ServiceRequest, error) {
	if kind == "" || summary == "" {
		return nil, ErrEmptyRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	s.requests++
	if s.ticketsAfter > 0 {
		s.ticketsAfter--
	}
	req := &ServiceRequest{
		ID:            fmt.Sprintf("req-%d", s.counter),
		Kind:          kind,
		Summary:       summary,
		TicketRemoved: true,
		Path:          "self_service",
	}
	s.items[req.ID] = req
	out := *req
	return &out, nil
}

func (s *SelfServiceStore) Metrics() AutomationMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AutomationMetrics{
		TicketsBefore:     s.ticketsBefore,
		TicketsAfter:      s.ticketsAfter,
		Requests:          s.requests,
		TicketRemovedPath: s.requests > 0,
		AutomationMetrics: map[string]any{
			"tickets_before": s.ticketsBefore,
			"tickets_after":  s.ticketsAfter,
			"requests":       s.requests,
			"ticket_removed": s.requests > 0,
		},
	}
}
