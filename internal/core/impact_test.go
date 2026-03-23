package core_test

import (
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewBlastRadius_EmptyInput(t *testing.T) {
	t.Parallel()
	br := core.NewBlastRadius("sec-a1", nil)
	if br == nil {
		t.Fatal("NewBlastRadius returned nil")
	}
	if len(br.Groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(br.Groups))
	}
	if br.ChangedSectionID != "sec-a1" {
		t.Fatalf("expected ChangedSectionID %q, got %q", "sec-a1", br.ChangedSectionID)
	}
}

func TestNewBlastRadius_GroupsByCounterparty(t *testing.T) {
	t.Parallel()
	links := []core.ImpactedLink{
		{
			LinkID:                 "link-1",
			State:                  "VERIFIED",
			RelationshipType:       "depends-on",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-b1"},
		},
		{
			LinkID:                 "link-2",
			State:                  "VERIFIED",
			RelationshipType:       "implements",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-b1"},
		},
	}
	br := core.NewBlastRadius("sec-a1", links)
	if len(br.Groups) != 1 {
		t.Fatalf("expected 1 group (same counterparty), got %d", len(br.Groups))
	}
	if len(br.Groups[0].Links) != 2 {
		t.Fatalf("expected 2 links in group, got %d", len(br.Groups[0].Links))
	}
}

func TestNewBlastRadius_DifferentCounterparties(t *testing.T) {
	t.Parallel()
	links := []core.ImpactedLink{
		{
			LinkID:                 "link-1",
			RelationshipType:       "depends-on",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-b1"},
		},
		{
			LinkID:                 "link-2",
			RelationshipType:       "depends-on",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-c1"},
		},
	}
	br := core.NewBlastRadius("sec-a1", links)
	if len(br.Groups) != 2 {
		t.Fatalf("expected 2 groups (different counterparties), got %d", len(br.Groups))
	}
}

func TestNewBlastRadius_GroupsSortedByRelationshipType(t *testing.T) {
	t.Parallel()
	links := []core.ImpactedLink{
		{
			LinkID:                 "link-1",
			RelationshipType:       "implements",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-c1"},
		},
		{
			LinkID:                 "link-2",
			RelationshipType:       "depends-on",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-b1"},
		},
	}
	br := core.NewBlastRadius("sec-a1", links)
	if len(br.Groups) < 2 {
		t.Fatalf("expected at least 2 groups, got %d", len(br.Groups))
	}
	// Groups should be sorted: depends-on before implements
	firstType := br.Groups[0].Links[0].RelationshipType
	secondType := br.Groups[1].Links[0].RelationshipType
	if firstType > secondType {
		t.Errorf("groups not sorted: %q before %q", firstType, secondType)
	}
}

func TestBlastRadius_Summary_Empty(t *testing.T) {
	t.Parallel()
	br := core.NewBlastRadius("sec-a1", nil)
	summary := br.Summary()
	if !strings.Contains(summary, "sec-a1") {
		t.Errorf("summary should mention changed section ID, got: %s", summary)
	}
	if !strings.Contains(summary, "no impact") {
		t.Errorf("summary should indicate no impact, got: %s", summary)
	}
}

func TestBlastRadius_Summary_WithLinks(t *testing.T) {
	t.Parallel()
	links := []core.ImpactedLink{
		{
			LinkID:                 "link-1",
			RelationshipType:       "depends-on",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-b1"},
		},
		{
			LinkID:                 "link-2",
			RelationshipType:       "implements",
			ChangedSide:            "left",
			CounterpartySectionIDs: []string{"sec-c1"},
		},
	}
	br := core.NewBlastRadius("sec-a1", links)
	summary := br.Summary()
	if !strings.Contains(summary, "sec-a1") {
		t.Errorf("summary should mention changed section ID, got: %s", summary)
	}
	if !strings.Contains(summary, "2") {
		t.Errorf("summary should mention count, got: %s", summary)
	}
}
