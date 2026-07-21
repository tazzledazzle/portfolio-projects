package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

var ErrQuorumUnavailable = errors.New("write quorum unavailable")

type nodeStore struct {
	healthy bool
	blobs   map[string][]byte
}

type Durability struct {
	Replicas     int    `json:"replicas"`
	Checksum     string `json:"checksum"`
	HealthyNodes int    `json:"healthy_nodes"`
	Quorum       int    `json:"quorum"`
	Durable      bool   `json:"durable"`
}

type Facade struct {
	mu     sync.RWMutex
	nodes  []nodeStore
	quorum int
}

func NewFacade(nodes, quorum int) *Facade {
	if nodes < 1 {
		nodes = 1
	}
	if quorum < 1 {
		quorum = 1
	}
	stores := make([]nodeStore, nodes)
	for i := range stores {
		stores[i] = nodeStore{healthy: true, blobs: make(map[string][]byte)}
	}
	return &Facade{nodes: stores, quorum: quorum}
}

func (f *Facade) Put(blob []byte) (string, error) {
	digest := digestSHA256(blob)
	f.mu.Lock()
	defer f.mu.Unlock()

	healthy := 0
	for i := range f.nodes {
		if f.nodes[i].healthy {
			healthy++
		}
	}
	if healthy < f.quorum {
		return "", ErrQuorumUnavailable
	}
	for i := range f.nodes {
		if f.nodes[i].healthy {
			f.nodes[i].blobs[digest] = append([]byte(nil), blob...)
		}
	}
	return digest, nil
}

func (f *Facade) Get(digest string) ([]byte, bool) {
	if !validSHA256Digest(digest) {
		return nil, false
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	for i := range f.nodes {
		if !f.nodes[i].healthy {
			continue
		}
		if blob, ok := f.nodes[i].blobs[digest]; ok {
			return append([]byte(nil), blob...), true
		}
	}
	return nil, false
}

func (f *Facade) Durability(digest string) (Durability, bool) {
	if !validSHA256Digest(digest) {
		return Durability{}, false
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	d := Durability{Checksum: digest, Quorum: f.quorum}
	for i := range f.nodes {
		if f.nodes[i].healthy {
			d.HealthyNodes++
		}
		if _, ok := f.nodes[i].blobs[digest]; ok {
			d.Replicas++
		}
	}
	d.Durable = d.Replicas >= f.quorum && d.HealthyNodes >= f.quorum
	return d, d.Replicas > 0
}

func (f *Facade) SetNodeHealthy(index int, healthy bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if index < 0 || index >= len(f.nodes) {
		return false
	}
	f.nodes[index].healthy = healthy
	return true
}

func (f *Facade) NodeCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.nodes)
}

func digestSHA256(blob []byte) string {
	sum := sha256.Sum256(blob)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validSHA256Digest(digest string) bool {
	const prefix = "sha256:"
	if len(digest) != len(prefix)+64 || digest[:len(prefix)] != prefix {
		return false
	}
	_, err := hex.DecodeString(digest[len(prefix):])
	return err == nil
}
