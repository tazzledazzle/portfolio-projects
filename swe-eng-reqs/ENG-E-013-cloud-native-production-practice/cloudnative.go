package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Packaging reports Dockerfile / compose / probe / HPA facts derived from real files.
type Packaging struct {
	mu         sync.Mutex
	root       string
	deployPath string
	demos      int
}

// NewPackaging inspects packaging facts under root using the given deploy manifest path.
func NewPackaging(root, deployPath string) *Packaging {
	return &Packaging{root: root, deployPath: deployPath}
}

// Probes returns liveness and readiness HTTP paths from the deploy manifest (or defaults).
func (p *Packaging) Probes() map[string]string {
	content := p.readDeploy()
	liveness := "/healthz"
	readiness := "/readyz"
	if strings.Contains(content, "livenessProbe") {
		if path := probePath(content, "livenessProbe"); path != "" {
			liveness = path
		}
	}
	if strings.Contains(content, "readinessProbe") {
		if path := probePath(content, "readinessProbe"); path != "" {
			readiness = path
		}
	}
	return map[string]string{
		"liveness":  liveness,
		"readiness": readiness,
	}
}

// HPAReady is true when the deploy manifest includes HPA / autoscaling markers.
func (p *Packaging) HPAReady() bool {
	content := p.readDeploy()
	markers := []string{
		"HorizontalPodAutoscaler",
		"autoscaling/v2",
		"hpa-ready",
		"hpa:",
	}
	lower := strings.ToLower(content)
	for _, m := range markers {
		if strings.Contains(content, m) || strings.Contains(lower, strings.ToLower(m)) {
			return true
		}
	}
	return false
}

// Packaging reports dockerfile and compose_parity derived from filesystem presence.
func (p *Packaging) Packaging() map[string]any {
	dockerfile := fileExists(filepath.Join(p.root, "Dockerfile"))
	compose := fileExists(filepath.Join(p.root, "compose.yaml")) ||
		fileExists(filepath.Join(p.root, "docker-compose.yaml")) ||
		fileExists(filepath.Join(p.root, "compose.yml"))
	return map[string]any{
		"dockerfile":     dockerfile,
		"compose_parity": dockerfile && compose,
		"hpa_ready":      p.HPAReady(),
		"probes":         p.Probes(),
	}
}

// IncrementDemo records a demo invocation.
func (p *Packaging) IncrementDemo() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.demos++
	return p.demos
}

// DemoCount returns demo invocations.
func (p *Packaging) DemoCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.demos
}

func (p *Packaging) readDeploy() string {
	b, err := os.ReadFile(p.deployPath)
	if err != nil {
		return ""
	}
	return string(b)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func probePath(content, probeKey string) string {
	idx := strings.Index(content, probeKey)
	if idx < 0 {
		return ""
	}
	slice := content[idx:]
	pathIdx := strings.Index(slice, "path:")
	if pathIdx < 0 || pathIdx > 200 {
		return ""
	}
	line := strings.TrimSpace(strings.SplitN(slice[pathIdx:], "\n", 2)[0])
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
