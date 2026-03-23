package core_test

import (
	"encoding/json"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewLinkProposedEvent(t *testing.T) {
	t.Parallel()
	link := core.NewLink(
		[]string{"sec-left-1"}, []string{"sec-right-1"},
		core.RelImplements,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	evt := core.NewLinkProposedEvent(link)
	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "link.proposed" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "link.proposed")
	}
	if evt.AggregateID != link.ID {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, link.ID)
	}
	if evt.PrincipalID != "user-1" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-1")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["relationship_type"] != "implements" {
		t.Errorf("payload relationship_type = %v, want %q", payload["relationship_type"], "implements")
	}
}

func TestNewEndpointApprovedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewEndpointApprovedEvent("link-1", "user-2", "right")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "endpoint.approved" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "endpoint.approved")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["side"] != "right" {
		t.Errorf("payload side = %v, want %q", payload["side"], "right")
	}
}

func TestNewLinkAlignedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewLinkAlignedEvent("link-1", "user-1")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "link.aligned" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "link.aligned")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
}

func TestNewLinkStaleEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewLinkStaleEvent("link-1", "system", "upstream content changed")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "link.stale" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "link.stale")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["reason"] != "upstream content changed" {
		t.Errorf("payload reason = %v, want %q", payload["reason"], "upstream content changed")
	}
}

func TestNewLinkBrokenEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewLinkBrokenEvent("link-1", "user-1", "section deleted")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "link.broken" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "link.broken")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["reason"] != "section deleted" {
		t.Errorf("payload reason = %v, want %q", payload["reason"], "section deleted")
	}
}

func TestNewLinkArchivedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewLinkArchivedEvent("link-1", "user-1")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "link.archived" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "link.archived")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
}

func TestNewCommentAddedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewCommentAddedEvent("link-1", "user-1", "LGTM")

	if evt.AggregateType != "link" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "link")
	}
	if evt.EventType != "comment.added" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "comment.added")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["body"] != "LGTM" {
		t.Errorf("payload body = %v, want %q", payload["body"], "LGTM")
	}
}
