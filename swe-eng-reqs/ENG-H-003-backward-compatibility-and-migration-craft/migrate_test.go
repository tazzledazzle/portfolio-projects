package main

import (
	"sync"
	"testing"
)

func TestMigrate_DualWrite_BothReadable(t *testing.T) {
	s := NewMigrateStore()
	s.EnableDualWrite(true)

	id, err := s.Put(map[string]any{"name": "alpha"})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	v1, ok := s.GetV1(id)
	if !ok {
		t.Fatal("expected GetV1 readable after dual-write Put")
	}
	v2, ok := s.GetV2(id)
	if !ok {
		t.Fatal("expected GetV2 readable after dual-write Put")
	}
	if v1["name"] != "alpha" {
		t.Fatalf("v1 name=%v", v1["name"])
	}
	if v2["display_name"] != "alpha" {
		t.Fatalf("v2 display_name=%v", v2["display_name"])
	}
}

func TestMigrate_Compat_FieldRename(t *testing.T) {
	s := NewMigrateStore()
	s.EnableDualWrite(true)

	id, err := s.Put(map[string]any{"name": "widget"})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	v1, _ := s.GetV1(id)
	v2, _ := s.GetV2(id)

	if _, hasOld := v1["name"]; !hasOld {
		t.Fatal("v1 must retain old field name")
	}
	if _, hasNew := v2["display_name"]; !hasNew {
		t.Fatal("v2 must use renamed field display_name")
	}
	if _, leaked := v2["name"]; leaked {
		t.Fatal("v2 should not keep old name field after rename")
	}
}

func TestMigrate_MidMigration_NoSplitBrain(t *testing.T) {
	s := NewMigrateStore()
	s.EnableDualWrite(true)

	const n = 50
	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id, err := s.Put(map[string]any{"name": "concurrent"})
			ids[i] = id
			errs[i] = err
		}()
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("Put[%d]: %v", i, errs[i])
		}
		if _, ok := s.GetV1(ids[i]); !ok {
			t.Fatalf("split-brain: v1 missing for %s", ids[i])
		}
		if _, ok := s.GetV2(ids[i]); !ok {
			t.Fatalf("split-brain: v2 missing for %s when dual-write on", ids[i])
		}
	}
}

func TestMigrate_CompatPass(t *testing.T) {
	s := NewMigrateStore()
	s.EnableDualWrite(true)
	id, err := s.Put(map[string]any{"name": "compat"})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	report := s.Compat(id)
	if !report.CompatPass {
		t.Fatalf("expected compat_pass, got %#v", report)
	}
	if !report.DualWrite || !report.V1Readable || !report.V2Readable {
		t.Fatalf("expected dual_write+v1_readable+v2_readable, got %#v", report)
	}
}
