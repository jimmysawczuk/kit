package requestid

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestRandomPrefix(t *testing.T) {
	p1 := RandomPrefix(defaultLength)
	p2 := RandomPrefix(defaultLength)

	if p1 == p2 {
		t.Fatal("expected different prefixes, got equal")
	}
	if p1 == "" {
		t.Fatal("expected non-empty prefix")
	}
}

func TestHostnamePrefix(t *testing.T) {
	p := HostnamePrefix(defaultLength)
	if !strings.Contains(p, "/") {
		t.Fatalf("expected hostname/random format, got %q", p)
	}
	parts := strings.SplitN(p, "/", 2)
	if parts[0] == "" {
		t.Fatal("expected non-empty hostname part")
	}
	if parts[1] == "" {
		t.Fatal("expected non-empty random part")
	}
}

func TestGenerator_Next_GeneratesSequentialIDs(t *testing.T) {
	g := &Generator{Prefix: "test"}

	r1, _ := http.NewRequest(http.MethodGet, "/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/", nil)

	id1 := g.Next(r1)
	id2 := g.Next(r2)

	if id1 == id2 {
		t.Fatalf("expected different IDs, got %q twice", id1)
	}
	if !strings.HasPrefix(id1, "test-") {
		t.Fatalf("expected prefix 'test-', got %q", id1)
	}
}

func TestGenerator_Next_RespectsIncomingHeader(t *testing.T) {
	g := &Generator{Prefix: "test"}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-Id", "incoming-id-123")

	id := g.Next(r)
	if id != "incoming-id-123" {
		t.Fatalf("expected incoming-id-123, got %q", id)
	}
}

func TestGenerator_SetAndGet(t *testing.T) {
	g := &Generator{Prefix: "test"}
	ctx := context.Background()

	if got := g.Get(ctx); got != "" {
		t.Fatalf("expected empty string from empty context, got %q", got)
	}

	ctx = g.Set(ctx, "req-001")
	if got := g.Get(ctx); got != "req-001" {
		t.Fatalf("expected req-001, got %q", got)
	}
}

func TestGenerator_SetDoesNotMutateParent(t *testing.T) {
	g := &Generator{Prefix: "test"}
	parent := context.Background()
	_ = g.Set(parent, "req-001")

	if got := g.Get(parent); got != "" {
		t.Fatalf("parent context was mutated, got %q", got)
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-Id", "pkg-level-id")

	if id := Next(r); id != "pkg-level-id" {
		t.Fatalf("expected pkg-level-id, got %q", id)
	}

	ctx := Set(context.Background(), "pkg-test")
	if got := Get(ctx); got != "pkg-test" {
		t.Fatalf("expected pkg-test, got %q", got)
	}
}
