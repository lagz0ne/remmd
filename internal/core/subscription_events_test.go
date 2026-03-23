package core_test

import (
	"encoding/json"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewSubscriptionCreatedEvent(t *testing.T) {
	t.Parallel()
	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	evt := core.NewSubscriptionCreatedEvent(sub)

	if evt.AggregateType != "subscription" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "subscription")
	}
	if evt.EventType != "subscription.created" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "subscription.created")
	}
	if evt.AggregateID != sub.ID {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, sub.ID)
	}
	if evt.PrincipalID != "user-1" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-1")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["subscriber_ref"] != "@a1" {
		t.Errorf("payload subscriber_ref = %v, want %q", payload["subscriber_ref"], "@a1")
	}
	if payload["tag_expression"] != "api-spec" {
		t.Errorf("payload tag_expression = %v, want %q", payload["tag_expression"], "api-spec")
	}
}

func TestNewSubscriptionFiredEvent(t *testing.T) {
	t.Parallel()
	fire := &core.SubscriptionFire{
		ID:              "fire-1",
		SubscriptionID:  "sub-1",
		TriggeringDocID: "doc-1",
		TriggeringRef:   "@c1",
		MatchedTag:      "api-spec",
	}
	evt := core.NewSubscriptionFiredEvent(fire, "system")

	if evt.AggregateType != "subscription" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "subscription")
	}
	if evt.EventType != "subscription.fired" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "subscription.fired")
	}
	if evt.AggregateID != "sub-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "sub-1")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["fire_id"] != "fire-1" {
		t.Errorf("payload fire_id = %v, want %q", payload["fire_id"], "fire-1")
	}
	if payload["triggering_doc_id"] != "doc-1" {
		t.Errorf("payload triggering_doc_id = %v, want %q", payload["triggering_doc_id"], "doc-1")
	}
	if payload["triggering_ref"] != "@c1" {
		t.Errorf("payload triggering_ref = %v, want %q", payload["triggering_ref"], "@c1")
	}
	if payload["matched_tag"] != "api-spec" {
		t.Errorf("payload matched_tag = %v, want %q", payload["matched_tag"], "api-spec")
	}
}

func TestNewSubscriptionDismissedEvent(t *testing.T) {
	t.Parallel()
	evt := core.NewSubscriptionDismissedEvent("fire-1", "sub-1", "user-1")

	if evt.AggregateType != "subscription" {
		t.Errorf("AggregateType = %q, want %q", evt.AggregateType, "subscription")
	}
	if evt.EventType != "subscription.dismissed" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "subscription.dismissed")
	}
	if evt.AggregateID != "sub-1" {
		t.Errorf("AggregateID = %q, want %q", evt.AggregateID, "sub-1")
	}
	if evt.PrincipalID != "user-1" {
		t.Errorf("PrincipalID = %q, want %q", evt.PrincipalID, "user-1")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if payload["fire_id"] != "fire-1" {
		t.Errorf("payload fire_id = %v, want %q", payload["fire_id"], "fire-1")
	}
}
