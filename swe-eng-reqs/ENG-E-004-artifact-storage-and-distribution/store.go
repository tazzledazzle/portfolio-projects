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

// ErrDigestConflict is returned when a digest already stores different bytes.
var ErrDigestConflict = errors.New("digest conflict: immutable blob already exists with different content")

// BlobStore is an in-memory content-addressed blob store (sha256: digests).
type BlobStore struct {
	mu       sync.Mutex
	blobs    map[string][]byte
	metadata map[string]map[string]string
}

// NewBlobStore creates an empty BlobStore.
func NewBlobStore() *BlobStore {
	return &BlobStore{
		blobs:    make(map[string][]byte),
		metadata: make(map[string]map[string]string),
	}
}

// DigestSHA256 returns an OCI-form digest sha256:<64hex>.
func DigestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ValidDigest reports whether d matches sha256:<64 lowercase hex> with no traversal.
func ValidDigest(d string) bool {
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

// Put stores data under its content digest. Idempotent when bytes match.
func (s *BlobStore) Put(data []byte, meta map[string]string) (string, error) {
	digest := DigestSHA256(data)
	if err := s.putAt(digest, data, meta); err != nil {
		return "", err
	}
	return digest, nil
}

// putAt stores data under digest, rejecting unequal overwrite (T-3-01).
func (s *BlobStore) putAt(digest string, data []byte, meta map[string]string) error {
	if !ValidDigest(digest) {
		return fmt.Errorf("invalid digest: %s", digest)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)

	if existing, ok := s.blobs[digest]; ok {
		if !bytes.Equal(existing, cp) {
			return ErrDigestConflict
		}
		if meta != nil {
			s.metadata[digest] = copyMeta(meta)
		}
		return nil
	}

	s.blobs[digest] = cp
	if meta != nil {
		s.metadata[digest] = copyMeta(meta)
	} else {
		s.metadata[digest] = make(map[string]string)
	}
	return nil
}

// Get returns a copy of blob bytes for digest.
func (s *BlobStore) Get(digest string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.blobs[digest]
	if !ok {
		return nil, false
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, true
}

// Head returns a copy of metadata for digest.
func (s *BlobStore) Head(digest string) (map[string]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.metadata[digest]
	if !ok {
		if _, exists := s.blobs[digest]; !exists {
			return nil, false
		}
		return map[string]string{}, true
	}
	return copyMeta(m), true
}

// Count returns the number of stored blobs.
func (s *BlobStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.blobs)
}

// MetadataKeys returns a de-duplicated list of metadata keys across all blobs.
func (s *BlobStore) MetadataKeys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen := make(map[string]struct{})
	var keys []string
	for _, m := range s.metadata {
		for k := range m {
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				keys = append(keys, k)
			}
		}
	}
	return keys
}

func copyMeta(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
