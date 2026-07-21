package main

import (
	"strings"
	"testing"
)

const (
	digestX = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	digestY = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	digestZ = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestScope_DenyMissing(t *testing.T) {
	eng := NewPolicyEngine()
	ok, err := eng.Authorize(nil)
	if ok || err == nil {
		t.Fatalf("Authorize(nil): ok=%v err=%v, want deny", ok, err)
	}
	ok, err = eng.Authorize([]string{"artifacts:read"})
	if ok || err == nil {
		t.Fatalf("Authorize(read-only): ok=%v err=%v, want deny", ok, err)
	}
}

func TestScope_AllowWithScope(t *testing.T) {
	eng := NewPolicyEngine()
	ok, err := eng.Authorize([]string{"artifacts:write"})
	if !ok || err != nil {
		t.Fatalf("Authorize(write): ok=%v err=%v", ok, err)
	}
	if err := eng.PutArtifact("pkg", digestX, []string{"artifacts:write"}); err != nil {
		t.Fatalf("PutArtifact: %v", err)
	}
}

func TestRetention_DeletesOld(t *testing.T) {
	eng := NewPolicyEngine()
	scopes := []string{"artifacts:write"}
	for i, d := range []string{digestX, digestY, digestZ} {
		name := "art-" + string(rune('a'+i))
		if err := eng.PutArtifact(name, d, scopes); err != nil {
			t.Fatalf("PutArtifact %s: %v", name, err)
		}
	}
	deleted, err := eng.RunRetention(2)
	if err != nil {
		t.Fatalf("RunRetention: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted=%d, want 1", deleted)
	}
	arts := eng.ListArtifacts()
	if len(arts) != 2 {
		t.Fatalf("kept=%d, want 2", len(arts))
	}
	for _, a := range arts {
		if a.Digest != digestY && a.Digest != digestZ {
			t.Fatalf("unexpected kept digest %q", a.Digest)
		}
		if !strings.HasPrefix(a.Digest, "sha256:") {
			t.Fatalf("digest form: %q", a.Digest)
		}
	}
}

func TestScan_FixtureFindings(t *testing.T) {
	eng := NewPolicyEngine()
	findings := eng.Scan()
	if len(findings) == 0 {
		t.Fatal("Scan: expected fixture findings")
	}
	for _, f := range findings {
		if f.Severity == "" || f.ID == "" {
			t.Fatalf("incomplete finding: %+v", f)
		}
	}
}

func TestSimulator_Flag(t *testing.T) {
	eng := NewPolicyEngine()
	info := eng.Info()
	if info["simulator"] != true {
		t.Fatalf("simulator=%v, want true", info["simulator"])
	}
	if info["vendor_model"] != "custom-registry" {
		t.Fatalf("vendor_model=%v, want custom-registry", info["vendor_model"])
	}
}

func TestRetention_NeverRewritesDigest(t *testing.T) {
	eng := NewPolicyEngine()
	scopes := []string{"artifacts:write"}
	_ = eng.PutArtifact("keep-me", digestX, scopes)
	_ = eng.PutArtifact("drop-me", digestY, scopes)
	_ = eng.PutArtifact("keep-too", digestZ, scopes)
	before := map[string]string{}
	for _, a := range eng.ListArtifacts() {
		before[a.Name] = a.Digest
	}
	if _, err := eng.RunRetention(2); err != nil {
		t.Fatalf("RunRetention: %v", err)
	}
	for _, a := range eng.ListArtifacts() {
		if before[a.Name] != a.Digest {
			t.Fatalf("digest rewritten for %s: %q → %q", a.Name, before[a.Name], a.Digest)
		}
	}
}
