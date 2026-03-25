package core

import "time"

// Relation is a lightweight structural edge between two documents.
// Unlike Links, relations have no approval workflow or state machine.
type Relation struct {
	ID           string
	FromDocID    string
	ToDocID      string
	RelationType string
	Metadata     string
	CreatedAt    time.Time
}

// NewRelation creates a new Relation with auto-generated ID.
func NewRelation(fromDocID, toDocID, relationType string) *Relation {
	return &Relation{
		ID:           NewID().String(),
		FromDocID:    fromDocID,
		ToDocID:      toDocID,
		RelationType: relationType,
		Metadata:     "{}",
		CreatedAt:    time.Now(),
	}
}
