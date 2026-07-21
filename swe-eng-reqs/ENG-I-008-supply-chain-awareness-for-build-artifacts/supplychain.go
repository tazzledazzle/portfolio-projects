package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
)

const requiredPushScope = "artifacts:push"

var (
	ErrInvalidDigest = errors.New("invalid digest")
	ErrUnsigned      = errors.New("signature required before push")
	ErrBadSignature  = errors.New("signature verification failed")
	ErrScopeDenied   = errors.New("missing required scope: artifacts:push")
)

// SupplyChain signs digests with ed25519, emits SPDX-inspired SBOMs, and enforces push scopes.
// Honesty: spdx_inspired=true, sigstore=false — no cosign/Sigstore.
type SupplyChain struct {
	mu         sync.Mutex
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	artifacts  map[string]string // digest → signature
}

// NewSupplyChain constructs a supply-chain engine with the given key pair.
func NewSupplyChain(priv ed25519.PrivateKey, pub ed25519.PublicKey) *SupplyChain {
	return &SupplyChain{
		privateKey: priv,
		publicKey:  pub,
		artifacts:  make(map[string]string),
	}
}

// LoadSupplyChainFromFiles loads raw ed25519 key bytes from paths (fixture keys only).
func LoadSupplyChainFromFiles(privPath, pubPath string) (*SupplyChain, error) {
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, err
	}
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, err
	}
	if len(privBytes) != ed25519.PrivateKeySize || len(pubBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid ed25519 key size")
	}
	return NewSupplyChain(ed25519.PrivateKey(privBytes), ed25519.PublicKey(pubBytes)), nil
}

// Sign signs a content digest with ed25519 (stdlib only).
func (s *SupplyChain) Sign(digest string) (string, error) {
	if !validDigest(digest) {
		return "", ErrInvalidDigest
	}
	sig := ed25519.Sign(s.privateKey, []byte(digest))
	encoded := base64.StdEncoding.EncodeToString(sig)
	s.mu.Lock()
	s.artifacts[digest] = encoded
	s.mu.Unlock()
	return encoded, nil
}

// Verify checks an ed25519 signature over digest.
func (s *SupplyChain) Verify(digest, signature string) (bool, error) {
	if !validDigest(digest) {
		return false, nil
	}
	raw, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, nil
	}
	return ed25519.Verify(s.publicKey, []byte(digest), raw), nil
}

// SBOM returns an SPDX-inspired document (not a full SPDX conformance claim).
func (s *SupplyChain) SBOM(name, digest string) map[string]any {
	return map[string]any{
		"spdx_inspired": true,
		"spdxVersion":   "SPDX-2.3-inspired",
		"name":          name,
		"documentNamespace": "https://example.invalid/spdx/" + name,
		"packages": []map[string]any{
			{
				"name":             name,
				"SPDXID":           "SPDXRef-Package-" + sanitizeID(name),
				"versionInfo":      "0.0.0-demo",
				"downloadLocation": "NOASSERTION",
				"filesAnalyzed":    false,
				"checksums": []map[string]string{
					{"algorithm": "SHA256", "checksumValue": strings.TrimPrefix(digest, "sha256:")},
				},
			},
		},
		"creationInfo": map[string]any{
			"created":  "2026-07-18T00:00:00Z",
			"creators": []string{"Tool: eng-i-008-supplychain-simulator"},
		},
		"note": "SPDX-inspired fields only — not a signed SPDX document or Sigstore attestation",
	}
}

// AuthorizePush default-denies unless artifacts:push is present (T-5-17).
func (s *SupplyChain) AuthorizePush(scopes []string) (bool, error) {
	for _, scope := range scopes {
		if scope == requiredPushScope {
			return true, nil
		}
	}
	return false, ErrScopeDenied
}

// Push requires a verified signature and push scope before accepting (T-5-17).
func (s *SupplyChain) Push(name, digest, signature string, scopes []string) error {
	ok, err := s.AuthorizePush(scopes)
	if !ok {
		return err
	}
	if signature == "" {
		return ErrUnsigned
	}
	valid, err := s.Verify(digest, signature)
	if err != nil {
		return err
	}
	if !valid {
		return ErrBadSignature
	}
	s.mu.Lock()
	s.artifacts[digest] = signature
	s.mu.Unlock()
	_ = name
	return nil
}

// Info returns honesty labels for /v1/info and demos.
func (s *SupplyChain) Info() map[string]any {
	return map[string]any{
		"requirement_id": "ENG-I-008",
		"service":        "eng-i-008",
		"signing":        "ed25519",
		"sigstore":       false,
		"spdx_inspired":  true,
		"simulator":      true,
		"note":           "stdlib crypto/ed25519 + SPDX-inspired SBOM; no Sigstore/cosign",
	}
}

func validDigest(digest string) bool {
	if !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	hexPart := strings.TrimPrefix(digest, "sha256:")
	if hexPart == "" {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil
}

func sanitizeID(name string) string {
	out := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name)
	if out == "" {
		return "pkg"
	}
	return out
}
