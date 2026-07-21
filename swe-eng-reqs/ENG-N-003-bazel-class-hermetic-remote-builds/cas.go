package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

type BazelCAS struct {
	blobs    map[string][]byte
	capacity int
	mu       sync.RWMutex
}

func NewBazelCAS(capacity int) *BazelCAS {
	return &BazelCAS{
		blobs:    make(map[string][]byte),
		capacity: capacity,
	}
}

func (b *BazelCAS) Write(blob []byte) string {
	digest := computeDigest(blob)
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.blobs[digest] = blob
	return digest
}

func (b *BazelCAS) Read(digest string) ([]byte, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	blob, ok := b.blobs[digest]
	return blob, ok
}

func (b *BazelCAS) FindMissing(digests []string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var missing []string
	for _, digest := range digests {
		if _, ok := b.blobs[digest]; !ok {
			missing = append(missing, digest)
		}
	}
	return missing
}

func (b *BazelCAS) BatchRead(digests []string) map[string][]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	results := make(map[string][]byte)
	for _, digest := range digests {
		if blob, ok := b.blobs[digest]; ok {
			results[digest] = blob
		}
	}
	return results
}

func (b *BazelCAS) BatchWrite(blobs map[string][]byte) map[string]bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	results := make(map[string]bool)
	for digest, blob := range blobs {
		b.blobs[digest] = blob
		results[digest] = true
	}
	return results
}

func (b *BazelCAS) ValidateDigest(digest string) error {
	if len(digest) != 64 {
		return errors.New("digest must be 64 hex characters")
	}
	
	for _, c := range digest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return errors.New("digest must contain only lowercase hex characters")
		}
	}
	
	return nil
}

func computeDigest(blob []byte) string {
	hash := sha256.Sum256(blob)
	return fmt.Sprintf("%x", hash)
}
