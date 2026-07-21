package main

import (
	"strings"
	"testing"
)

func TestPipelineService_Submit(t *testing.T) {
	svc := NewPipelineService()

	dag := map[string][]string{
		"lint":  {},
		"test":  {"lint"},
		"build": {"test"},
	}

	pipeline, err := svc.SubmitPipeline(dag)
	if err != nil {
		t.Fatalf("SubmitPipeline failed: %v", err)
	}

	if pipeline.ID == "" {
		t.Error("pipeline should have an ID")
	}

	if len(pipeline.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(pipeline.Stages))
	}

	for _, stage := range pipeline.Stages {
		if stage.Status != "pending" {
			t.Errorf("all stages should be pending, got %s", stage.Status)
		}
	}
}

func TestPipelineService_Submit_CyclicDAG(t *testing.T) {
	svc := NewPipelineService()

	dag := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}

	_, err := svc.SubmitPipeline(dag)
	if err == nil {
		t.Error("cyclic DAG should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention cycle, got: %v", err)
	}
}

func TestPipelineService_GetPipeline(t *testing.T) {
	svc := NewPipelineService()

	dag := map[string][]string{"stage1": {}}
	submitted, _ := svc.SubmitPipeline(dag)

	retrieved, err := svc.GetPipeline(submitted.ID)
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}

	if retrieved.ID != submitted.ID {
		t.Errorf("expected ID %s, got %s", submitted.ID, retrieved.ID)
	}
}

func TestPipelineService_GetPipeline_NotFound(t *testing.T) {
	svc := NewPipelineService()

	_, err := svc.GetPipeline("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent pipeline")
	}
}

func TestPipelineService_TransitionStage(t *testing.T) {
	svc := NewPipelineService()

	dag := map[string][]string{"stage1": {}}
	pipeline, _ := svc.SubmitPipeline(dag)

	err := svc.TransitionStage(pipeline.ID, "stage1", "running")
	if err != nil {
		t.Fatalf("TransitionStage failed: %v", err)
	}

	updated, _ := svc.GetPipeline(pipeline.ID)
	if updated.Stages["stage1"].Status != "running" {
		t.Errorf("expected status running, got %s", updated.Stages["stage1"].Status)
	}
}

func TestPipelineService_TransitionStage_InvalidTransition(t *testing.T) {
	svc := NewPipelineService()

	dag := map[string][]string{"stage1": {}}
	pipeline, _ := svc.SubmitPipeline(dag)

	err := svc.TransitionStage(pipeline.ID, "stage1", "succeeded")
	if err == nil {
		t.Error("expected error for invalid transition pending→succeeded")
	}
}

func TestPipelineService_ListPipelines(t *testing.T) {
	svc := NewPipelineService()

	svc.SubmitPipeline(map[string][]string{"a": {}})
	svc.SubmitPipeline(map[string][]string{"b": {}})
	svc.SubmitPipeline(map[string][]string{"c": {}})

	pipelines := svc.ListPipelines()
	if len(pipelines) != 3 {
		t.Errorf("expected 3 pipelines, got %d", len(pipelines))
	}
}
