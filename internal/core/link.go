package core

import (
	"slices"
	"time"
)

// Link is the aggregate root for a verification agreement between two groups
// of sections. Left and right sides form the two endpoints.
type Link struct {
	ID                string
	LeftSectionIDs    []string
	RightSectionIDs   []string
	RelationshipType  RelationshipType
	Rationale         Rationale
	State             LinkState
	LeftIntervention  InterventionLevel
	RightIntervention InterventionLevel
	ProposerID        string
	CreatedAt         time.Time
}

// NewLink creates a link in the pending state.
func NewLink(leftIDs, rightIDs []string, relType RelationshipType, rationale Rationale, proposerID string) *Link {
	return &Link{
		ID:                NewID().String(),
		LeftSectionIDs:    leftIDs,
		RightSectionIDs:   rightIDs,
		RelationshipType:  relType,
		Rationale:         rationale,
		State:             LinkPending,
		LeftIntervention:  InterventionNotify,
		RightIntervention: InterventionNotify,
		ProposerID:        proposerID,
		CreatedAt:         time.Now(),
	}
}

// ContainsSection returns true if sectionID appears on either side of the link.
func (l *Link) ContainsSection(sectionID string) bool {
	return slices.Contains(l.LeftSectionIDs, sectionID) || slices.Contains(l.RightSectionIDs, sectionID)
}
