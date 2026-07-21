package main

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrMissingTenantID = errors.New("tenant id required")
	ErrUnsafeTenantID  = errors.New("unsafe tenant id")
	ErrQuotaExceeded   = errors.New("tenant quota exceeded")
	ErrRateLimited     = errors.New("noisy-neighbor rate limit exceeded")
	ErrInvalidQuota    = errors.New("quota must be positive")
	ErrInvalidUnits    = errors.New("units must be positive")
	ErrUnknownTenant   = errors.New("tenant quota not configured")
)

// ScheduleResult is one accepted schedule grant.
type ScheduleResult struct {
	TenantID string `json:"tenant_id"`
	Units    int    `json:"units"`
	Used     int    `json:"used"`
	Quota    int    `json:"quota"`
}

// TenantScheduler enforces per-tenant quotas and noisy-neighbor rate limits.
type TenantScheduler struct {
	mu            sync.Mutex
	quotas        map[string]int
	used          map[string]int
	rateLimits    map[string]int // max schedule calls (units) before limit
	rateUsed      map[string]int
	quotaDenied   int
	rateDenied    int
	scheduledOK   int
}

func NewTenantScheduler() *TenantScheduler {
	return &TenantScheduler{
		quotas:     make(map[string]int),
		used:       make(map[string]int),
		rateLimits: make(map[string]int),
		rateUsed:   make(map[string]int),
	}
}

func validateTenantID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrMissingTenantID
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return ErrUnsafeTenantID
	}
	return nil
}

func (s *TenantScheduler) SetQuota(tenantID string, quota int) error {
	if err := validateTenantID(tenantID); err != nil {
		return err
	}
	if quota <= 0 {
		return ErrInvalidQuota
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quotas[tenantID] = quota
	return nil
}

func (s *TenantScheduler) SetRateLimit(tenantID string, limit int) error {
	if err := validateTenantID(tenantID); err != nil {
		return err
	}
	if limit <= 0 {
		return ErrInvalidQuota
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rateLimits[tenantID] = limit
	return nil
}

func (s *TenantScheduler) Schedule(tenantID string, units int) (*ScheduleResult, error) {
	if err := validateTenantID(tenantID); err != nil {
		return nil, err
	}
	if units <= 0 {
		return nil, ErrInvalidUnits
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, ok := s.quotas[tenantID]
	if !ok {
		return nil, ErrUnknownTenant
	}
	if limit, limited := s.rateLimits[tenantID]; limited {
		if s.rateUsed[tenantID]+units > limit {
			s.rateDenied++
			return nil, ErrRateLimited
		}
	}
	if s.used[tenantID]+units > quota {
		s.quotaDenied++
		return nil, ErrQuotaExceeded
	}
	s.used[tenantID] += units
	s.rateUsed[tenantID] += units
	s.scheduledOK++
	out := &ScheduleResult{
		TenantID: tenantID,
		Units:    units,
		Used:     s.used[tenantID],
		Quota:    quota,
	}
	return out, nil
}

func (s *TenantScheduler) TenantCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.quotas)
}

func (s *TenantScheduler) Proof() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"tenants":                 len(s.quotas),
		"quota_enforced":          s.quotaDenied > 0,
		"noisy_neighbor_limited":  s.rateDenied > 0,
		"scheduled_ok":            s.scheduledOK,
		"quota_denials":           s.quotaDenied,
		"rate_denials":            s.rateDenied,
	}
}

// SeedDemo configures two tenants and exercises quota + noisy-neighbor limits.
func (s *TenantScheduler) SeedDemo() error {
	if err := s.SetQuota("tenant-a", 2); err != nil {
		return err
	}
	if err := s.SetQuota("tenant-b", 10); err != nil {
		return err
	}
	if err := s.SetRateLimit("tenant-a", 1); err != nil {
		return err
	}
	// tenant-a: first schedule OK; second hits rate limit (noisy)
	if _, err := s.Schedule("tenant-a", 1); err != nil {
		return err
	}
	if _, err := s.Schedule("tenant-a", 1); err == nil {
		return errors.New("expected noisy-neighbor deny for tenant-a")
	} else if !errors.Is(err, ErrRateLimited) {
		return err
	}
	// Fill remaining quota path: raise rate so quota denial can fire
	s.mu.Lock()
	s.rateLimits["tenant-a"] = 100
	s.mu.Unlock()
	if _, err := s.Schedule("tenant-a", 1); err != nil {
		return err
	}
	if _, err := s.Schedule("tenant-a", 1); err == nil {
		return errors.New("expected quota deny for tenant-a")
	} else if !errors.Is(err, ErrQuotaExceeded) {
		return err
	}
	// Quiet tenant still works
	if _, err := s.Schedule("tenant-b", 1); err != nil {
		return err
	}
	return nil
}
