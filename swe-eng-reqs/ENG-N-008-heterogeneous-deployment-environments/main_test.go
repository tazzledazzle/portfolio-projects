package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleProfile_Schedule(t *testing.T) {
	store = NewProfileStore()
	mux := newMux()

	for _, name := range []string{"k8s-standard", "k8s-gpu", "vm-bake"} {
		body, _ := json.Marshal(map[string]any{"runtime": "kubernetes"})
		req := httptest.NewRequest(http.MethodPut, "/v1/profiles/"+name, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
			t.Fatalf("upsert %s status=%d body=%s", name, rr.Code, rr.Body.String())
		}
	}

	schedBody, _ := json.Marshal(map[string]any{"profiles": []string{"k8s-standard", "k8s-gpu", "vm-bake"}})
	req := httptest.NewRequest(http.MethodPost, "/v1/workloads/checkout-api/schedule", bytes.NewReader(schedBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		t.Fatalf("schedule status=%d body=%s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/workloads/checkout-api/placements", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("placements status=%d body=%s", rr.Code, rr.Body.String())
	}
	var placements []Placement
	if err := json.Unmarshal(rr.Body.Bytes(), &placements); err != nil {
		t.Fatalf("decode placements: %v", err)
	}
	if len(placements) < 3 {
		t.Fatalf("placements=%d, want >=3", len(placements))
	}
}

func TestHandleDemo_ProfileProof(t *testing.T) {
	store = NewProfileStore()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPost, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo status=%d body=%s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	proof, ok := result["proof"].(map[string]any)
	if !ok {
		t.Fatalf("missing proof: %v", result)
	}
	profiles, ok := proof["profiles"].(float64)
	if !ok || profiles < 3 {
		t.Fatalf("proof.profiles=%v, want >=3", proof["profiles"])
	}
	if proof["same_workload"] != true {
		t.Fatalf("proof.same_workload=%v, want true", proof["same_workload"])
	}
	placements, ok := proof["placements"].(float64)
	if !ok || placements < 3 {
		t.Fatalf("proof.placements=%v, want >=3", proof["placements"])
	}
}
