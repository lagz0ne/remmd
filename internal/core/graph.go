package core

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

// LinkQuerier is the interface the graph walker needs from the store.
// Defined here so the walker doesn't depend on concrete Link or store types.
type LinkQuerier interface {
	LinksContainingSection(ctx context.Context, sectionID string) ([]*LinkInfo, error)
}

// LinkInfo is the minimal link data the graph walker needs.
type LinkInfo struct {
	ID               string
	State            string
	RelationshipType string
	LeftSectionIDs   []string
	RightSectionIDs  []string
}

const maxWalkNodes = 10000

// GraphWalker traverses the link graph from a changed section to find
// all impacted counterparties.
type GraphWalker struct {
	querier LinkQuerier
}

// NewGraphWalker creates a GraphWalker backed by the given LinkQuerier.
func NewGraphWalker(q LinkQuerier) *GraphWalker {
	return &GraphWalker{querier: q}
}

// ImpactedLink describes one link affected by a section change.
type ImpactedLink struct {
	LinkID              string
	State               string
	RelationshipType    string
	ChangedSide         string   // "left" or "right"
	CounterpartySectionIDs []string // the other side's section IDs
}

// WalkFromSection finds all links containing sectionID, then for each
// link identifies the counterparty side. Uses BFS with visited-link
// deduplication to handle cycles.
func (g *GraphWalker) WalkFromSection(ctx context.Context, sectionID string) ([]ImpactedLink, error) {
	start := time.Now()

	slog.DebugContext(ctx, "walking graph from section",
		"op", "graph.walk",
		"section_id", sectionID,
	)

	var result []ImpactedLink
	visitedLinks := map[string]bool{}
	queue := []string{sectionID}
	visitedIDs := map[string]bool{sectionID: true}

	for len(queue) > 0 {
		if len(visitedIDs) > maxWalkNodes {
			return result, fmt.Errorf("graph walk exceeded %d nodes: possible cycle or pathological graph", maxWalkNodes)
		}

		id := queue[0]
		queue = queue[1:]

		links, err := g.querier.LinksContainingSection(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "graph walk failed",
				"op", "graph.walk",
				"section_id", sectionID,
				"outcome", "err",
				"error", err.Error(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return nil, err
		}

		for _, link := range links {
			if visitedLinks[link.ID] {
				continue
			}
			visitedLinks[link.ID] = true

			il := g.toLinkImpact(link, id)
			result = append(result, il)

			slog.InfoContext(ctx, "impacted link found",
				"op", "graph.walk",
				"link_id", link.ID,
				"changed_side", il.ChangedSide,
				"counterparty_count", len(il.CounterpartySectionIDs),
			)

			for _, cp := range il.CounterpartySectionIDs {
				if !visitedIDs[cp] {
					visitedIDs[cp] = true
					queue = append(queue, cp)
				}
			}
		}
	}

	slog.InfoContext(ctx, "graph walk complete",
		"op", "graph.walk",
		"section_id", sectionID,
		"links_found", len(result),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return result, nil
}

func (g *GraphWalker) toLinkImpact(link *LinkInfo, changedID string) ImpactedLink {
	onLeft := slices.Contains(link.LeftSectionIDs, changedID)
	onRight := slices.Contains(link.RightSectionIDs, changedID)

	var changedSide string
	var counterparty []string

	switch {
	case onLeft && !onRight:
		changedSide = "left"
		counterparty = link.RightSectionIDs
	case onRight && !onLeft:
		changedSide = "right"
		counterparty = link.LeftSectionIDs
	default:
		changedSide = "left" // on both sides: treat as left changed

		counterparty = link.RightSectionIDs
	}

	return ImpactedLink{
		LinkID:                 link.ID,
		State:                  link.State,
		RelationshipType:       link.RelationshipType,
		ChangedSide:            changedSide,
		CounterpartySectionIDs: counterparty,
	}
}
