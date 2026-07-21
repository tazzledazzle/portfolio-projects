package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestCost_RecordBuild_ComputesUSD(t *testing.T) {
	m := NewCostMeter()
	rec, err := m.RecordBuild(BuildInput{
		DurationSec: 120,
		CPUCores:    4,
		MemoryGB:    8,
		CacheHit:    false,
	})
	if err != nil {
		t.Fatalf("RecordBuild: %v", err)
	}
	if rec.CostUSD <= 0 {
		t.Fatalf("expected cost_per_build_usd > 0, got %v", rec.CostUSD)
	}
}

func TestCost_CacheHit_Savings(t *testing.T) {
	baseline := NewCostMeter()
	_, _ = baseline.RecordBuild(BuildInput{DurationSec: 100, CPUCores: 2, MemoryGB: 4, CacheHit: false})
	_, _ = baseline.RecordBuild(BuildInput{DurationSec: 100, CPUCores: 2, MemoryGB: 4, CacheHit: false})
	baseReport := baseline.Report()

	withHits := NewCostMeter()
	_, _ = withHits.RecordBuild(BuildInput{DurationSec: 100, CPUCores: 2, MemoryGB: 4, CacheHit: false})
	_, _ = withHits.RecordBuild(BuildInput{DurationSec: 10, CPUCores: 2, MemoryGB: 4, CacheHit: true})
	hitReport := withHits.Report()

	if hitReport.CacheSavingsPct <= baseReport.CacheSavingsPct {
		t.Fatalf("expected cache hit to increase cache_savings_pct: baseline=%.2f with_hits=%.2f",
			baseReport.CacheSavingsPct, hitReport.CacheSavingsPct)
	}
	if hitReport.CacheSavingsPct <= 0 {
		t.Fatalf("expected positive cache_savings_pct, got %.2f", hitReport.CacheSavingsPct)
	}
}

func TestCost_Report_Fields(t *testing.T) {
	m := NewCostMeter()
	_, _ = m.RecordBuild(BuildInput{DurationSec: 60, CPUCores: 2, MemoryGB: 4, CacheHit: true})
	report := m.Report()
	if report.CostPerBuildUSD <= 0 {
		t.Fatalf("expected cost_per_build_usd > 0, got %v", report.CostPerBuildUSD)
	}
	// cache_savings_pct must be present (may be 0–100)
	if report.CacheSavingsPct < 0 || report.CacheSavingsPct > 100 {
		t.Fatalf("cache_savings_pct out of range: %v", report.CacheSavingsPct)
	}
}

func TestCost_NotCASCache(t *testing.T) {
	m := NewCostMeter()
	rt := reflect.TypeOf(m).Elem()
	for i := 0; i < rt.NumField(); i++ {
		name := strings.ToLower(rt.Field(i).Name)
		if strings.Contains(name, "blob") || strings.Contains(name, "digest") || strings.Contains(name, "cas") {
			t.Fatalf("CostMeter must not own CAS cache fields (N-010); found %s", rt.Field(i).Name)
		}
	}
	// No Put/Get digest API on CostMeter
	methods := map[string]bool{}
	mt := reflect.TypeOf(m)
	for i := 0; i < mt.NumMethod(); i++ {
		methods[mt.Method(i).Name] = true
	}
	for _, banned := range []string{"Put", "Get", "HitRate"} {
		if methods[banned] {
			t.Fatalf("CostMeter must not implement CAS method %s (N-010 ownership)", banned)
		}
	}
}
