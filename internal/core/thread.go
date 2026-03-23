package core

import "time"

// EntryType classifies a thread entry.
type EntryType string

const (
	EntryComment EntryType = "comment"
	EntrySystem  EntryType = "system"
)

// ThreadEntry is a single immutable entry in a link's discussion thread.
type ThreadEntry struct {
	ID          string
	LinkID      string
	Type        EntryType
	PrincipalID string
	Body        string
	CreatedAt   time.Time
}

// Thread is an append-only discussion log attached to a link.
type Thread struct {
	entries []ThreadEntry
}

// NewThread creates an empty thread.
func NewThread() *Thread {
	return &Thread{}
}

// AddComment appends a human comment to the thread.
func (t *Thread) AddComment(principalID, body string) ThreadEntry {
	entry := ThreadEntry{
		ID:          NewID().String(),
		Type:        EntryComment,
		PrincipalID: principalID,
		Body:        body,
		CreatedAt:   time.Now(),
	}
	t.entries = append(t.entries, entry)
	return entry
}

// AddSystemEvent appends a system-generated entry to the thread.
func (t *Thread) AddSystemEvent(eventType, body string) ThreadEntry {
	entry := ThreadEntry{
		ID:          NewID().String(),
		Type:        EntrySystem,
		PrincipalID: "",
		Body:        body,
		CreatedAt:   time.Now(),
	}
	t.entries = append(t.entries, entry)
	return entry
}

// Entries returns a copy of all entries in chronological order.
func (t *Thread) Entries() []ThreadEntry {
	out := make([]ThreadEntry, len(t.entries))
	copy(out, t.entries)
	return out
}
