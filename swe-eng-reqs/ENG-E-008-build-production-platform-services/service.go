package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// StructuredError is a client-safe error payload without stack traces or secrets.
type StructuredError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ServiceRuntime owns production runtime counters and request-ID helpers.
type ServiceRuntime struct {
	mu      sync.Mutex
	echoes  int
	demos   int
	metrics int
}

// NewServiceRuntime returns an empty production service runtime.
func NewServiceRuntime() *ServiceRuntime {
	return &ServiceRuntime{}
}

// WithRequestID returns the client-provided request ID or generates one.
func (r *ServiceRuntime) WithRequestID(incoming string) string {
	if incoming != "" {
		return incoming
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-fallback"
	}
	return "req-" + hex.EncodeToString(b[:])
}

// IncrementEcho records an echo call against the metrics counter.
func (r *ServiceRuntime) IncrementEcho() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.echoes++
	r.metrics++
}

// IncrementDemo records a demo call against the metrics counter.
func (r *ServiceRuntime) IncrementDemo() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.demos++
	r.metrics++
}

// MetricsCount returns the combined echo/demo metrics counter.
func (r *ServiceRuntime) MetricsCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.metrics
}

// EchoCount returns echo invocations.
func (r *ServiceRuntime) EchoCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.echoes
}

// DemoCount returns demo invocations.
func (r *ServiceRuntime) DemoCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.demos
}

// SampleStructuredError returns a non-leaking structured error for proofs.
func SampleStructuredError() StructuredError {
	return StructuredError{Code: "demo_error", Message: "structured error without internals"}
}
