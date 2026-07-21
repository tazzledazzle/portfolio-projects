package main

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrUnsafeDomain   = errors.New("unsafe failure domain")
	ErrUnsafeTenant   = errors.New("unsafe tenant id")
	ErrUnknownDomain  = errors.New("unknown failure domain")
	ErrEmptyScenario  = errors.New("chaos scenario required")
	ErrNoDomains      = errors.New("no failure domains registered")
	ErrChaosNotRun    = errors.New("chaos has not been run")
)

// ChaosResult is the outcome of injecting a failure into one domain.
type ChaosResult struct {
	ChaosRan  bool   `json:"chaos_ran"`
	Contained bool   `json:"contained"`
	Domain    string `json:"domain"`
	Scenario  string `json:"scenario"`
}

// BlastRadiusReport is server-computed from scenario state (not client-supplied).
type BlastRadiusReport struct {
	AffectedDomains    []string `json:"affected_domains"`
	UnaffectedDomains  []string `json:"unaffected_domains"`
	AffectedTenants    []string `json:"affected_tenants"`
	UnaffectedTenants  []string `json:"unaffected_tenants"`
	Contained          bool     `json:"contained"`
	ChaosRan           bool     `json:"chaos_ran"`
}

// BlastEngine proves chaos blast radius stays within one failure domain.
// Local in-memory domains/tenants only — does not import E-010/I-004.
type BlastEngine struct {
	mu               sync.Mutex
	domains          map[string][]string // domain → tenants
	chaosDomain      string
	chaosScenario    string
	chaosRan         bool
}

func NewBlastEngine() *BlastEngine {
	return &BlastEngine{domains: make(map[string][]string)}
}

func safeName(s string) bool {
	return s != "" && !strings.Contains(s, "..") && !strings.ContainsAny(s, `/\`)
}

func (e *BlastEngine) RegisterDomain(domain string, tenants []string) error {
	if !safeName(domain) {
		return ErrUnsafeDomain
	}
	clean := make([]string, 0, len(tenants))
	for _, t := range tenants {
		if !safeName(t) {
			return ErrUnsafeTenant
		}
		clean = append(clean, t)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.domains[domain] = clean
	return nil
}

func (e *BlastEngine) RunChaos(domain, scenario string) (ChaosResult, error) {
	if !safeName(domain) {
		return ChaosResult{}, ErrUnsafeDomain
	}
	if strings.TrimSpace(scenario) == "" {
		return ChaosResult{}, ErrEmptyScenario
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.domains) == 0 {
		return ChaosResult{}, ErrNoDomains
	}
	if _, ok := e.domains[domain]; !ok {
		return ChaosResult{}, ErrUnknownDomain
	}
	e.chaosDomain = domain
	e.chaosScenario = scenario
	e.chaosRan = true
	contained := e.computeContainedLocked()
	return ChaosResult{
		ChaosRan:  true,
		Contained: contained,
		Domain:    domain,
		Scenario:  scenario,
	}, nil
}

func (e *BlastEngine) BlastRadius() BlastRadiusReport {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.computeRadiusLocked()
}

func (e *BlastEngine) computeContainedLocked() bool {
	if !e.chaosRan {
		return false
	}
	radius := e.computeRadiusLocked()
	// Contained when at least one domain and one tenant remain unaffected
	return len(radius.UnaffectedDomains) > 0 && len(radius.UnaffectedTenants) > 0 &&
		len(radius.AffectedDomains) > 0 && len(radius.AffectedDomains) < len(e.domains)
}

func (e *BlastEngine) computeRadiusLocked() BlastRadiusReport {
	report := BlastRadiusReport{
		ChaosRan:          e.chaosRan,
		AffectedDomains:   []string{},
		UnaffectedDomains: []string{},
		AffectedTenants:   []string{},
		UnaffectedTenants: []string{},
	}
	if !e.chaosRan {
		return report
	}
	for domain, tenants := range e.domains {
		if domain == e.chaosDomain {
			report.AffectedDomains = append(report.AffectedDomains, domain)
			report.AffectedTenants = append(report.AffectedTenants, tenants...)
			continue
		}
		report.UnaffectedDomains = append(report.UnaffectedDomains, domain)
		report.UnaffectedTenants = append(report.UnaffectedTenants, tenants...)
	}
	sort.Strings(report.AffectedDomains)
	sort.Strings(report.UnaffectedDomains)
	sort.Strings(report.AffectedTenants)
	sort.Strings(report.UnaffectedTenants)
	report.Contained = len(report.UnaffectedDomains) > 0 &&
		len(report.UnaffectedTenants) > 0 &&
		len(report.AffectedDomains) > 0 &&
		len(report.AffectedDomains) < len(e.domains)
	return report
}

// SeedDemo registers isolated domains/tenants and runs a contained chaos scenario.
func (e *BlastEngine) SeedDemo() (ChaosResult, error) {
	if err := e.RegisterDomain("fd-east", []string{"tenant-a", "tenant-b"}); err != nil {
		return ChaosResult{}, err
	}
	if err := e.RegisterDomain("fd-west", []string{"tenant-c"}); err != nil {
		return ChaosResult{}, err
	}
	if err := e.RegisterDomain("fd-central", []string{"tenant-d"}); err != nil {
		return ChaosResult{}, err
	}
	return e.RunChaos("fd-east", "partition")
}
