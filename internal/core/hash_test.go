package core

import (
	"testing"
)

func TestContentHash_Consistent(t *testing.T) {
	t.Parallel()
	h1 := ContentHash("hello world")
	h2 := ContentHash("hello world")
	if h1 != h2 {
		t.Fatalf("expected consistent hash, got %q and %q", h1, h2)
	}
	// SHA-256 hex = 64 chars
	if len(h1) != 64 {
		t.Fatalf("expected 64-char hex string, got len %d: %q", len(h1), h1)
	}
}

func TestContentHash_TrailingSpacesStripped(t *testing.T) {
	t.Parallel()
	h1 := ContentHash("hello world  ")
	h2 := ContentHash("hello world")
	if h1 != h2 {
		t.Fatalf("trailing spaces should be stripped: %q != %q", h1, h2)
	}
}

func TestContentHash_CRLFNormalized(t *testing.T) {
	t.Parallel()
	h1 := ContentHash("hello\r\nworld")
	h2 := ContentHash("hello\nworld")
	if h1 != h2 {
		t.Fatalf("CRLF should normalize to LF: %q != %q", h1, h2)
	}
}

func TestContentHash_ConsecutiveBlanksCollapsed(t *testing.T) {
	t.Parallel()
	h1 := ContentHash("hello\n\n\nworld")
	h2 := ContentHash("hello\n\nworld")
	if h1 != h2 {
		t.Fatalf("consecutive blank lines should collapse: %q != %q", h1, h2)
	}
}

func TestContentHash_InternalSpacesPreserved(t *testing.T) {
	t.Parallel()
	h1 := ContentHash("hello  world")
	h2 := ContentHash("hello world")
	if h1 == h2 {
		t.Fatal("internal spaces should be preserved: hashes must differ")
	}
}

func TestContentHash_EmptyString(t *testing.T) {
	t.Parallel()
	h := ContentHash("")
	if h == "" {
		t.Fatal("empty string should produce a defined hash")
	}
	if len(h) != 64 {
		t.Fatalf("expected 64-char hex string, got len %d: %q", len(h), h)
	}
}
