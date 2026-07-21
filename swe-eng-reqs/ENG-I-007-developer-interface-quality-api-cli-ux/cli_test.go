package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_StatusJSON_Valid(t *testing.T) {
	stdout, err := RunCLI([]string{"status", "--json"})
	if err != nil {
		t.Fatalf("RunCLI status --json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if payload["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", payload["ok"])
	}
	if _, ok := payload["status"]; !ok {
		t.Fatalf("expected status field in JSON, got %#v", payload)
	}
}

func TestCLI_Golden_Match(t *testing.T) {
	stdout, err := RunCLI([]string{"status", "--json"})
	if err != nil {
		t.Fatalf("RunCLI: %v", err)
	}
	goldenPath := filepath.Join("testdata", "status.json.golden")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	got := normalizeJSONWhitespace(stdout)
	expected := normalizeJSONWhitespace(string(want))
	if got != expected {
		t.Fatalf("golden mismatch\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestCLI_ClearError_UnknownCommand(t *testing.T) {
	stdout, err := RunCLI([]string{"not-a-command"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	msg := err.Error()
	if strings.TrimSpace(msg) == "" {
		t.Fatal("expected clear non-empty error message")
	}
	if strings.Contains(strings.ToLower(msg), "panic") {
		t.Fatalf("error must not mention panic: %s", msg)
	}
	if strings.Contains(msg, "SECRET") || strings.Contains(msg, "password") {
		t.Fatalf("error must not leak secrets: %s", msg)
	}
	_ = stdout
}

func TestCLI_MissingFlag_ClearError(t *testing.T) {
	_, err := RunCLI([]string{"status"})
	if err == nil {
		t.Fatal("expected clear error when --json is missing")
	}
	msg := err.Error()
	if !strings.Contains(strings.ToLower(msg), "json") && !strings.Contains(strings.ToLower(msg), "usage") {
		t.Fatalf("expected usage/--json guidance in error, got: %s", msg)
	}
}

