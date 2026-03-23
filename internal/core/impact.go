package core

import (
	"fmt"
	"sort"
	"strings"
)

// BlastRadius groups impacted links by counterparty for review.
type BlastRadius struct {
	ChangedSectionID string
	Groups           []ImpactGroup
}

// ImpactGroup clusters links sharing the same counterparty section IDs.
type ImpactGroup struct {
	CounterpartySectionIDs []string
	Links                  []ImpactedLink
}

// NewBlastRadius groups the impacted links by their counterparty section IDs,
// then sorts groups by the relationship type of their first link.
func NewBlastRadius(changedSectionID string, links []ImpactedLink) *BlastRadius {
	if len(links) == 0 {
		return &BlastRadius{ChangedSectionID: changedSectionID}
	}

	groupMap := map[string]*ImpactGroup{}
	var order []string
	for _, l := range links {
		key := counterpartyKey(l.CounterpartySectionIDs)
		g, ok := groupMap[key]
		if !ok {
			g = &ImpactGroup{CounterpartySectionIDs: l.CounterpartySectionIDs}
			groupMap[key] = g
			order = append(order, key)
		}
		g.Links = append(g.Links, l)
	}

	groups := make([]ImpactGroup, 0, len(groupMap))
	for _, key := range order {
		groups = append(groups, *groupMap[key])
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Links[0].RelationshipType < groups[j].Links[0].RelationshipType
	})

	return &BlastRadius{
		ChangedSectionID: changedSectionID,
		Groups:           groups,
	}
}

// Summary returns a human-readable summary of the blast radius.
func (b *BlastRadius) Summary() string {
	total := 0
	for _, g := range b.Groups {
		total += len(g.Links)
	}
	if total == 0 {
		return fmt.Sprintf("%s: no impact detected", b.ChangedSectionID)
	}
	return fmt.Sprintf("%s: %d link(s) impacted across %d counterpart(s)", b.ChangedSectionID, total, len(b.Groups))
}

func counterpartyKey(ids []string) string {
	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
