package core

import "fmt"

// SectionType classifies a parsed section of a document.
type SectionType string

const (
	SectionHeading   SectionType = "heading"
	SectionListItem  SectionType = "list-item"
	SectionChecklist SectionType = "checklist"
	SectionTableRow  SectionType = "table-row"
	SectionCodeBlock SectionType = "code-block"
)

// ContentType distinguishes native vs externally-sourced section content.
type ContentType string

const (
	ContentNative   ContentType = "native"
	ContentExternal ContentType = "external"
)

// Section represents a single parsed block within a document.
type Section struct {
	ID          string
	Ref         Ref
	DocID       string
	Type        SectionType
	Title       string
	Content     string
	ContentHash string
	ContentType ContentType
	Metadata    string
	ParentRef   *Ref
	Order       int
	Checked     bool // only meaningful for SectionChecklist
}

// ValidateExternal checks invariants for external sections.
func (s Section) ValidateExternal() error {
	if s.ContentType != ContentExternal {
		return nil
	}
	if !s.Ref.IsExternal() {
		return fmt.Errorf("external section %q must have an external ref", s.ID)
	}
	if s.ContentHash == "" {
		return fmt.Errorf("external section %q must have a content hash", s.ID)
	}
	if s.Content != "" {
		return fmt.Errorf("external section %q must not have inline content", s.ID)
	}
	return nil
}
