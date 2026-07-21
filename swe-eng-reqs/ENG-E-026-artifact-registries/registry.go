package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ErrManifestConflict is returned when a digest already stores different manifest bytes.
var ErrManifestConflict = errors.New("manifest conflict: digest immutable with different content")

// ErrUnsafeName is returned for repository or tag names containing path traversal.
var ErrUnsafeName = errors.New("unsafe repository or tag name")

// ErrInvalidDigest is returned when a digest fails OCI sha256: grammar.
var ErrInvalidDigest = errors.New("invalid digest")

// Registry is an OCI-inspired in-memory tag→digest registry (not Distribution Spec conformance).
type Registry struct {
	mu        sync.RWMutex
	tags      map[string]string // "name:tag" -> digest
	manifests map[string][]byte // digest -> bytes
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		tags:      make(map[string]string),
		manifests: make(map[string][]byte),
	}
}

func digestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(d string) bool {
	if strings.Contains(d, "..") {
		return false
	}
	if !strings.HasPrefix(d, "sha256:") {
		return false
	}
	hexPart := strings.TrimPrefix(d, "sha256:")
	if len(hexPart) != 64 {
		return false
	}
	for _, c := range hexPart {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func validName(s string) bool {
	if s == "" || strings.Contains(s, "..") || strings.Contains(s, "//") {
		return false
	}
	return true
}

// PutManifest stores manifest bytes under their content digest.
func (r *Registry) PutManifest(data []byte) (string, error) {
	digest := digestSHA256(data)
	if err := r.putManifestAt(digest, data); err != nil {
		return "", err
	}
	return digest, nil
}

func (r *Registry) putManifestAt(digest string, data []byte) error {
	if !validDigest(digest) {
		return ErrInvalidDigest
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)

	if existing, ok := r.manifests[digest]; ok {
		if !bytes.Equal(existing, cp) {
			return ErrManifestConflict
		}
		return nil
	}
	r.manifests[digest] = cp
	return nil
}

// GetManifest returns a copy of manifest bytes for digest.
func (r *Registry) GetManifest(digest string) ([]byte, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.manifests[digest]
	if !ok {
		return nil, false
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, true
}

// PutTag points name:tag at digest (mutable pointer; does not change manifest bytes).
func (r *Registry) PutTag(name, tag, digest string) error {
	if !validName(name) || !validName(tag) {
		return ErrUnsafeName
	}
	if !validDigest(digest) {
		return ErrInvalidDigest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.manifests[digest]; !ok {
		return fmt.Errorf("unknown digest: %s", digest)
	}
	r.tags[name+":"+tag] = digest
	return nil
}

// Resolve returns the digest currently pointed to by name:tag.
func (r *Registry) Resolve(name, tag string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.tags[name+":"+tag]
	return d, ok
}

// TagCount returns the number of tags.
func (r *Registry) TagCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tags)
}

// ManifestCount returns the number of manifests.
func (r *Registry) ManifestCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.manifests)
}
