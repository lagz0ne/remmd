package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewThread_Empty(t *testing.T) {
	t.Parallel()
	thread := core.NewThread()
	entries := thread.Entries()
	if len(entries) != 0 {
		t.Errorf("new thread should have 0 entries, got %d", len(entries))
	}
}

func TestThread_AddComment(t *testing.T) {
	t.Parallel()
	thread := core.NewThread()
	entry := thread.AddComment("user-1", "This looks good")

	if entry.Type != core.EntryComment {
		t.Errorf("Type = %q, want %q", entry.Type, core.EntryComment)
	}
	if entry.PrincipalID != "user-1" {
		t.Errorf("PrincipalID = %q, want %q", entry.PrincipalID, "user-1")
	}
	if entry.Body != "This looks good" {
		t.Errorf("Body = %q, want %q", entry.Body, "This looks good")
	}
	if entry.ID == "" {
		t.Error("entry ID should be non-empty")
	}

	entries := thread.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestThread_AddSystemEvent(t *testing.T) {
	t.Parallel()
	thread := core.NewThread()
	entry := thread.AddSystemEvent("link.stale", "Upstream content changed")

	if entry.Type != core.EntrySystem {
		t.Errorf("Type = %q, want %q", entry.Type, core.EntrySystem)
	}
	if entry.Body != "Upstream content changed" {
		t.Errorf("Body = %q, want %q", entry.Body, "Upstream content changed")
	}
}

func TestThread_Entries_ChronologicalOrder(t *testing.T) {
	t.Parallel()
	thread := core.NewThread()
	thread.AddComment("user-1", "first")
	thread.AddSystemEvent("state.change", "second")
	thread.AddComment("user-2", "third")

	entries := thread.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Body != "first" {
		t.Errorf("entries[0].Body = %q, want %q", entries[0].Body, "first")
	}
	if entries[1].Body != "second" {
		t.Errorf("entries[1].Body = %q, want %q", entries[1].Body, "second")
	}
	if entries[2].Body != "third" {
		t.Errorf("entries[2].Body = %q, want %q", entries[2].Body, "third")
	}
}

func TestThread_Entries_ReturnsCopy(t *testing.T) {
	t.Parallel()
	thread := core.NewThread()
	thread.AddComment("user-1", "original")

	entries := thread.Entries()
	entries[0].Body = "mutated"

	// Original should be unaffected
	fresh := thread.Entries()
	if fresh[0].Body != "original" {
		t.Errorf("mutation leaked: Body = %q, want %q", fresh[0].Body, "original")
	}
}
