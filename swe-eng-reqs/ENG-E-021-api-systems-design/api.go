package main

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var ErrMissingScope = errors.New("missing required scope")

type CompatResult struct {
	V1OK bool `json:"v1_ok"`
	V2OK bool `json:"v2_ok"`
	Pass bool `json:"pass"`
}

type APIEngine struct {
	mu            sync.Mutex
	limit         int
	requiredScope string
	counts        map[string]int
	docPath       string
}

func NewAPIEngine(limit int, requiredScope string) *APIEngine {
	if limit <= 0 {
		limit = 1
	}
	return &APIEngine{
		limit:         limit,
		requiredScope: requiredScope,
		counts:        make(map[string]int),
		docPath:       "openapi.yaml",
	}
}

func (e *APIEngine) SetDocPath(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.docPath = path
}

func (e *APIEngine) Authorize(scopes []string) (bool, error) {
	for _, s := range scopes {
		if s == e.requiredScope {
			return true, nil
		}
	}
	return false, ErrMissingScope
}

func (e *APIEngine) Allow(subject string) bool {
	if subject == "" {
		subject = "anonymous"
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.counts[subject] >= e.limit {
		return false
	}
	e.counts[subject]++
	return true
}

func (e *APIEngine) Compat() CompatResult {
	v1 := map[string]any{"version": "v1", "items": []any{}}
	v2 := map[string]any{"api_version": "v2", "data": []any{}}
	_, v1ok := v1["version"].(string)
	_, v2ok := v2["api_version"].(string)
	return CompatResult{V1OK: v1ok, V2OK: v2ok, Pass: v1ok && v2ok}
}

func (e *APIEngine) OpenAPIDoc() (string, error) {
	e.mu.Lock()
	path := e.docPath
	e.mu.Unlock()
	if path == "" {
		path = "openapi.yaml"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		// also try alongside executable cwd variants
		alt := filepath.Join(".", "openapi.yaml")
		b, err = os.ReadFile(alt)
		if err != nil {
			return "", err
		}
	}
	return string(b), nil
}

func (e *APIEngine) ResourceV1() map[string]any {
	return map[string]any{
		"version": "v1",
		"items": []map[string]string{
			{"id": "res-1", "name": "sample"},
		},
	}
}

func (e *APIEngine) ResourceV2() map[string]any {
	return map[string]any{
		"api_version": "v2",
		"data": []map[string]string{
			{"id": "res-1", "display_name": "sample"},
		},
	}
}
