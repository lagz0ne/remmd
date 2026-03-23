package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewSubscription_AutoID(t *testing.T) {
	t.Parallel()
	sub := core.NewSubscription("@a1", "api-spec", "user-1")

	if sub.ID == "" {
		t.Error("ID should be auto-generated, got empty")
	}
	if sub.Status != core.SubActive {
		t.Errorf("Status = %q, want %q", sub.Status, core.SubActive)
	}
}

func TestNewSubscription_FieldsSet(t *testing.T) {
	t.Parallel()
	sub := core.NewSubscription("@b2", "design-doc", "user-42")

	if sub.SubscriberRef != "@b2" {
		t.Errorf("SubscriberRef = %q, want %q", sub.SubscriberRef, "@b2")
	}
	if sub.TagExpression != "design-doc" {
		t.Errorf("TagExpression = %q, want %q", sub.TagExpression, "design-doc")
	}
	if sub.CreatedBy != "user-42" {
		t.Errorf("CreatedBy = %q, want %q", sub.CreatedBy, "user-42")
	}
	if sub.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
