package main

import (
	"sync"
	"testing"
)

func TestCASCache_PutGet_Hit(t *testing.T) {
	cache := NewCASCache(10)
	blob := []byte("hello world")
	
	digest := cache.Put("ns1", blob)
	
	got, ok := cache.Get("ns1", digest)
	if !ok {
		t.Error("Get() returned false, want true")
	}
	if string(got) != string(blob) {
		t.Errorf("Get() = %q, want %q", string(got), string(blob))
	}
}

func TestCASCache_Get_Miss(t *testing.T) {
	cache := NewCASCache(10)
	
	got, ok := cache.Get("ns1", "nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true, want false")
	}
	if got != nil {
		t.Errorf("Get(nonexistent) = %v, want nil", got)
	}
}

func TestCASCache_HitRate(t *testing.T) {
	cache := NewCASCache(10)
	
	digest := cache.Put("ns1", []byte("data"))
	
	cache.Get("ns1", digest)
	cache.Get("ns1", digest)
	cache.Get("ns1", digest)
	cache.Get("ns1", "miss")
	
	rate := cache.HitRate()
	if rate != 0.75 {
		t.Errorf("HitRate() = %v, want 0.75", rate)
	}
}

func TestCASCache_NamespaceIsolation(t *testing.T) {
	cache := NewCASCache(10)
	
	digest := cache.Put("ns1", []byte("secret"))
	
	got, ok := cache.Get("ns2", digest)
	if ok {
		t.Error("Get from different namespace returned true, want false")
	}
	if got != nil {
		t.Errorf("Get from different namespace = %v, want nil", got)
	}
}

func TestCASCache_LRUEviction(t *testing.T) {
	cache := NewCASCache(2)
	
	d1 := cache.Put("ns1", []byte("first"))
	d2 := cache.Put("ns1", []byte("second"))
	_ = cache.Put("ns1", []byte("third"))
	
	_, ok1 := cache.Get("ns1", d1)
	_, ok2 := cache.Get("ns1", d2)
	
	if ok1 {
		t.Error("first blob should be evicted")
	}
	if !ok2 {
		t.Error("second blob should still exist (was accessed implicitly)")
	}
}

func TestCASCache_DigestFormat(t *testing.T) {
	cache := NewCASCache(10)
	
	digest := cache.Put("ns1", []byte("test"))
	
	if len(digest) != 64 {
		t.Errorf("digest length = %d, want 64", len(digest))
	}
	
	for _, c := range digest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("digest contains non-hex char: %c", c)
		}
	}
}

func TestCASCache_ConcurrentAccess(t *testing.T) {
	cache := NewCASCache(100)
	var wg sync.WaitGroup
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			data := []byte{byte(n)}
			digest := cache.Put("ns1", data)
			cache.Get("ns1", digest)
		}(i)
	}
	
	wg.Wait()
	
	rate := cache.HitRate()
	if rate != rate {
		t.Error("HitRate is NaN after concurrent access")
	}
}
