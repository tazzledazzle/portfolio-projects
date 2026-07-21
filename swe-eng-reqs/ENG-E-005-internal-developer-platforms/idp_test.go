package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestIDP_CreateProject_Lists(t *testing.T) {
	store := NewIDPStore()
	proj, err := store.CreateProject("payments")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if proj.ID == "" || proj.Name != "payments" {
		t.Fatalf("unexpected project: %#v", proj)
	}
	list := store.ListProjects()
	found := false
	for _, p := range list {
		if p.ID == proj.ID && p.Name == "payments" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListProjects missing created project: %#v", list)
	}
}

func TestIDP_CreatePipeline_UnderProject(t *testing.T) {
	store := NewIDPStore()
	proj, err := store.CreateProject("payments")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	pipe, err := store.CreatePipeline(proj.ID, "ci")
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	if pipe.ProjectID != proj.ID || pipe.Name != "ci" {
		t.Fatalf("unexpected pipeline: %#v", pipe)
	}
	if _, err := store.CreatePipeline("missing-project", "ci"); err == nil {
		t.Fatal("expected error for unknown project")
	}
}

func TestIDP_CreateEnvironment_UnderProject(t *testing.T) {
	store := NewIDPStore()
	proj, err := store.CreateProject("payments")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	env, err := store.CreateEnvironment(proj.ID, "staging")
	if err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if env.ProjectID != proj.ID || env.Name != "staging" {
		t.Fatalf("unexpected environment: %#v", env)
	}
	if _, err := store.CreateEnvironment("missing-project", "prod"); err == nil {
		t.Fatal("expected error for unknown project")
	}
}

func TestIDP_SafeIDs_RejectPathTraversal(t *testing.T) {
	store := NewIDPStore()
	for _, name := range []string{"..", "../x", "a/b", `a\b`, ""} {
		if _, err := store.CreateProject(name); err == nil {
			t.Fatalf("expected reject for project name %q", name)
		}
	}
	proj, err := store.CreateProject("ok")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	for _, name := range []string{"../pipe", "a/b", `a\b`} {
		if _, err := store.CreatePipeline(proj.ID, name); err == nil {
			t.Fatalf("expected reject for pipeline name %q", name)
		}
		if _, err := store.CreateEnvironment(proj.ID, name); err == nil {
			t.Fatalf("expected reject for environment name %q", name)
		}
	}
}

func TestIDP_ConcurrentCreate(t *testing.T) {
	store := NewIDPStore()
	var wg sync.WaitGroup
	errs := make(chan error, 60)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("proj-%d", i)
			proj, err := store.CreateProject(name)
			if err != nil {
				errs <- err
				return
			}
			if _, err := store.CreatePipeline(proj.ID, fmt.Sprintf("pipe-%d", i)); err != nil {
				errs <- err
				return
			}
			if _, err := store.CreateEnvironment(proj.ID, fmt.Sprintf("env-%d", i)); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent create: %v", err)
	}
	if got := len(store.ListProjects()); got != 20 {
		t.Fatalf("expected 20 projects, got %d", got)
	}
	if got := len(store.ListPipelines()); got != 20 {
		t.Fatalf("expected 20 pipelines, got %d", got)
	}
	if got := len(store.ListEnvironments()); got != 20 {
		t.Fatalf("expected 20 environments, got %d", got)
	}
}
