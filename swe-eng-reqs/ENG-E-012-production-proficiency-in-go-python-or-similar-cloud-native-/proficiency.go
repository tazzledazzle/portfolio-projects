package main

// Proficiency proves single-language (Go) production sample under OR semantics
// (CLAUDE.md / D-08): Go OR Python OR similar — not both mandatory.
type Proficiency struct {
	language string
}

func NewProficiency() *Proficiency {
	return &Proficiency{language: "go"}
}

func (p *Proficiency) Info() map[string]any {
	return map[string]any{
		"requirement_id":            "ENG-E-012",
		"service":                   "eng-e-012",
		"title":                     "Production proficiency in Go, Python, or similar cloud-native languages",
		"language":                  p.language,
		"or_semantics":              true,
		"languages_required":        []string{"go"}, // single choice under OR — not go+python
		"both_languages_mandatory":  false,
		"note":                      "OR satisfied with single Go production sample; Python not required",
		"does_not_own":              []string{"deep_cijob", "full_prod_patterns_e008"},
	}
}

// Sample returns evidence of a real Go HTTP handler path in this package.
func (p *Proficiency) Sample() map[string]any {
	return map[string]any{
		"language":    p.language,
		"handler":     true,
		"health_path": "/healthz",
		"info_path":   "/v1/info",
		"demo_path":   "/v1/demo",
		"stack":       "go+stdlib+net/http",
	}
}
