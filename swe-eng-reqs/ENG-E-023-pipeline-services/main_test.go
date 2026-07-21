package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePipelineSubmit(t *testing.T) {
	pipelineSvc = NewPipelineService()

	body := `{"stages":{"lint":[],"test":["lint"],"build":["test"]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePipelineSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["id"] == nil || resp["id"] == "" {
		t.Error("response should contain pipeline id")
	}

	stages, ok := resp["stages"].(map[string]any)
	if !ok || len(stages) != 3 {
		t.Errorf("response should contain 3 stages, got %v", resp["stages"])
	}
}

func TestHandlePipelineSubmit_CyclicDAG(t *testing.T) {
	pipelineSvc = NewPipelineService()

	body := `{"stages":{"A":["B"],"B":["A"]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePipelineSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "cycle") {
		t.Errorf("error should mention cycle, got: %s", w.Body.String())
	}
}

func TestHandlePipelineGet(t *testing.T) {
	pipelineSvc = NewPipelineService()

	dag := map[string][]string{"stage1": {}}
	pipeline, _ := pipelineSvc.SubmitPipeline(dag)

	req := httptest.NewRequest(http.MethodGet, "/v1/pipelines/"+pipeline.ID, nil)
	w := httptest.NewRecorder()

	handlePipelineGet(w, req, pipeline.ID)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["id"] != pipeline.ID {
		t.Errorf("expected id %s, got %v", pipeline.ID, resp["id"])
	}
}

func TestHandlePipelineNotFound(t *testing.T) {
	pipelineSvc = NewPipelineService()

	req := httptest.NewRequest(http.MethodGet, "/v1/pipelines/nonexistent", nil)
	w := httptest.NewRecorder()

	handlePipelineGet(w, req, "nonexistent")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePipelineTransition(t *testing.T) {
	pipelineSvc = NewPipelineService()

	dag := map[string][]string{"stage1": {}}
	pipeline, _ := pipelineSvc.SubmitPipeline(dag)

	body := `{"status":"running"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines/"+pipeline.ID+"/stages/stage1/transition", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleStageTransition(w, req, pipeline.ID, "stage1")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, _ := pipelineSvc.GetPipeline(pipeline.ID)
	if updated.Stages["stage1"].Status != "running" {
		t.Errorf("expected status running, got %s", updated.Stages["stage1"].Status)
	}
}

func TestHandlePipelineList(t *testing.T) {
	pipelineSvc = NewPipelineService()

	pipelineSvc.SubmitPipeline(map[string][]string{"a": {}})
	pipelineSvc.SubmitPipeline(map[string][]string{"b": {}})

	req := httptest.NewRequest(http.MethodGet, "/v1/pipelines", nil)
	w := httptest.NewRecorder()

	handlePipelineList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("expected 2 pipelines, got %d", len(resp))
	}
}
