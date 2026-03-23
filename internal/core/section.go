package core

// SectionType classifies a parsed section of a document.
type SectionType string

const (
	SectionHeading   SectionType = "heading"
	SectionListItem  SectionType = "list-item"
	SectionChecklist SectionType = "checklist"
	SectionTableRow  SectionType = "table-row"
	SectionCodeBlock SectionType = "code-block"
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
	ParentRef   *Ref
	Order       int
	Checked     bool // only meaningful for SectionChecklist
}
