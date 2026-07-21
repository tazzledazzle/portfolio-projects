package main

import (
	"fmt"
	"sync"
)

// Job is a queued Actions workflow job (simulator).
type Job struct {
	ID     string
	Status string
	Matrix map[string]string
	Runner string
}

// Runner is a registered self-hosted runner simulator.
type Runner struct {
	ID         string
	Status     string
	CurrentJob string
}

// RunnerPool manages runners and a job queue.
type RunnerPool struct {
	mu       sync.Mutex
	runners  map[string]*Runner
	jobs     map[string]*Job
	queue    []*Job
	capacity int
}

func NewRunnerPool(capacity int) *RunnerPool {
	if capacity <= 0 {
		capacity = 64
	}
	return &RunnerPool{
		runners:  map[string]*Runner{},
		jobs:     map[string]*Job{},
		queue:    make([]*Job, 0, capacity),
		capacity: capacity,
	}
}

func (p *RunnerPool) Register(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if id == "" {
		return fmt.Errorf("runner id required")
	}
	if _, ok := p.runners[id]; ok {
		return fmt.Errorf("runner already registered")
	}
	p.runners[id] = &Runner{ID: id, Status: "idle"}
	return nil
}

func (p *RunnerPool) Enqueue(job *Job) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if job.Status == "" {
		job.Status = "queued"
	}
	p.jobs[job.ID] = job
	p.queue = append(p.queue, job)
}

func (p *RunnerPool) ClaimJob(runnerID string) *Job {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.runners[runnerID]
	if !ok {
		return nil
	}
	if len(p.queue) == 0 {
		return nil
	}
	job := p.queue[0]
	p.queue = p.queue[1:]
	job.Status = "in_progress"
	job.Runner = runnerID
	r.Status = "busy"
	r.CurrentJob = job.ID
	return job
}

func (p *RunnerPool) ReportComplete(jobID, status string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	job, ok := p.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found")
	}
	if status != "success" && status != "failure" {
		return fmt.Errorf("invalid status")
	}
	job.Status = status
	if r, ok := p.runners[job.Runner]; ok {
		r.Status = "idle"
		r.CurrentJob = ""
	}
	return nil
}

func (p *RunnerPool) RunnerCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.runners)
}

func (p *RunnerPool) CompletedCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, j := range p.jobs {
		if j.Status == "success" || j.Status == "failure" {
			n++
		}
	}
	return n
}

// InjectSecrets mutates env with real secret values and returns a log-safe copy.
func InjectSecrets(env map[string]string, secrets map[string]string) map[string]string {
	masked := make(map[string]string, len(env)+len(secrets))
	for k, v := range env {
		masked[k] = v
	}
	for k, v := range secrets {
		env[k] = v
		masked[k] = "***"
	}
	return masked
}
