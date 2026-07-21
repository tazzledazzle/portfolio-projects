package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

const workloadFinalizer = "managedworkload.devex.coreweave.example/cleanup"

var (
	ErrInvalidWorkload = errors.New("invalid ManagedWorkload")
	ErrWorkloadMissing = errors.New("ManagedWorkload not found")
)

// Condition describes the observed state of a ManagedWorkload.
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// ManagedWorkload is the in-memory analogue of the ManagedWorkload CRD.
type ManagedWorkload struct {
	ID         string      `json:"id"`
	Replicas   int         `json:"replicas"`
	Image      string      `json:"image"`
	Finalizers []string    `json:"finalizers"`
	Deleting   bool        `json:"deleting"`
	Conditions []Condition `json:"conditions"`
}

// Controller reconciles ManagedWorkloads and enforces finalizer deletion.
type Controller struct {
	mu        sync.RWMutex
	workloads map[string]*ManagedWorkload
}

// NewController returns an empty, concurrency-safe controller.
func NewController() *Controller {
	return &Controller{workloads: make(map[string]*ManagedWorkload)}
}

// Create registers a ManagedWorkload with the cleanup finalizer.
func (c *Controller) Create(id string, replicas int, image string) (ManagedWorkload, error) {
	if !safeWorkloadID(id) || replicas < 1 || strings.TrimSpace(image) == "" {
		return ManagedWorkload{}, ErrInvalidWorkload
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.workloads[id]; exists {
		return ManagedWorkload{}, errors.New("ManagedWorkload already exists")
	}
	workload := &ManagedWorkload{
		ID:         id,
		Replicas:   replicas,
		Image:      image,
		Finalizers: []string{workloadFinalizer},
	}
	c.workloads[id] = workload
	return copyWorkload(workload), nil
}

// Get returns an isolated copy of a ManagedWorkload.
func (c *Controller) Get(id string) (ManagedWorkload, bool) {
	if !safeWorkloadID(id) {
		return ManagedWorkload{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	workload, ok := c.workloads[id]
	if !ok {
		return ManagedWorkload{}, false
	}
	return copyWorkload(workload), true
}

// Reconcile records Ready=True for a non-deleting ManagedWorkload.
func (c *Controller) Reconcile(id string) (ManagedWorkload, error) {
	if !safeWorkloadID(id) {
		return ManagedWorkload{}, ErrInvalidWorkload
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	workload, ok := c.workloads[id]
	if !ok {
		return ManagedWorkload{}, ErrWorkloadMissing
	}
	if workload.Deleting {
		return copyWorkload(workload), nil
	}
	setCondition(workload, "Ready", "True", "Reconciled", "desired replicas and image applied", time.Now().UTC())
	return copyWorkload(workload), nil
}

// Delete marks a ManagedWorkload for deletion while retaining its finalizer.
func (c *Controller) Delete(id string) (ManagedWorkload, error) {
	if !safeWorkloadID(id) {
		return ManagedWorkload{}, ErrInvalidWorkload
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	workload, ok := c.workloads[id]
	if !ok {
		return ManagedWorkload{}, ErrWorkloadMissing
	}
	workload.Deleting = true
	setCondition(workload, "Ready", "False", "Deleting", "waiting for finalizer cleanup", time.Now().UTC())
	return copyWorkload(workload), nil
}

// Finalize clears cleanup state and removes a deleting ManagedWorkload.
func (c *Controller) Finalize(id string) (ManagedWorkload, error) {
	if !safeWorkloadID(id) {
		return ManagedWorkload{}, ErrInvalidWorkload
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	workload, ok := c.workloads[id]
	if !ok {
		return ManagedWorkload{}, ErrWorkloadMissing
	}
	if !workload.Deleting {
		return ManagedWorkload{}, errors.New("ManagedWorkload is not deleting")
	}
	workload.Finalizers = nil
	finalized := copyWorkload(workload)
	delete(c.workloads, id)
	return finalized, nil
}

func setCondition(workload *ManagedWorkload, kind, status, reason, message string, now time.Time) {
	for i := range workload.Conditions {
		if workload.Conditions[i].Type != kind {
			continue
		}
		if workload.Conditions[i].Status != status {
			workload.Conditions[i].LastTransitionTime = now.UTC()
		}
		workload.Conditions[i].Status = status
		workload.Conditions[i].Reason = reason
		workload.Conditions[i].Message = message
		return
	}
	workload.Conditions = append(workload.Conditions, Condition{
		Type: kind, Status: status, Reason: reason, Message: message, LastTransitionTime: now.UTC(),
	})
}

func conditionTrue(workload ManagedWorkload, kind string) bool {
	for _, condition := range workload.Conditions {
		if condition.Type == kind {
			return condition.Status == "True"
		}
	}
	return false
}

func copyWorkload(workload *ManagedWorkload) ManagedWorkload {
	copied := *workload
	copied.Finalizers = append([]string(nil), workload.Finalizers...)
	copied.Conditions = append([]Condition(nil), workload.Conditions...)
	return copied
}

func safeWorkloadID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}
