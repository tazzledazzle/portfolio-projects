package main

import (
	"fmt"
	"strings"
	"sync"
)

// Agent is a Buildkite agent simulator.
type Agent struct {
	ID     string
	Status string
	Tags   []string
}

// BKJob is a Buildkite job in the agent queue.
type BKJob struct {
	ID         string
	Pipeline   string
	Stage      string
	Status     string
	ExitCode   int
	AgentID    string
	Group      string
	offeredTo  string
}

// AgentPool manages agents and FIFO job queue.
type AgentPool struct {
	mu     sync.Mutex
	agents map[string]*Agent
	jobs   map[string]*BKJob
	queue  []*BKJob
	seq    int
}

func NewAgentPool() *AgentPool {
	return &AgentPool{
		agents: map[string]*Agent{},
		jobs:   map[string]*BKJob{},
	}
}

func (p *AgentPool) Register(id string, tags []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if id == "" {
		return fmt.Errorf("agent id required")
	}
	if _, ok := p.agents[id]; ok {
		return fmt.Errorf("agent already registered")
	}
	p.agents[id] = &Agent{ID: id, Status: "idle", Tags: tags}
	return nil
}

func (p *AgentPool) Enqueue(job *BKJob) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if job.Status == "" {
		job.Status = "waiting"
	}
	p.jobs[job.ID] = job
	p.queue = append(p.queue, job)
}

func (p *AgentPool) Poll(agentID string) *BKJob {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.agents[agentID]; !ok {
		return nil
	}
	for _, job := range p.queue {
		if job.Status == "waiting" && job.offeredTo == "" {
			job.offeredTo = agentID
			return job
		}
	}
	return nil
}

func (p *AgentPool) Claim(agentID, jobID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	job, ok := p.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found")
	}
	if job.Status == "running" {
		return fmt.Errorf("job already claimed")
	}
	if job.offeredTo != "" && job.offeredTo != agentID {
		return fmt.Errorf("job already claimed")
	}
	job.Status = "running"
	job.AgentID = agentID
	if a, ok := p.agents[agentID]; ok {
		a.Status = "busy"
	}
	// remove from waiting queue
	nq := p.queue[:0]
	for _, j := range p.queue {
		if j.ID != jobID {
			nq = append(nq, j)
		}
	}
	p.queue = nq
	return nil
}

func (p *AgentPool) Complete(jobID string, exitCode int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	job, ok := p.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found")
	}
	job.ExitCode = exitCode
	if exitCode == 0 {
		job.Status = "succeeded"
	} else {
		job.Status = "failed"
	}
	if a, ok := p.agents[job.AgentID]; ok {
		a.Status = "idle"
	}
	return nil
}

func (p *AgentPool) UploadPipeline(pipelineID, yaml string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	stages := parseSimpleStages(yaml)
	if len(stages) == 0 {
		return fmt.Errorf("no stages found in pipeline yaml")
	}
	for _, stage := range stages {
		p.seq++
		id := fmt.Sprintf("%s-%d", pipelineID, p.seq)
		job := &BKJob{
			ID:       id,
			Pipeline: pipelineID,
			Stage:    stage,
			Status:   "waiting",
		}
		p.jobs[id] = job
		p.queue = append(p.queue, job)
	}
	return nil
}

func (p *AgentPool) AgentCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.agents)
}

func parseSimpleStages(yaml string) []string {
	var stages []string
	for _, line := range strings.Split(yaml, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- label:") {
			label := strings.TrimSpace(strings.TrimPrefix(line, "- label:"))
			label = strings.Trim(label, `"'`)
			if label != "" {
				stages = append(stages, label)
			}
		}
	}
	return stages
}
