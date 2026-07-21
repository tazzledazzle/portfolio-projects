package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthz(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "golden-path.md"), []byte("# golden\n"), 0o644)
	mux := newMux(NewProductStore(dir))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleProduct_SLAAdoptionGoldenPath(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "golden-path.md"), []byte("# golden path\n"), 0o644)
	store := NewProductStore(dir)
	store.RecordAdoption("alpha")
	mux := newMux(store)

	for _, path := range []string{"/v1/sla", "/v1/adoption", "/v1/golden-path"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, rr.Code, rr.Body.String())
		}
	}
}

func TestHandleDemo_ProductProof(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "golden-path.md"), []byte("# golden path\n"), 0o644)
	mux := newMux(NewProductStore(dir))
	req := httptest.NewRequest(http.MethodGet, "/v1/demo", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("demo %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	proof, _ := result["proof"].(map[string]any)
	if proof["sla"] != true || proof["adoption"] != true || proof["golden_path"] != true {
		t.Fatalf("proof incomplete: %#v", proof)
	}
}
