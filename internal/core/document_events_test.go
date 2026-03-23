package core_test

import (
	"encoding/json"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewDocumentCreatedEvent(t *testing.T) {
	t.Parallel()
	doc := core.NewDocument("Test Doc", "owner-1")
	evt := core.NewDocumentCreatedEvent(doc)

	if evt.EventType != "document.created" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "document.created")
	}
	if evt.AggregateID != doc.ID {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, doc.ID)
	}
	if evt.AggregateType != "document" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "document")
	}
	if evt.PrincipalID != doc.OwnerID {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, doc.OwnerID)
	}
	if evt.ID == "" {
		t.Fatal("expected non-empty event ID")
	}
}

func TestNewDocumentCreatedEvent_PayloadIsValidJSON(t *testing.T) {
	t.Parallel()
	doc := core.NewDocument("Test Doc", "owner-1")
	evt := core.NewDocumentCreatedEvent(doc)

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if payload["title"] != doc.Title {
		t.Errorf("payload title = %v, want %q", payload["title"], doc.Title)
	}
}

func TestNewSectionEditedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionEditedEvent("sec-1", "doc-1", "oldhash", "newhash", "user-1")

	if evt.EventType != "section.edited" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "section.edited")
	}
	if evt.AggregateID != "doc-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "doc-1")
	}
	if evt.AggregateType != "document" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "document")
	}
	if evt.PrincipalID != "user-1" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-1")
	}
}

func TestNewSectionEditedEvent_PayloadIsValidJSON(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionEditedEvent("sec-1", "doc-1", "oldhash", "newhash", "user-1")

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if payload["section_id"] != "sec-1" {
		t.Errorf("payload section_id = %v, want %q", payload["section_id"], "sec-1")
	}
	if payload["old_hash"] != "oldhash" {
		t.Errorf("payload old_hash = %v, want %q", payload["old_hash"], "oldhash")
	}
	if payload["new_hash"] != "newhash" {
		t.Errorf("payload new_hash = %v, want %q", payload["new_hash"], "newhash")
	}
}

func TestNewSectionDeletedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionDeletedEvent("sec-2", "doc-1", "obsolete", "user-2")

	if evt.EventType != "section.deleted" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "section.deleted")
	}
	if evt.AggregateID != "doc-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "doc-1")
	}
	if evt.AggregateType != "document" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "document")
	}
	if evt.PrincipalID != "user-2" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-2")
	}
}

func TestNewSectionDeletedEvent_PayloadIsValidJSON(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionDeletedEvent("sec-2", "doc-1", "obsolete", "user-2")

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if payload["section_id"] != "sec-2" {
		t.Errorf("payload section_id = %v, want %q", payload["section_id"], "sec-2")
	}
	if payload["reason"] != "obsolete" {
		t.Errorf("payload reason = %v, want %q", payload["reason"], "obsolete")
	}
}

func TestNewSectionTaggedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionTaggedEvent("sec-3", "doc-1", "important", "user-3")

	if evt.EventType != "section.tagged" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "section.tagged")
	}
	if evt.AggregateID != "doc-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "doc-1")
	}
	if evt.AggregateType != "document" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "document")
	}
	if evt.PrincipalID != "user-3" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-3")
	}
}

func TestNewSectionTaggedEvent_PayloadIsValidJSON(t *testing.T) {
	t.Parallel()
	evt := core.NewSectionTaggedEvent("sec-3", "doc-1", "important", "user-3")

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if payload["section_id"] != "sec-3" {
		t.Errorf("payload section_id = %v, want %q", payload["section_id"], "sec-3")
	}
	if payload["tag"] != "important" {
		t.Errorf("payload tag = %v, want %q", payload["tag"], "important")
	}
}
