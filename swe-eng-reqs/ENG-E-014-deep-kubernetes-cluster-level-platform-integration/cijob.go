package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCIJob = errors.New("invalid CIJob id or image")
	ErrCIJobMissing = errors.New("CIJob not found")
	ErrInvalidJob   = errors.New("invalid Job outcome")
)

// Condition mirrors Kubernetes status.conditions for CIJob.
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// Job is the child workload scheduled by a CIJob (in-memory analogue).
type Job struct {
	Name   string `json:"name"`
	Status string `json:"status"` // Pending | Succeeded | Failed
}

// CIJob is the in-memory analogue of the CIJob CRD (not ManagedWorkload).
type CIJob struct {
	ID           string      `json:"id"`
	Kind         string      `json:"kind"`
	Image        string      `json:"image"`
	Job          *Job        `json:"job,omitempty"`
	JobScheduled bool        `json:"job_scheduled"`
	Conditions   []Condition `json:"conditions"`
}

// CIJobController schedules Jobs and derives Complete/Failed from Job state.
type CIJobController struct {
	mu    sync.RWMutex
	jobs  map[string]*CIJob
	seq   int
}

// NewCIJobController returns an empty concurrency-safe controller.
func NewCIJobController() *CIJobController {
	return &CIJobController{jobs: make(map[string]*CIJob)}
}

// Create registers a CIJob and schedules a Job child (job_scheduled=true).
func (c *CIJobController) Create(id, image string) (CIJob, error) {
	if !safeCIJobID(id) || strings.TrimSpace(image) == "" {
		return CIJob{}, ErrInvalidCIJob
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.jobs[id]; exists {
		return CIJob{}, errors.New("CIJob already exists")
	}
	c.seq++
	now := time.Now().UTC()
	job := &CIJob{
		ID:           id,
		Kind:         "CIJob",
		Image:        image,
		JobScheduled: true,
		Job: &Job{
			Name:   fmt.Sprintf("%s-job-%d", id, c.seq),
			Status: "Pending",
		},
	}
	setCondition(job, "Ready", "True", "JobScheduled", "Job child scheduled", now)
	setCondition(job, "Complete", "False", "Pending", "Job not finished", now)
	setCondition(job, "Failed", "False", "Pending", "Job not failed", now)
	c.jobs[id] = job
	return copyCIJob(job), nil
}

// Get returns an isolated copy of a CIJob.
func (c *CIJobController) Get(id string) (CIJob, bool) {
	if !safeCIJobID(id) {
		return CIJob{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	job, ok := c.jobs[id]
	if !ok {
		return CIJob{}, false
	}
	return copyCIJob(job), true
}

// SetJobOutcome updates the child Job status (server-side Job observation).
func (c *CIJobController) SetJobOutcome(id, outcome string) error {
	if outcome != "Succeeded" && outcome != "Failed" {
		return ErrInvalidJob
	}
	if !safeCIJobID(id) {
		return ErrInvalidCIJob
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	job, ok := c.jobs[id]
	if !ok {
		return ErrCIJobMissing
	}
	if job.Job == nil {
		return ErrInvalidJob
	}
	job.Job.Status = outcome
	return nil
}

// Reconcile derives Complete/Failed conditions from Job state (T-5-16).
func (c *CIJobController) Reconcile(id string) (CIJob, error) {
	if !safeCIJobID(id) {
		return CIJob{}, ErrInvalidCIJob
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	job, ok := c.jobs[id]
	if !ok {
		return CIJob{}, ErrCIJobMissing
	}
	reconcileCIJob(job, time.Now().UTC())
	return copyCIJob(job), nil
}

// ReconcileAll reconciles every CIJob; returns count with terminal conditions.
func (c *CIJobController) ReconcileAll() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	terminal := 0
	for _, job := range c.jobs {
		reconcileCIJob(job, now)
		if conditionTruePtr(job, "Complete") || conditionTruePtr(job, "Failed") {
			terminal++
		}
	}
	return terminal
}

func reconcileCIJob(job *CIJob, now time.Time) {
	if job.Job == nil {
		setCondition(job, "Ready", "False", "NoJob", "no Job child", now)
		return
	}
	switch job.Job.Status {
	case "Succeeded":
		// Server-computed: Complete from Job state; never both True.
		setCondition(job, "Complete", "True", "JobSucceeded", "Job completed successfully", now)
		setCondition(job, "Failed", "False", "JobSucceeded", "Job did not fail", now)
		setCondition(job, "Ready", "False", "Terminal", "CIJob finished", now)
	case "Failed":
		setCondition(job, "Failed", "True", "JobFailed", "Job failed", now)
		setCondition(job, "Complete", "False", "JobFailed", "Job did not complete", now)
		setCondition(job, "Ready", "False", "Terminal", "CIJob finished", now)
	default:
		setCondition(job, "Ready", "True", "JobScheduled", "Job child scheduled", now)
		setCondition(job, "Complete", "False", "Pending", "Job not finished", now)
		setCondition(job, "Failed", "False", "Pending", "Job not failed", now)
	}
}

func setCondition(job *CIJob, kind, status, reason, message string, now time.Time) {
	for i := range job.Conditions {
		if job.Conditions[i].Type != kind {
			continue
		}
		if job.Conditions[i].Status != status {
			job.Conditions[i].LastTransitionTime = now.UTC()
		}
		job.Conditions[i].Status = status
		job.Conditions[i].Reason = reason
		job.Conditions[i].Message = message
		return
	}
	job.Conditions = append(job.Conditions, Condition{
		Type: kind, Status: status, Reason: reason, Message: message, LastTransitionTime: now.UTC(),
	})
}

func conditionTrue(job CIJob, kind string) bool {
	for _, condition := range job.Conditions {
		if condition.Type == kind {
			return condition.Status == "True"
		}
	}
	return false
}

func conditionTruePtr(job *CIJob, kind string) bool {
	return conditionTrue(*job, kind)
}

func copyCIJob(job *CIJob) CIJob {
	out := *job
	out.Conditions = append([]Condition(nil), job.Conditions...)
	if job.Job != nil {
		j := *job.Job
		out.Job = &j
	}
	return out
}

func safeCIJobID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}
