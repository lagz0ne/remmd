package core_test

import (
	"context"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

type mockLinkQuerier struct {
	links []*core.LinkInfo
}

func (m *mockLinkQuerier) LinksContainingSection(ctx context.Context, sectionID string) ([]*core.LinkInfo, error) {
	var result []*core.LinkInfo
	for _, l := range m.links {
		for _, id := range append(l.LeftSectionIDs, l.RightSectionIDs...) {
			if id == sectionID {
				result = append(result, l)
				break
			}
		}
	}
	return result, nil
}

func TestWalkFromSection_NoLinks_EmptyResult(t *testing.T) {
	t.Parallel()
	q := &mockLinkQuerier{links: nil}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 impacted links, got %d", len(got))
	}
}

func TestWalkFromSection_OneLink_OneImpacted(t *testing.T) {
	t.Parallel()
	q := &mockLinkQuerier{links: []*core.LinkInfo{
		{
			ID:               "link-1",
			State:            "VERIFIED",
			RelationshipType: "depends-on",
			LeftSectionIDs:   []string{"sec-a1"},
			RightSectionIDs:  []string{"sec-b1"},
		},
	}}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 impacted link, got %d", len(got))
	}
	if got[0].LinkID != "link-1" {
		t.Errorf("expected LinkID %q, got %q", "link-1", got[0].LinkID)
	}
	if got[0].ChangedSide != "left" {
		t.Errorf("expected ChangedSide %q, got %q", "left", got[0].ChangedSide)
	}
	if len(got[0].CounterpartySectionIDs) != 1 || got[0].CounterpartySectionIDs[0] != "sec-b1" {
		t.Errorf("expected counterparty [sec-b1], got %v", got[0].CounterpartySectionIDs)
	}
}

func TestWalkFromSection_RightSideChanged(t *testing.T) {
	t.Parallel()
	q := &mockLinkQuerier{links: []*core.LinkInfo{
		{
			ID:               "link-2",
			State:            "VERIFIED",
			RelationshipType: "implements",
			LeftSectionIDs:   []string{"sec-a1"},
			RightSectionIDs:  []string{"sec-b1"},
		},
	}}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 impacted link, got %d", len(got))
	}
	if got[0].ChangedSide != "right" {
		t.Errorf("expected ChangedSide %q, got %q", "right", got[0].ChangedSide)
	}
	if len(got[0].CounterpartySectionIDs) != 1 || got[0].CounterpartySectionIDs[0] != "sec-a1" {
		t.Errorf("expected counterparty [sec-a1], got %v", got[0].CounterpartySectionIDs)
	}
}

func TestWalkFromSection_MultipleLinks(t *testing.T) {
	t.Parallel()
	q := &mockLinkQuerier{links: []*core.LinkInfo{
		{
			ID:               "link-1",
			State:            "VERIFIED",
			RelationshipType: "depends-on",
			LeftSectionIDs:   []string{"sec-a1"},
			RightSectionIDs:  []string{"sec-b1"},
		},
		{
			ID:               "link-2",
			State:            "SUSPECT",
			RelationshipType: "implements",
			LeftSectionIDs:   []string{"sec-a1", "sec-a2"},
			RightSectionIDs:  []string{"sec-c1"},
		},
	}}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 impacted links, got %d", len(got))
	}
}

func TestWalkFromSection_CycleDetection(t *testing.T) {
	t.Parallel()
	// A links to B, B links to A — walk must terminate
	q := &mockLinkQuerier{links: []*core.LinkInfo{
		{
			ID:               "link-ab",
			State:            "VERIFIED",
			RelationshipType: "depends-on",
			LeftSectionIDs:   []string{"sec-a1"},
			RightSectionIDs:  []string{"sec-b1"},
		},
		{
			ID:               "link-ba",
			State:            "VERIFIED",
			RelationshipType: "depends-on",
			LeftSectionIDs:   []string{"sec-b1"},
			RightSectionIDs:  []string{"sec-a1"},
		},
	}}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should find both links but not loop infinitely
	if len(got) < 1 {
		t.Fatal("expected at least 1 impacted link from cycle")
	}
	// Should not produce duplicates
	seen := map[string]bool{}
	for _, il := range got {
		if seen[il.LinkID] {
			t.Fatalf("duplicate link %q in results — cycle not detected", il.LinkID)
		}
		seen[il.LinkID] = true
	}
}

func TestWalkFromSection_SectionOnBothSides(t *testing.T) {
	t.Parallel()
	// Edge case: section appears on both left and right of same link
	q := &mockLinkQuerier{links: []*core.LinkInfo{
		{
			ID:               "link-self",
			State:            "VERIFIED",
			RelationshipType: "references",
			LeftSectionIDs:   []string{"sec-a1"},
			RightSectionIDs:  []string{"sec-a1", "sec-b1"},
		},
	}}
	gw := core.NewGraphWalker(q)

	got, err := gw.WalkFromSection(context.Background(), "sec-a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 impacted link, got %d", len(got))
	}
}
