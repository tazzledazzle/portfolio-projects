package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProbe_LivenessReadyz(t *testing.T) {
	p := NewPackaging(".", "k8s/deploy.yaml")
	probes := p.Probes()
	if probes["liveness"] != "/healthz" {
		t.Fatalf("expected liveness /healthz, got %#v", probes)
	}
	if probes["readiness"] != "/readyz" {
		t.Fatalf("expected readiness /readyz, got %#v", probes)
	}
}

func TestHPA_ReadyFlag(t *testing.T) {
	p := NewPackaging(".", "k8s/deploy.yaml")
	if !p.HPAReady() {
		t.Fatal("HPAReady must be true when deploy.yaml has HPA / autoscaling markers")
	}
}

func TestPackaging_ComposeParity(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Dockerfile"), "FROM alpine\n")
	mustWrite(t, filepath.Join(root, "compose.yaml"), "services: {}\n")
	deploy := filepath.Join(root, "k8s", "deploy.yaml")
	mustWrite(t, deploy, "kind: HorizontalPodAutoscaler\nlivenessProbe:\nreadinessProbe:\n")

	p := NewPackaging(root, deploy)
	info := p.Packaging()
	if info["dockerfile"] != true {
		t.Fatalf("expected dockerfile true, got %#v", info)
	}
	if info["compose_parity"] != true {
		t.Fatalf("expected compose_parity true, got %#v", info)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
