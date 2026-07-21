package main

import (
	"errors"
	"sync"
)

var (
	ErrInvalidDuration = errors.New("duration_sec must be > 0")
	ErrInvalidCPU      = errors.New("cpu_cores must be > 0")
	ErrInvalidMemory   = errors.New("memory_gb must be > 0")
)

// Pricing constants (demo rates — not production billing).
const (
	usdPerCPUHour    = 0.04
	usdPerMemoryGBHour = 0.01
	cacheHitDiscount = 0.85 // cache hits cost 15% of full build
)

// BuildInput is the untrusted client payload for a recorded build.
// Clients must NOT supply cost_usd or savings_pct (T-5-21).
type BuildInput struct {
	DurationSec float64 `json:"duration_sec"`
	CPUCores    float64 `json:"cpu_cores"`
	MemoryGB    float64 `json:"memory_gb"`
	CacheHit    bool    `json:"cache_hit"`
}

// BuildRecord is a server-computed cost for one build.
type BuildRecord struct {
	CostUSD  float64 `json:"cost_usd"`
	CacheHit bool    `json:"cache_hit"`
}

// CostReport is server-computed from recorded builds (ignore client savings).
type CostReport struct {
	CostPerBuildUSD float64 `json:"cost_per_build_usd"`
	CacheSavingsPct float64 `json:"cache_savings_pct"`
	Builds          int     `json:"builds"`
	CacheHits       int     `json:"cache_hits"`
	CacheMisses     int     `json:"cache_misses"`
	TotalCostUSD    float64 `json:"total_cost_usd"`
}

// CostMeter meters cost-per-build and cache-hit savings.
// Does NOT own CAS digest store (N-010) — hits are meter inputs only.
type CostMeter struct {
	mu      sync.Mutex
	builds  []BuildRecord
	hits    int
	misses  int
	totalUSD float64
}

func NewCostMeter() *CostMeter {
	return &CostMeter{builds: make([]BuildRecord, 0)}
}

func computeCostUSD(in BuildInput) float64 {
	hours := in.DurationSec / 3600.0
	full := hours*in.CPUCores*usdPerCPUHour + hours*in.MemoryGB*usdPerMemoryGBHour
	if in.CacheHit {
		return full * (1.0 - cacheHitDiscount)
	}
	return full
}

func (m *CostMeter) RecordBuild(in BuildInput) (BuildRecord, error) {
	if in.DurationSec <= 0 {
		return BuildRecord{}, ErrInvalidDuration
	}
	if in.CPUCores <= 0 {
		return BuildRecord{}, ErrInvalidCPU
	}
	if in.MemoryGB <= 0 {
		return BuildRecord{}, ErrInvalidMemory
	}

	cost := computeCostUSD(in)
	rec := BuildRecord{CostUSD: cost, CacheHit: in.CacheHit}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.builds = append(m.builds, rec)
	m.totalUSD += cost
	if in.CacheHit {
		m.hits++
	} else {
		m.misses++
	}
	return rec, nil
}

func (m *CostMeter) Report() CostReport {
	m.mu.Lock()
	defer m.mu.Unlock()

	n := len(m.builds)
	report := CostReport{
		Builds:      n,
		CacheHits:   m.hits,
		CacheMisses: m.misses,
		TotalCostUSD: m.totalUSD,
	}
	if n == 0 {
		return report
	}
	report.CostPerBuildUSD = m.totalUSD / float64(n)

	// Savings: compare observed total vs hypothetical all-miss cost.
	var fullMissTotal float64
	for _, b := range m.builds {
		if b.CacheHit {
			// Invert discount to estimate full miss cost for this build.
			if (1.0 - cacheHitDiscount) > 0 {
				fullMissTotal += b.CostUSD / (1.0 - cacheHitDiscount)
			}
		} else {
			fullMissTotal += b.CostUSD
		}
	}
	if fullMissTotal > 0 {
		saved := fullMissTotal - m.totalUSD
		report.CacheSavingsPct = (saved / fullMissTotal) * 100.0
	}
	return report
}

// SeedDemo records miss + hit builds and returns a live cost report.
func (m *CostMeter) SeedDemo() (CostReport, error) {
	if _, err := m.RecordBuild(BuildInput{DurationSec: 300, CPUCores: 4, MemoryGB: 8, CacheHit: false}); err != nil {
		return CostReport{}, err
	}
	if _, err := m.RecordBuild(BuildInput{DurationSec: 30, CPUCores: 4, MemoryGB: 8, CacheHit: true}); err != nil {
		return CostReport{}, err
	}
	if _, err := m.RecordBuild(BuildInput{DurationSec: 45, CPUCores: 2, MemoryGB: 4, CacheHit: true}); err != nil {
		return CostReport{}, err
	}
	return m.Report(), nil
}
