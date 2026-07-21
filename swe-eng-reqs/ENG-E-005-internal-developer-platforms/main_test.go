package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux(NewIDPStore())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleIDP_CRUD(t *testing.T) {
	mux := newMux(NewIDPStore())

	create := func(path, name string) map[string]any {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"name": name})
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("POST %s: expected 201, got %d body=%s", path, rr.Code, rr.Body.String())
		}
		var out map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return out
	}

	proj := create("/v1/projects", "payments")
	projID, _ := proj["id"].(string)
	if projID == "" {
		t.Fatalf("missing project id: %#v", proj)
	}

	pipeBody, _ := json.Marshal(map[string]string{"name": "ci", "project_id": projID})
	req := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewReader(pipeBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /v1/pipelines: %d %s", rr.Code, rr.Body.String())
	}

	envBody, _ := json.Marshal(map[string]string{"name": "staging", "project_id": projID})
	req = httptest.NewRequest(http.MethodPost, "/v1/environments", bytes.NewReader(envBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /v1/environments: %d %s", rr.Code, rr.Body.String())
	}

	for _, path := range []string{"/v1/projects", "/v1/pipelines", "/v1/environments"} {
		req = httptest.NewRequest(http.MethodGet, path, nil)
		rr = httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s: %d", path, rr.Code)
		}
		var list []any
		if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
			t.Fatalf("GET %s decode: %v body=%s", path, err, rr.Body.String())
		}
		if len(list) < 1 {
			t.Fatalf("GET %s: expected items, got %#v", path, list)
		}
	}
}

func TestHandleDemo_LiveProof(t *testing.T) {
	mux := newMux(NewIDPStore())
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo status %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %#v", result)
	}
	projects, _ := proof["projects"].(float64)
	pipelines, _ := proof["pipelines"].(float64)
	environments, _ := proof["environments"].(float64)
	selfService, _ := proof["self_service"].(bool)
	if projects < 1 || pipelines < 1 || environments < 1 || !selfService {
		t.Fatalf("live proof incomplete: %#v", proof)
	}
}
