package core

import (
	"context"
	"time"
)

// SectionVersion is a historical snapshot of a section's content.
type SectionVersion struct {
	ID          string
	SectionID   string
	Version     int
	Content     string
	ContentHash string
	CreatedAt   time.Time
}

// DocumentRepository defines persistence operations for documents and sections.
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *Document) error
	FindDocumentByID(ctx context.Context, id string) (*Document, error)
	ListDocuments(ctx context.Context) ([]*Document, error)

	CreateSection(ctx context.Context, s *Section) error
	ListSections(ctx context.Context, docID string) ([]*Section, error)
	FindSectionByRef(ctx context.Context, docID string, ref string) (*Section, error)
	FindSectionByRefGlobal(ctx context.Context, ref string) (*Section, string, error) // returns section + docID
	UpdateSectionContent(ctx context.Context, sectionID, content, contentHash string) error
	GetSectionVersions(ctx context.Context, sectionID string) ([]SectionVersion, error)
	DeleteSection(ctx context.Context, sectionID string) error

	AddTag(ctx context.Context, sectionID, tag string) error
	RemoveTag(ctx context.Context, sectionID, tag string) error
	GetTags(ctx context.Context, sectionID string) ([]string, error)

	NextRefSeq(ctx context.Context, count int) (int, error) // reserves `count` sequential ref numbers, returns the first
}
