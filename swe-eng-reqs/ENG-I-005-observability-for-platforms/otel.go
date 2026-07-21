package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Span is an in-process, OTel-inspired trace record.
type Span struct {
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
}

// TraceExport is an honesty-labeled, OTLP-JSON-inspired export shape.
type TraceExport struct {
	InstrumentationModel string `json:"instrumentation_model"`
	Collector            string `json:"collector"`
	Spans                []Span `json:"spans"`
}

// AlertResult is the deterministic outcome of a fixture threshold rule.
type AlertResult struct {
	RuleID    string  `json:"rule_id"`
	Threshold float64 `json:"threshold"`
	Observed  float64 `json:"observed"`
	Status    string  `json:"status"`
	Fired     bool    `json:"fired"`
}

// OTelStore records inspired telemetry without an OTel SDK or collector.
type OTelStore struct {
	mu          sync.Mutex
	spans       map[string]*Span
	order       []string
	seq         uint64
	exportCount uint64
	alertCount  uint64
}

// NewOTelStore creates an empty in-process telemetry simulator.
func NewOTelStore() *OTelStore {
	return &OTelStore{
		spans: make(map[string]*Span),
		order: make([]string, 0),
	}
}

// StartSpan starts a named in-process span.
func (s *OTelStore) StartSpan(name string) (Span, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 128 {
		return Span{}, errors.New("invalid span name")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	traceID := fmt.Sprintf("trace-%016x", s.seq)
	span := &Span{
		TraceID:   traceID,
		SpanID:    fmt.Sprintf("span-%016x", s.seq),
		Name:      name,
		StartedAt: time.Now().UTC(),
	}
	s.spans[traceID] = span
	s.order = append(s.order, traceID)
	return *span, nil
}

// EndSpan records completion for a previously started span.
func (s *OTelStore) EndSpan(traceID string) (Span, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	span, ok := s.spans[traceID]
	if !ok {
		return Span{}, errors.New("span not found")
	}
	if span.EndedAt.IsZero() {
		span.EndedAt = time.Now().UTC()
	}
	return *span, nil
}

// ExportTraces returns an isolated, honesty-labeled snapshot.
func (s *OTelStore) ExportTraces() TraceExport {
	s.mu.Lock()
	defer s.mu.Unlock()
	spans := make([]Span, 0, len(s.order))
	for _, id := range s.order {
		spans = append(spans, *s.spans[id])
	}
	s.exportCount++
	return TraceExport{
		InstrumentationModel: "otel-inspired",
		Collector:            "none",
		Spans:                spans,
	}
}

// EvaluateAlerts evaluates one of the compiled-in fixture rules.
func (s *OTelStore) EvaluateAlerts(ruleID string, samples []float64) (AlertResult, error) {
	rules := map[string]float64{"high-error-rate": 0.05}
	threshold, ok := rules[ruleID]
	if !ok || len(samples) == 0 || len(samples) > 1000 {
		return AlertResult{}, errors.New("unknown fixture rule or invalid samples")
	}
	observed := samples[0]
	for _, sample := range samples {
		if sample < 0 {
			return AlertResult{}, errors.New("samples must be non-negative")
		}
		if sample > observed {
			observed = sample
		}
	}
	fired := observed > threshold
	status := "alert_ok"
	if fired {
		status = "alert_fired"
	}
	s.mu.Lock()
	s.alertCount++
	s.mu.Unlock()
	return AlertResult{
		RuleID: ruleID, Threshold: threshold, Observed: observed, Status: status, Fired: fired,
	}, nil
}

// Metrics returns in-process counters exported by the simulator.
func (s *OTelStore) Metrics() map[string]uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]uint64{
		"spans_started_total":     uint64(len(s.spans)),
		"trace_exports_total":     s.exportCount,
		"alert_evaluations_total": s.alertCount,
	}
}

// AlertRuleCount reports the number of fixed, non-executable fixture rules.
func (s *OTelStore) AlertRuleCount() int {
	return 1
}

// Info describes the simulator without claiming a real OTel collector.
func (s *OTelStore) Info() map[string]any {
	return map[string]any{
		"requirement_id":        "ENG-I-005",
		"service":               "eng-i-005",
		"title":                 "OTel-inspired observability simulator",
		"otel_inspired":         true,
		"simulator":             true,
		"instrumentation_model": "otel-inspired",
		"collector":             "none",
		"note":                  "In-process stdlib simulator; does NOT connect to an OTel collector, Tempo, or Grafana",
	}
}
