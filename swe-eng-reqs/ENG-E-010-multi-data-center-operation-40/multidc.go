package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const maxDataCenters = 256

var (
	ErrUnsafeDCID     = errors.New("unsafe dc id")
	ErrUnsafeDomain   = errors.New("unsafe failure domain")
	ErrDCCapExceeded  = errors.New("data center cap exceeded")
	ErrDuplicateDC    = errors.New("duplicate data center id")
	ErrNoDataCenters  = errors.New("no data centers registered")
	ErrEmptyFanoutCfg = errors.New("empty fanout config")
)

// DataCenter is one simulated site in the multi-DC control plane.
type DataCenter struct {
	ID      string `json:"id"`
	Domain  string `json:"failure_domain"`
	Healthy bool   `json:"healthy"`
	Config  map[string]any `json:"config,omitempty"`
}

// FanoutResult reports partial success across healthy vs unhealthy DCs.
type FanoutResult struct {
	FanoutOK bool `json:"fanout_ok"`
	Pushed   int  `json:"pushed"`
	Failed   int  `json:"failed"`
	Skipped  int  `json:"skipped"`
}

// MultiDC simulates ≥40 DC topology + failure domains + config fan-out.
// This is an in-process simulator — not physical multi-DC infrastructure.
type MultiDC struct {
	mu  sync.Mutex
	dcs map[string]*DataCenter
}

func NewMultiDC() *MultiDC {
	return &MultiDC{dcs: make(map[string]*DataCenter)}
}

func safeName(s string) bool {
	return s != "" && !strings.Contains(s, "..") && !strings.ContainsAny(s, `/\`)
}

func (m *MultiDC) RegisterDC(id, domain string, healthy bool) error {
	if !safeName(id) {
		return ErrUnsafeDCID
	}
	if !safeName(domain) {
		return ErrUnsafeDomain
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.dcs[id]; exists {
		return ErrDuplicateDC
	}
	if len(m.dcs) >= maxDataCenters {
		return ErrDCCapExceeded
	}
	m.dcs[id] = &DataCenter{ID: id, Domain: domain, Healthy: healthy}
	return nil
}

func (m *MultiDC) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.dcs)
}

func (m *MultiDC) Domains() map[string][]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string][]string)
	for _, dc := range m.dcs {
		out[dc.Domain] = append(out[dc.Domain], dc.ID)
	}
	for domain := range out {
		sort.Strings(out[domain])
	}
	return out
}

func (m *MultiDC) List() []DataCenter {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]DataCenter, 0, len(m.dcs))
	ids := make([]string, 0, len(m.dcs))
	for id := range m.dcs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		dc := *m.dcs[id]
		if dc.Config != nil {
			cfg := make(map[string]any, len(dc.Config))
			for k, v := range dc.Config {
				cfg[k] = v
			}
			dc.Config = cfg
		}
		out = append(out, dc)
	}
	return out
}

func (m *MultiDC) Fanout(cfg map[string]any) (FanoutResult, error) {
	if len(cfg) == 0 {
		return FanoutResult{}, ErrEmptyFanoutCfg
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.dcs) == 0 {
		return FanoutResult{}, ErrNoDataCenters
	}
	copied := make(map[string]any, len(cfg))
	for k, v := range cfg {
		copied[k] = v
	}
	var result FanoutResult
	for _, dc := range m.dcs {
		if !dc.Healthy {
			result.Failed++
			continue
		}
		dc.Config = copied
		result.Pushed++
	}
	result.FanoutOK = result.Pushed > 0
	return result, nil
}

func (m *MultiDC) Info() map[string]any {
	return map[string]any{
		"simulator":           true,
		"multi_dc_simulator":  true,
		"physical_dcs":        false,
		"max_data_centers":    maxDataCenters,
		"note":                "In-process multi-DC simulator; does NOT operate real physical data centers",
	}
}

// SeedDemo registers 42 DCs across 6 failure domains for live demos.
func (m *MultiDC) SeedDemo() error {
	for i := 0; i < 42; i++ {
		domain := fmt.Sprintf("fd-%d", i%6)
		healthy := i%11 != 0 // a few unhealthy for partial fan-out proof
		if err := m.RegisterDC(fmt.Sprintf("dc-%02d", i), domain, healthy); err != nil {
			return err
		}
	}
	return nil
}
