package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrUnsafeID    = errors.New("unsafe item id")
	ErrEmptyName   = errors.New("name required")
	ErrNotFound    = errors.New("item not found")
	ErrDualWriteOff = errors.New("dual-write disabled")
)

// ItemV1 is the legacy API shape (field: name).
type ItemV1 struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ItemV2 is the migrated API shape (field rename: name → display_name).
type ItemV2 struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// CompatReport is server-computed migration compatibility evidence.
type CompatReport struct {
	DualWrite   bool   `json:"dual_write"`
	V1Readable  bool   `json:"v1_readable"`
	V2Readable  bool   `json:"v2_readable"`
	CompatPass  bool   `json:"compat_pass"`
	ItemID      string `json:"item_id,omitempty"`
	FieldRename string `json:"field_rename,omitempty"`
}

// MigrateStore dual-writes v1+v2 under a single mutex (T-5-20).
type MigrateStore struct {
	mu        sync.Mutex
	dualWrite bool
	v1        map[string]ItemV1
	v2        map[string]ItemV2
	counter   int
}

func NewMigrateStore() *MigrateStore {
	return &MigrateStore{
		v1: make(map[string]ItemV1),
		v2: make(map[string]ItemV2),
	}
}

func (s *MigrateStore) EnableDualWrite(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dualWrite = on
}

func (s *MigrateStore) DualWriteEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dualWrite
}

func safeID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.ContainsAny(id, `/\`)
}

// Put writes both versions in one critical section when dual-write is enabled.
func (s *MigrateStore) Put(fields map[string]any) (string, error) {
	name, _ := fields["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		if dn, ok := fields["display_name"].(string); ok {
			name = strings.TrimSpace(dn)
		}
	}
	if name == "" {
		return "", ErrEmptyName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	id := fmt.Sprintf("item-%d", s.counter)
	v1 := ItemV1{ID: id, Name: name}
	s.v1[id] = v1

	if s.dualWrite {
		s.v2[id] = ItemV2{ID: id, DisplayName: name}
	}
	return id, nil
}

func (s *MigrateStore) GetV1(id string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.v1[id]
	if !ok {
		return nil, false
	}
	return map[string]any{"id": item.ID, "name": item.Name}, true
}

func (s *MigrateStore) GetV2(id string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.v2[id]
	if !ok {
		return nil, false
	}
	return map[string]any{"id": item.ID, "display_name": item.DisplayName}, true
}

func (s *MigrateStore) Compat(id string) CompatReport {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, v1ok := s.v1[id]
	_, v2ok := s.v2[id]
	report := CompatReport{
		DualWrite:   s.dualWrite,
		V1Readable:  v1ok,
		V2Readable:  v2ok,
		ItemID:      id,
		FieldRename: "name→display_name",
	}
	report.CompatPass = s.dualWrite && v1ok && v2ok
	return report
}

// SeedDemo enables dual-write, puts one item, returns live compat proof.
func (s *MigrateStore) SeedDemo() (CompatReport, error) {
	s.EnableDualWrite(true)
	id, err := s.Put(map[string]any{"name": "demo-item"})
	if err != nil {
		return CompatReport{}, err
	}
	return s.Compat(id), nil
}
