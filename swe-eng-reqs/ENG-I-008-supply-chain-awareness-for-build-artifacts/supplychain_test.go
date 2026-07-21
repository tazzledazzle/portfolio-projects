package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestSign_Ed25519_VerifyRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSupplyChain(priv, pub)
	digest := "sha256:" + hex.EncodeToString(make([]byte, 32))
	sig, err := sc.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	ok, err := sc.Verify(digest, sig)
	if err != nil || !ok {
		t.Fatalf("Verify = %v, %v", ok, err)
	}
	ok, err = sc.Verify(digest+"x", sig)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Verify must fail for mutated digest")
	}
}

func TestSBOM_SPDXInspired(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSupplyChain(priv, pub)
	sbom := sc.SBOM("demo-artifact", "sha256:abc")
	if sbom["spdx_inspired"] != true {
		t.Fatalf("spdx_inspired missing: %+v", sbom)
	}
	if sbom["name"] != "demo-artifact" {
		t.Fatalf("name=%v", sbom["name"])
	}
	if sbom["spdxVersion"] == nil && sbom["spdx_version"] == nil {
		t.Fatalf("expected SPDX-inspired version field: %+v", sbom)
	}
}

func TestScope_AuthorizePush_DefaultDeny(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSupplyChain(priv, pub)
	if ok, _ := sc.AuthorizePush(nil); ok {
		t.Fatal("missing scope must deny")
	}
	if ok, _ := sc.AuthorizePush([]string{"artifacts:read"}); ok {
		t.Fatal("read-only must deny push")
	}
	if ok, err := sc.AuthorizePush([]string{"artifacts:push"}); !ok || err != nil {
		t.Fatalf("push scope must allow: ok=%v err=%v", ok, err)
	}
}

func TestInfo_SigstoreFalse(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSupplyChain(priv, pub)
	info := sc.Info()
	if info["sigstore"] != false {
		t.Fatalf("sigstore want false: %+v", info)
	}
	if info["signing"] != "ed25519" {
		t.Fatalf("signing want ed25519: %+v", info)
	}
	if info["spdx_inspired"] != true {
		t.Fatalf("spdx_inspired want true: %+v", info)
	}
}
