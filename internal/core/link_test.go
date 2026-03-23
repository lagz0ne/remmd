package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewLink_PendingState(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-left-1"},
		[]string{"sec-right-1"},
		core.RelImplements,
		core.Rationale{Claim: "impl matches spec", Scope: "all endpoints", Exclusions: ""},
		"user-1",
	)

	if link.State != core.LinkPending {
		t.Errorf("State = %q, want %q", link.State, core.LinkPending)
	}
	if link.ID == "" {
		t.Error("ID should be non-empty")
	}
	if link.ProposerID != "user-1" {
		t.Errorf("ProposerID = %q, want %q", link.ProposerID, "user-1")
	}
	if link.RelationshipType != core.RelImplements {
		t.Errorf("RelationshipType = %q, want %q", link.RelationshipType, core.RelImplements)
	}
}

func TestNewLink_HasLeftAndRightSectionIDs(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-l1", "sec-l2"},
		[]string{"sec-r1"},
		core.RelAgreesWith,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	if len(link.LeftSectionIDs) != 2 {
		t.Errorf("LeftSectionIDs len = %d, want 2", len(link.LeftSectionIDs))
	}
	if len(link.RightSectionIDs) != 1 {
		t.Errorf("RightSectionIDs len = %d, want 1", len(link.RightSectionIDs))
	}
}

func TestLink_ContainsSection_Left(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-left"},
		[]string{"sec-right"},
		core.RelTests,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	if !link.ContainsSection("sec-left") {
		t.Error("ContainsSection should return true for left section ID")
	}
}

func TestLink_ContainsSection_Right(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-left"},
		[]string{"sec-right"},
		core.RelTests,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	if !link.ContainsSection("sec-right") {
		t.Error("ContainsSection should return true for right section ID")
	}
}

func TestLink_ContainsSection_NotFound(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-left"},
		[]string{"sec-right"},
		core.RelTests,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	if link.ContainsSection("sec-other") {
		t.Error("ContainsSection should return false for unrelated section ID")
	}
}

func TestNewLink_DefaultIntervention(t *testing.T) {
	t.Parallel()

	link := core.NewLink(
		[]string{"sec-left"},
		[]string{"sec-right"},
		core.RelAgreesWith,
		core.Rationale{Claim: "c", Scope: "s", Exclusions: ""},
		"user-1",
	)

	if link.LeftIntervention != core.InterventionNotify {
		t.Errorf("LeftIntervention = %q, want %q", link.LeftIntervention, core.InterventionNotify)
	}
	if link.RightIntervention != core.InterventionNotify {
		t.Errorf("RightIntervention = %q, want %q", link.RightIntervention, core.InterventionNotify)
	}
}
