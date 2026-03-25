package core

import (
	"sort"
	"time"
)

// DocumentStatus represents the lifecycle state of a document.
type DocumentStatus string

const (
	DocumentActive   DocumentStatus = "active"
	DocumentArchived DocumentStatus = "archived"
)

// Document is the primary aggregate for content authoring.
type Document struct {
	ID          string
	Title       string
	OwnerID     string
	Status      DocumentStatus
	Source      string
	DocType     string
	ParentDocID string
	CreatedAt   time.Time
	sections    []Section
	tags        map[string]bool
}

// NewDocument creates a new Document with an auto-generated ID, active status,
// and current timestamp.
func NewDocument(title, ownerID string) *Document {
	return &Document{
		ID:          NewID().String(),
		Title:       title,
		OwnerID:     ownerID,
		Status:      DocumentActive,
		Source:    "native",
		CreatedAt: time.Now(),
		tags:        make(map[string]bool),
	}
}

// AddSections appends sections to the document.
func (d *Document) AddSections(sections []Section) {
	d.sections = append(d.sections, sections...)
}

// Sections returns a copy of the document's sections.
func (d *Document) Sections() []Section {
	out := make([]Section, len(d.sections))
	copy(out, d.sections)
	return out
}

// AddTag adds a tag to the document. Idempotent.
func (d *Document) AddTag(tag string) {
	d.tags[tag] = true
}

// RemoveTag removes a tag from the document.
func (d *Document) RemoveTag(tag string) {
	delete(d.tags, tag)
}

// Tags returns the document's tags in sorted order.
func (d *Document) Tags() []string {
	out := make([]string, 0, len(d.tags))
	for tag := range d.tags {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}
