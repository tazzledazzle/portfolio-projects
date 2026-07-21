package main

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type CASCache struct {
	blobs    map[string]map[string][]byte
	order    []cacheKey
	capacity int
	hits     int
	misses   int
	mu       sync.Mutex
}

type cacheKey struct {
	namespace string
	digest    string
}

func NewCASCache(capacity int) *CASCache {
	return &CASCache{
		blobs:    make(map[string]map[string][]byte),
		order:    make([]cacheKey, 0),
		capacity: capacity,
	}
}

func (c *CASCache) Get(namespace, digest string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	nsBlobs, ok := c.blobs[namespace]
	if !ok {
		c.misses++
		return nil, false
	}

	blob, ok := nsBlobs[digest]
	if !ok {
		c.misses++
		return nil, false
	}

	c.hits++
	c.moveToFront(cacheKey{namespace, digest})
	return blob, true
}

func (c *CASCache) Put(namespace string, blob []byte) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	digest := computeDigest(blob)

	if _, ok := c.blobs[namespace]; !ok {
		c.blobs[namespace] = make(map[string][]byte)
	}

	if _, exists := c.blobs[namespace][digest]; exists {
		c.moveToFront(cacheKey{namespace, digest})
		return digest
	}

	c.blobs[namespace][digest] = blob
	c.order = append([]cacheKey{{namespace, digest}}, c.order...)

	for c.totalItems() > c.capacity {
		c.evictOldest()
	}

	return digest
}

func (c *CASCache) HitRate() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}

func (c *CASCache) Hits() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits
}

func (c *CASCache) Misses() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.misses
}

func (c *CASCache) moveToFront(key cacheKey) {
	for i, k := range c.order {
		if k.namespace == key.namespace && k.digest == key.digest {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append([]cacheKey{key}, c.order...)
			return
		}
	}
}

func (c *CASCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}

	oldest := c.order[len(c.order)-1]
	c.order = c.order[:len(c.order)-1]

	if nsBlobs, ok := c.blobs[oldest.namespace]; ok {
		delete(nsBlobs, oldest.digest)
		if len(nsBlobs) == 0 {
			delete(c.blobs, oldest.namespace)
		}
	}
}

func (c *CASCache) totalItems() int {
	count := 0
	for _, nsBlobs := range c.blobs {
		count += len(nsBlobs)
	}
	return count
}

func computeDigest(blob []byte) string {
	hash := sha256.Sum256(blob)
	return fmt.Sprintf("%x", hash)
}
