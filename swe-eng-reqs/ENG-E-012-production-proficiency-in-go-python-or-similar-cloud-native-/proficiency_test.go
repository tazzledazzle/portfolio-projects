package main

import (
	"testing"
)

func TestProficiency_Info_LanguageGo(t *testing.T) {
	p := NewProficiency()
	info := p.Info()
	if info["language"] != "go" {
		t.Fatalf("expected language=go, got %#v", info["language"])
	}
}

func TestProficiency_ORSemantics(t *testing.T) {
	p := NewProficiency()
	info := p.Info()
	if info["or_semantics"] != true {
		t.Fatalf("expected or_semantics=true, got %#v", info["or_semantics"])
	}
	// Must NOT require both Go and Python.
	if info["languages_required"] == "go+python" {
		t.Fatal("languages_required must not mandate go+python")
	}
	if both, ok := info["both_languages_mandatory"].(bool); ok && both {
		t.Fatal("both_languages_mandatory must not be true")
	}
	langs, _ := info["languages_required"].([]string)
	hasGo, hasPy := false, false
	for _, l := range langs {
		if l == "go" {
			hasGo = true
		}
		if l == "python" {
			hasPy = true
		}
	}
	if hasGo && hasPy {
		t.Fatal("languages_required must not list both go and python as mandatory")
	}
}

func TestProficiency_ProductionSample(t *testing.T) {
	p := NewProficiency()
	sample := p.Sample()
	if sample == nil {
		t.Fatal("expected production sample")
	}
	if sample["handler"] != true && sample["health_path"] == nil {
		t.Fatalf("expected real Go handler or health path in sample: %#v", sample)
	}
	if sample["language"] != "go" {
		t.Fatalf("expected sample language=go, got %#v", sample["language"])
	}
}
