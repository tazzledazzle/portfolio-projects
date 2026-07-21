package main

import (
	"strings"
	"testing"
)

func TestParseJUnit_SingleTestCase(t *testing.T) {
	xml := `<?xml version="1.0"?>
<testsuite name="suite1" tests="1">
  <testcase name="test1" classname="pkg.Test" time="0.5"/>
</testsuite>`

	results, err := ParseJUnit(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseJUnit() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("ParseJUnit() got %d results, want 1", len(results))
	}
	if results[0].Name != "test1" {
		t.Errorf("results[0].Name = %q, want %q", results[0].Name, "test1")
	}
	if results[0].Status != "passed" {
		t.Errorf("results[0].Status = %q, want %q", results[0].Status, "passed")
	}
}

func TestParseJUnit_TestSuiteNested(t *testing.T) {
	xml := `<?xml version="1.0"?>
<testsuites>
  <testsuite name="suite1" tests="2">
    <testcase name="test1" classname="pkg.Test" time="0.1"/>
    <testcase name="test2" classname="pkg.Test" time="0.2"/>
  </testsuite>
</testsuites>`

	results, err := ParseJUnit(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseJUnit() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ParseJUnit() got %d results, want 2", len(results))
	}
}

func TestParseJUnit_FailureElement(t *testing.T) {
	xml := `<?xml version="1.0"?>
<testsuite name="suite1" tests="1">
  <testcase name="test1" classname="pkg.Test" time="0.5">
    <failure message="assertion failed">expected true but got false</failure>
  </testcase>
</testsuite>`

	results, err := ParseJUnit(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseJUnit() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("ParseJUnit() got %d results, want 1", len(results))
	}
	if results[0].Status != "failed" {
		t.Errorf("results[0].Status = %q, want %q", results[0].Status, "failed")
	}
}

func TestParseJUnit_SkippedElement(t *testing.T) {
	xml := `<?xml version="1.0"?>
<testsuite name="suite1" tests="1">
  <testcase name="test1" classname="pkg.Test" time="0.0">
    <skipped message="not implemented"/>
  </testcase>
</testsuite>`

	results, err := ParseJUnit(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseJUnit() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("ParseJUnit() got %d results, want 1", len(results))
	}
	if results[0].Status != "skipped" {
		t.Errorf("results[0].Status = %q, want %q", results[0].Status, "skipped")
	}
}

func TestParseJUnit_MalformedXML(t *testing.T) {
	xml := `<?xml version="1.0"?><testsuite><testcase name=`

	_, err := ParseJUnit(strings.NewReader(xml))
	if err == nil {
		t.Error("ParseJUnit() expected error for malformed XML, got nil")
	}
}

func TestParseJUnit_EmptyXML(t *testing.T) {
	results, err := ParseJUnit(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ParseJUnit() error = %v, want nil for empty input", err)
	}
	if len(results) != 0 {
		t.Errorf("ParseJUnit() got %d results, want 0 for empty input", len(results))
	}
}
