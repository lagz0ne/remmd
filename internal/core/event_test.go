package core_test

import (
	"encoding/json"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestEvent_HasRequiredFields(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"title":"hello"}`)
	evt := core.NewEvent("agg-1", "Document", "DocumentCreated", payload, "user-42")

	if evt.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if evt.AggregateID != "agg-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "agg-1")
	}
	if evt.AggregateType != "Document" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "Document")
	}
	if evt.EventType != "DocumentCreated" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "DocumentCreated")
	}
	if string(evt.Payload) != `{"title":"hello"}` {
		t.Errorf("Payload = %s, want %s", evt.Payload, `{"title":"hello"}`)
	}
	if evt.PrincipalID != "user-42" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-42")
	}
	if evt.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	// Sequence is 0 by default — assigned by the store on append.
	if evt.Sequence != 0 {
		t.Errorf("Sequence = %d, want 0 (assigned by store)", evt.Sequence)
	}
}

func TestNewEvent_GeneratesUniqueIDs(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{}`)
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		evt := core.NewEvent("agg", "T", "E", payload, "p")
		if seen[evt.ID] {
			t.Fatalf("duplicate ID at iteration %d: %s", i, evt.ID)
		}
		seen[evt.ID] = true
	}
}

func TestNewEvent_SetsTimestamp(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{}`)
	evt := core.NewEvent("agg", "T", "E", payload, "p")
	if evt.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should not be zero")
	}
}

func TestEvent_Validate_RejectsEmptyAggregateID(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{}`)
	evt := core.NewEvent("", "Document", "DocumentCreated", payload, "user-1")
	if err := evt.Validate(); err == nil {
		t.Fatal("expected error for empty AggregateID")
	}
}

func TestEvent_Validate_RejectsEmptyEventType(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{}`)
	evt := core.NewEvent("agg-1", "Document", "", payload, "user-1")
	if err := evt.Validate(); err == nil {
		t.Fatal("expected error for empty EventType")
	}
}

func TestEvent_Validate_RejectsEmptyPrincipalID(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{}`)
	evt := core.NewEvent("agg-1", "Document", "DocumentCreated", payload, "")
	if err := evt.Validate(); err == nil {
		t.Fatal("expected error for empty PrincipalID")
	}
}

func TestEvent_Validate_AcceptsValidEvent(t *testing.T) {
	t.Parallel()
	payload := json.RawMessage(`{"x":1}`)
	evt := core.NewEvent("agg-1", "Document", "DocumentCreated", payload, "user-1")
	if err := evt.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
