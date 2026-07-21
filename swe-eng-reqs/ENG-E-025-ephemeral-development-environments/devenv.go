package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidDevEnv = errors.New("invalid DevEnv id or ttl")
	ErrDevEnvMissing = errors.New("DevEnv not found")
)

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

type DevEnv struct {
	ID         string      `json:"id"`
	TTLSeconds int64       `json:"ttlSeconds"`
	CreatedAt  time.Time   `json:"createdAt"`
	Conditions []Condition `json:"conditions"`
	Reclaimed  bool        `json:"reclaimed"`
}

type Controller struct {
	mu   sync.RWMutex
	envs map[string]*DevEnv
}

func NewController() *Controller {
	return &Controller{envs: make(map[string]*DevEnv)}
}

func (c *Controller) Create(id string, ttlSeconds int64, now time.Time) (DevEnv, error) {
	if !safeDevEnvID(id) || ttlSeconds <= 0 {
		return DevEnv{}, ErrInvalidDevEnv
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.envs[id]; exists {
		return DevEnv{}, errors.New("DevEnv already exists")
	}
	env := &DevEnv{ID: id, TTLSeconds: ttlSeconds, CreatedAt: now.UTC()}
	setCondition(env, "Ready", "True", "Provisioned", "environment ready", now)
	setCondition(env, "Expired", "False", "TTLActive", "TTL has not elapsed", now)
	c.envs[id] = env
	return copyDevEnv(env), nil
}

func (c *Controller) Get(id string) (DevEnv, bool) {
	if !safeDevEnvID(id) {
		return DevEnv{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	env, ok := c.envs[id]
	if !ok {
		return DevEnv{}, false
	}
	return copyDevEnv(env), true
}

func (c *Controller) Tick(id string, now time.Time) (DevEnv, error) {
	if !safeDevEnvID(id) {
		return DevEnv{}, ErrInvalidDevEnv
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	env, ok := c.envs[id]
	if !ok {
		return DevEnv{}, ErrDevEnvMissing
	}
	reconcileDevEnv(env, now.UTC())
	return copyDevEnv(env), nil
}

func (c *Controller) Reconcile(now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	reclaimed := 0
	for _, env := range c.envs {
		wasReclaimed := env.Reclaimed
		reconcileDevEnv(env, now.UTC())
		if !wasReclaimed && env.Reclaimed {
			reclaimed++
		}
	}
	return reclaimed
}

func reconcileDevEnv(env *DevEnv, now time.Time) {
	expiresAt := env.CreatedAt.Add(time.Duration(env.TTLSeconds) * time.Second)
	if !now.Before(expiresAt) {
		setCondition(env, "Expired", "True", "TTLElapsed", "TTL elapsed; resources reclaimed", now)
		setCondition(env, "Ready", "False", "Expired", "environment expired", now)
		env.Reclaimed = true
		return
	}
	setCondition(env, "Ready", "True", "Provisioned", "environment ready", now)
	setCondition(env, "Expired", "False", "TTLActive", "TTL has not elapsed", now)
}

func setCondition(env *DevEnv, kind, status, reason, message string, now time.Time) {
	for i := range env.Conditions {
		if env.Conditions[i].Type == kind {
			if env.Conditions[i].Status != status {
				env.Conditions[i].LastTransitionTime = now.UTC()
			}
			env.Conditions[i].Status = status
			env.Conditions[i].Reason = reason
			env.Conditions[i].Message = message
			return
		}
	}
	env.Conditions = append(env.Conditions, Condition{
		Type: kind, Status: status, Reason: reason, Message: message, LastTransitionTime: now.UTC(),
	})
}

func conditionTrue(env DevEnv, kind string) bool {
	for _, condition := range env.Conditions {
		if condition.Type == kind {
			return condition.Status == "True"
		}
	}
	return false
}

func copyDevEnv(env *DevEnv) DevEnv {
	copy := *env
	copy.Conditions = append([]Condition(nil), env.Conditions...)
	return copy
}

func safeDevEnvID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}
