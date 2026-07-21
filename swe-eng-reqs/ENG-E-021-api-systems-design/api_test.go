package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestAPI_Authorize_DefaultDeny(t *testing.T) {
	eng := NewAPIEngine(3, "resources:read")
	ok, err := eng.Authorize(nil)
	if ok || err == nil {
		t.Fatalf("expected deny without scopes, ok=%v err=%v", ok, err)
	}
	ok, err = eng.Authorize([]string{"other:scope"})
	if ok || err == nil {
		t.Fatalf("expected deny for wrong scope")
	}
}

func TestAPI_Authorize_AllowWithScope(t *testing.T) {
	eng := NewAPIEngine(3, "resources:read")
	ok, err := eng.Authorize([]string{"resources:read"})
	if !ok || err != nil {
		t.Fatalf("expected allow: ok=%v err=%v", ok, err)
	}
}

func TestAPI_RateLimit_ExceedsQuota(t *testing.T) {
	eng := NewAPIEngine(2, "resources:read")
	subject := "alice"
	if !eng.Allow(subject) || !eng.Allow(subject) {
		t.Fatal("first two Allows should succeed")
	}
	if eng.Allow(subject) {
		t.Fatal("third Allow should rate_limit")
	}
}

func TestAPI_Compat_V1V2Paths(t *testing.T) {
	eng := NewAPIEngine(10, "resources:read")
	c := eng.Compat()
	if !c.V1OK || !c.V2OK || !c.Pass {
		t.Fatalf("compat should pass for versioned shapes: %#v", c)
	}
}

func TestAPI_OpenAPI_DocumentPresent(t *testing.T) {
	eng := NewAPIEngine(10, "resources:read")
	doc, err := eng.OpenAPIDoc()
	if err != nil {
		t.Fatalf("OpenAPIDoc: %v", err)
	}
	if doc == "" || !containsOpenAPI(doc) {
		t.Fatalf("expected openapi document content, got %q", doc)
	}
}

func TestAPI_ConcurrentAllow(t *testing.T) {
	eng := NewAPIEngine(1000, "resources:read")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sub := fmt.Sprintf("u-%d", i%5)
			_ = eng.Allow(sub)
			_, _ = eng.Authorize([]string{"resources:read"})
		}(i)
	}
	wg.Wait()
}

func containsOpenAPI(s string) bool {
	return strings.Contains(s, "openapi:") || strings.Contains(s, `"openapi"`)
}
