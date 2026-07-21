package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Stage struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
	Retries      int      `json:"retries"`
}

type Pipeline struct {
	ID     string            `json:"id"`
	Stages map[string]*Stage `json:"stages"`
	mu     sync.Mutex
}

func NewPipeline(dag map[string][]string) *Pipeline {
	p := &Pipeline{
		Stages: make(map[string]*Stage),
	}
	for name, deps := range dag {
		p.Stages[name] = &Stage{
			Name:         name,
			Status:       "pending",
			Dependencies: deps,
			Retries:      0,
		}
	}
	return p
}

func (p *Pipeline) Validate() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.Stages) == 0 {
		return errors.New("empty DAG")
	}

	for _, stage := range p.Stages {
		for _, dep := range stage.Dependencies {
			if _, ok := p.Stages[dep]; !ok {
				return fmt.Errorf("stage not found: %s (dependency of %s)", dep, stage.Name)
			}
		}
	}

	inDegree := make(map[string]int)
	for name := range p.Stages {
		inDegree[name] = 0
	}
	for _, stage := range p.Stages {
		for _, dep := range stage.Dependencies {
			inDegree[stage.Name]++
			_ = dep
		}
	}

	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visited++

		for _, stage := range p.Stages {
			for _, dep := range stage.Dependencies {
				if dep == current {
					inDegree[stage.Name]--
					if inDegree[stage.Name] == 0 {
						queue = append(queue, stage.Name)
					}
				}
			}
		}
	}

	if visited != len(p.Stages) {
		return errors.New("cycle detected in DAG")
	}

	return nil
}

var validTransitions = map[string]map[string]bool{
	"pending": {
		"running": true,
	},
	"running": {
		"succeeded": true,
		"failed":    true,
	},
	"failed": {
		"running": true,
	},
	"succeeded": {},
}

func (p *Pipeline) TransitionStage(name, newStatus string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	stage, ok := p.Stages[name]
	if !ok {
		return fmt.Errorf("stage not found: %s", name)
	}

	currentStatus := stage.Status
	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid transition: unknown status %q", currentStatus)
	}

	if !allowed[newStatus] {
		return fmt.Errorf("invalid transition: %s → %s", currentStatus, newStatus)
	}

	stage.Status = newStatus
	if newStatus == "running" && currentStatus == "failed" {
		stage.Retries++
	}

	return nil
}

type RetryPolicy struct {
	BaseDelay  time.Duration
	MaxRetries int
	Jitter     float64
}

func NewRetryPolicy(baseDelay time.Duration, maxRetries int, jitter float64) *RetryPolicy {
	return &RetryPolicy{
		BaseDelay:  baseDelay,
		MaxRetries: maxRetries,
		Jitter:     jitter,
	}
}

func (rp *RetryPolicy) NextBackoff(attempt int) time.Duration {
	base := rp.BaseDelay * time.Duration(1<<(attempt-1))

	if rp.Jitter > 0 {
		jitterRange := float64(base) * rp.Jitter
		jitterAmount := (rand.Float64()*2 - 1) * jitterRange
		base = time.Duration(float64(base) + jitterAmount)
	}

	return base
}

func (rp *RetryPolicy) ShouldRetry(attempt int) error {
	if attempt > rp.MaxRetries {
		return errors.New("max retries exceeded")
	}
	return nil
}
