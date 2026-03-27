package core

import (
	"testing"
)

func TestParse_Headings(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# A\n## B\n## C", 0)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "A", nil)
	assertSection(t, sections[1], SectionHeading, "B", &sections[0].Ref)
	assertSection(t, sections[2], SectionHeading, "C", &sections[0].Ref)
}

func TestParse_NestedHeadings(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# A\n## B\n### C\n## D", 0)
	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "A", nil)
	assertSection(t, sections[1], SectionHeading, "B", &sections[0].Ref)
	assertSection(t, sections[2], SectionHeading, "C", &sections[1].Ref)
	assertSection(t, sections[3], SectionHeading, "D", &sections[0].Ref)
}

func TestParse_ListItems(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Heading\n- item1\n- item2", 0)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "Heading", nil)
	assertSection(t, sections[1], SectionListItem, "item1", &sections[0].Ref)
	assertSection(t, sections[2], SectionListItem, "item2", &sections[0].Ref)
}

func TestParse_NestedLists(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "- parent\n  - child", 0)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionListItem, "parent", nil)
	assertSection(t, sections[1], SectionListItem, "child", &sections[0].Ref)
}

func TestParse_Checklists(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "- [ ] todo\n- [x] done", 0)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionChecklist, "todo", nil)
	assertSection(t, sections[1], SectionChecklist, "done", nil)
	if sections[0].Checked {
		t.Error("first checklist item should not be checked")
	}
	if !sections[1].Checked {
		t.Error("second checklist item should be checked")
	}
}

func TestParse_TableRows(t *testing.T) {
	t.Parallel()
	md := "| Name | Status |\n|---|---|\n| Auth | Done |\n| Pay | WIP |"
	sections := Parse("doc1", md, 0)
	if len(sections) != 2 {
		t.Fatalf("expected 2 table-row sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionTableRow, "Auth | Done", nil)
	assertSection(t, sections[1], SectionTableRow, "Pay | WIP", nil)
}

func TestParse_TableRowsPreserveHeaders(t *testing.T) {
	t.Parallel()
	md := "| Name | Status |\n|---|---|\n| Auth | Done |\n| Pay | WIP |"
	sections := Parse("doc1", md, 0)
	if len(sections) != 2 {
		t.Fatalf("expected 2 table-row sections, got %d", len(sections))
	}
	// Each table-row section should have Kind set to the header column names
	for i, s := range sections {
		if s.Kind != "Name|Status" {
			t.Errorf("section[%d].Kind = %q, want %q", i, s.Kind, "Name|Status")
		}
	}
}

func TestParse_CodeBlock(t *testing.T) {
	t.Parallel()
	md := "```go\nfunc main(){}\n```"
	sections := Parse("doc1", md, 0)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionCodeBlock, "go", nil)
	if sections[0].Content != "func main(){}" {
		t.Errorf("expected content %q, got %q", "func main(){}", sections[0].Content)
	}
}

func TestParse_EmptyDocument(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "", 0)
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections, got %d", len(sections))
	}
}

func TestParse_SingleHeading(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Title", 0)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "Title", nil)
}

func TestParse_MixedContent(t *testing.T) {
	t.Parallel()
	md := "# Overview\n- item\n- [ ] check\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n```py\nprint(1)\n```"
	sections := Parse("doc1", md, 0)
	if len(sections) < 5 {
		t.Fatalf("expected at least 5 sections, got %d", len(sections))
	}
	// Verify types in order
	expectedTypes := []SectionType{SectionHeading, SectionListItem, SectionChecklist, SectionTableRow, SectionCodeBlock}
	for i, et := range expectedTypes {
		if sections[i].Type != et {
			t.Errorf("section %d: expected type %s, got %s", i, et, sections[i].Type)
		}
	}
}

func TestParse_SectionsHaveRequiredFields(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Hello", 0)
	s := sections[0]
	if s.Ref.String() == "" {
		t.Error("Ref should be auto-assigned")
	}
	if s.DocID != "doc1" {
		t.Errorf("DocID: expected %q, got %q", "doc1", s.DocID)
	}
	if s.ContentHash == "" {
		t.Error("ContentHash should be set")
	}
	if s.Order != 0 {
		t.Errorf("first section Order should be 0, got %d", s.Order)
	}
}

func TestParse_StartSeqOffset(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Hello\n# World", 5)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	// startSeq=5 → first section seq=5 → @e5, second seq=6 → @f6
	if got := sections[0].Ref.String(); got != "@e5" {
		t.Errorf("section 0 ref: expected @e5, got %s", got)
	}
	if got := sections[0].Ref.Seq; got != 5 {
		t.Errorf("section 0 seq: expected 5, got %d", got)
	}
	if got := sections[1].Ref.String(); got != "@f6" {
		t.Errorf("section 1 ref: expected @f6, got %s", got)
	}
	if got := sections[1].Ref.Seq; got != 6 {
		t.Errorf("section 1 seq: expected 6, got %d", got)
	}
}

func TestParse_StartSeqZero_DefaultBehavior(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Hello", 0)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	// startSeq=0 → backward compat → first section seq=1 → @a1
	if got := sections[0].Ref.String(); got != "@a1" {
		t.Errorf("expected @a1, got %s", got)
	}
	if got := sections[0].Ref.Seq; got != 1 {
		t.Errorf("expected seq 1, got %d", got)
	}
}

func TestParse_HeadingWithBody(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Title\n\nFirst paragraph.\n\nSecond paragraph.", 0)
	if len(sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(sections))
	}
	want := "Title\n\nFirst paragraph.\n\nSecond paragraph."
	if sections[0].Content != want {
		t.Errorf("Content = %q, want %q", sections[0].Content, want)
	}
}

func TestParse_BodyBetweenHeadings(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# A\n\nBody of A.\n\n# B\n\nBody of B.", 0)
	if len(sections) != 2 {
		t.Fatalf("got %d sections, want 2", len(sections))
	}
	if sections[0].Content != "A\n\nBody of A." {
		t.Errorf("section 0 Content = %q", sections[0].Content)
	}
	if sections[1].Content != "B\n\nBody of B." {
		t.Errorf("section 1 Content = %q", sections[1].Content)
	}
}

func TestParse_MultiParagraphBody(t *testing.T) {
	t.Parallel()
	md := "# Intro\n\nPara one.\n\nPara two.\n\nPara three."
	sections := Parse("doc1", md, 0)
	if len(sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(sections))
	}
	want := "Intro\n\nPara one.\n\nPara two.\n\nPara three."
	if sections[0].Content != want {
		t.Errorf("Content = %q, want %q", sections[0].Content, want)
	}
}

// --- helpers ---

func assertSection(t *testing.T, s Section, typ SectionType, title string, parentRef *Ref) {
	t.Helper()
	if s.Type != typ {
		t.Errorf("expected type %s, got %s", typ, s.Type)
	}
	if s.Title != title {
		t.Errorf("expected title %q, got %q", title, s.Title)
	}
	if parentRef == nil && s.ParentRef != nil {
		t.Errorf("expected nil ParentRef, got %s", s.ParentRef)
	}
	if parentRef != nil {
		if s.ParentRef == nil {
			t.Errorf("expected ParentRef %s, got nil", parentRef)
		} else if s.ParentRef.String() != parentRef.String() {
			t.Errorf("expected ParentRef %s, got %s", parentRef, s.ParentRef)
		}
	}
}
