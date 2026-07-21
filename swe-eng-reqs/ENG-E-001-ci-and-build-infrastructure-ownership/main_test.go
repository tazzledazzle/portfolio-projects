package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePipelines_POST_ValidDAG(t *testing.T) {
	state.Pipelines = map[string]any{}

	body := `{"stages":{"lint":[],"test":["lint"],"build":["test"]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePipelines(w, req)

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

	if resp["stages"] == nil {
		t.Error("response should contain stages")
	}
}

func TestHandlePipelines_POST_CyclicDAG(t *testing.T) {
	state.Pipelines = map[string]any{}

	body := `{"stages":{"A":["B"],"B":["A"]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePipelines(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for cyclic DAG, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "cycle") {
		t.Errorf("error should mention cycle, got: %s", w.Body.String())
	}
}

func TestHandlePipelines_GET_Existing(t *testing.T) {
	state.Pipelines = map[string]any{}

	createBody := `{"stages":{"lint":[],"test":["lint"]}}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handlePipelines(createW, createReq)

	var createResp map[string]any
	_ = json.Unmarshal(createW.Body.Bytes(), &createResp)
	pipelineID := createResp["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/v1/pipelines/"+pipelineID, nil)
	getW := httptest.NewRecorder()
	handlePipelineByID(getW, getReq, pipelineID)

	if getW.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var getResp map[string]any
	if err := json.Unmarshal(getW.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if getResp["id"] != pipelineID {
		t.Errorf("expected id %s, got %v", pipelineID, getResp["id"])
	}
}

func TestHandlePipelines_GET_NotFound(t *testing.T) {
	state.Pipelines = map[string]any{}

	req := httptest.NewRequest(http.MethodGet, "/v1/pipelines/nonexistent-id", nil)
	w := httptest.NewRecorder()
	handlePipelineByID(w, req, "nonexistent-id")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}
