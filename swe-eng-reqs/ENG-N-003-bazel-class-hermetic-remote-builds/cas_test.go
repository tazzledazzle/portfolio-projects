package main

import (
	"testing"
)

func TestBazelCAS_FindMissing_AllPresent(t *testing.T) {
	cas := NewBazelCAS(100)
	
	d1 := cas.Write([]byte("blob1"))
	d2 := cas.Write([]byte("blob2"))
	
	missing := cas.FindMissing([]string{d1, d2})
	
	if len(missing) != 0 {
		t.Errorf("FindMissing() = %v, want empty", missing)
	}
}

func TestBazelCAS_FindMissing_SomeMissing(t *testing.T) {
	cas := NewBazelCAS(100)
	
	d1 := cas.Write([]byte("blob1"))
	d2 := cas.Write([]byte("blob2"))
	missingDigest := "0000000000000000000000000000000000000000000000000000000000000000"
	
	missing := cas.FindMissing([]string{d1, d2, missingDigest})
	
	if len(missing) != 1 {
		t.Fatalf("FindMissing() returned %d, want 1", len(missing))
	}
	if missing[0] != missingDigest {
		t.Errorf("missing[0] = %q, want %q", missing[0], missingDigest)
	}
}

func TestBazelCAS_BatchRead(t *testing.T) {
	cas := NewBazelCAS(100)
	
	d1 := cas.Write([]byte("data1"))
	d2 := cas.Write([]byte("data2"))
	
	results := cas.BatchRead([]string{d1, d2})
	
	if len(results) != 2 {
		t.Fatalf("BatchRead() returned %d results, want 2", len(results))
	}
	if string(results[d1]) != "data1" {
		t.Errorf("results[d1] = %q, want %q", string(results[d1]), "data1")
	}
	if string(results[d2]) != "data2" {
		t.Errorf("results[d2] = %q, want %q", string(results[d2]), "data2")
	}
}

func TestBazelCAS_BatchWrite(t *testing.T) {
	cas := NewBazelCAS(100)
	
	blobs := map[string][]byte{
		"0000000000000000000000000000000000000000000000000000000000000001": []byte("one"),
		"0000000000000000000000000000000000000000000000000000000000000002": []byte("two"),
	}
	
	results := cas.BatchWrite(blobs)
	
	if len(results) != 2 {
		t.Fatalf("BatchWrite() returned %d results, want 2", len(results))
	}
	for digest, success := range results {
		if !success {
			t.Errorf("BatchWrite()[%q] = false, want true", digest)
		}
	}
}

func TestBazelCAS_DigestValidation(t *testing.T) {
	cas := NewBazelCAS(100)
	
	err := cas.ValidateDigest("abc")
	if err == nil {
		t.Error("ValidateDigest(short) = nil, want error")
	}
	
	err = cas.ValidateDigest("GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG")
	if err == nil {
		t.Error("ValidateDigest(non-hex) = nil, want error")
	}
	
	err = cas.ValidateDigest("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("ValidateDigest(valid) = %v, want nil", err)
	}
}

func TestBazelCAS_HermeticInputs(t *testing.T) {
	cas := NewBazelCAS(100)
	
	data := []byte("deterministic content")
	
	d1 := cas.Write(data)
	d2 := cas.Write(data)
	
	if d1 != d2 {
		t.Errorf("same inputs produced different digests: %q vs %q", d1, d2)
	}
}
