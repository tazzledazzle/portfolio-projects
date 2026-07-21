package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrUnsafeID        = errors.New("unsafe id: path separators or traversal rejected")
	ErrProjectNotFound = errors.New("project not found")
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Pipeline struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type Environment struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type IDPStore struct {
	mu           sync.Mutex
	projects     map[string]*Project
	pipelines    map[string]*Pipeline
	environments map[string]*Environment
	counter      int
}

func NewIDPStore() *IDPStore {
	return &IDPStore{
		projects:     make(map[string]*Project),
		pipelines:    make(map[string]*Pipeline),
		environments: make(map[string]*Environment),
	}
}

func safeID(name string) bool {
	return name != "" && !strings.Contains(name, "..") && !strings.ContainsAny(name, `/\`)
}

func (s *IDPStore) CreateProject(name string) (*Project, error) {
	if !safeID(name) {
		return nil, ErrUnsafeID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	p := &Project{ID: fmt.Sprintf("proj-%d", s.counter), Name: name}
	s.projects[p.ID] = p
	out := *p
	return &out, nil
}

func (s *IDPStore) CreatePipeline(projectID, name string) (*Pipeline, error) {
	if !safeID(name) {
		return nil, ErrUnsafeID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}
	s.counter++
	p := &Pipeline{ID: fmt.Sprintf("pipe-%d", s.counter), ProjectID: projectID, Name: name}
	s.pipelines[p.ID] = p
	out := *p
	return &out, nil
}

func (s *IDPStore) CreateEnvironment(projectID, name string) (*Environment, error) {
	if !safeID(name) {
		return nil, ErrUnsafeID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}
	s.counter++
	e := &Environment{ID: fmt.Sprintf("env-%d", s.counter), ProjectID: projectID, Name: name}
	s.environments[e.ID] = e
	out := *e
	return &out, nil
}

func (s *IDPStore) ListProjects() []Project {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Project, 0, len(s.projects))
	for _, p := range s.projects {
		out = append(out, *p)
	}
	return out
}

func (s *IDPStore) ListPipelines() []Pipeline {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Pipeline, 0, len(s.pipelines))
	for _, p := range s.pipelines {
		out = append(out, *p)
	}
	return out
}

func (s *IDPStore) ListEnvironments() []Environment {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Environment, 0, len(s.environments))
	for _, e := range s.environments {
		out = append(out, *e)
	}
	return out
}
