package main

import (
	"errors"
	"fmt"
	"sync"
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

type PipelineService struct {
	pipelines map[string]*Pipeline
	mu        sync.Mutex
	counter   int
}

func NewPipelineService() *PipelineService {
	return &PipelineService{
		pipelines: make(map[string]*Pipeline),
	}
}

func (ps *PipelineService) SubmitPipeline(dag map[string][]string) (*Pipeline, error) {
	pipeline := &Pipeline{
		Stages: make(map[string]*Stage),
	}

	for name, deps := range dag {
		pipeline.Stages[name] = &Stage{
			Name:         name,
			Status:       "pending",
			Dependencies: deps,
			Retries:      0,
		}
	}

	if err := pipeline.Validate(); err != nil {
		return nil, err
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.counter++
	pipeline.ID = fmt.Sprintf("pipe-%d", ps.counter)
	ps.pipelines[pipeline.ID] = pipeline

	return pipeline, nil
}

func (ps *PipelineService) GetPipeline(id string) (*Pipeline, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	pipeline, ok := ps.pipelines[id]
	if !ok {
		return nil, errors.New("pipeline not found")
	}

	return pipeline, nil
}

func (ps *PipelineService) TransitionStage(pipelineID, stageName, newStatus string) error {
	ps.mu.Lock()
	pipeline, ok := ps.pipelines[pipelineID]
	ps.mu.Unlock()

	if !ok {
		return errors.New("pipeline not found")
	}

	return pipeline.TransitionStage(stageName, newStatus)
}

func (ps *PipelineService) ListPipelines() []*Pipeline {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	result := make([]*Pipeline, 0, len(ps.pipelines))
	for _, p := range ps.pipelines {
		result = append(result, p)
	}
	return result
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
		for range stage.Dependencies {
			inDegree[stage.Name]++
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
