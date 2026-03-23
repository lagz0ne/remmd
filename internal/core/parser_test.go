package core

import (
	"testing"
)

func TestParse_Headings(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# A\n## B\n## C")
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "A", nil)
	assertSection(t, sections[1], SectionHeading, "B", &sections[0].Ref)
	assertSection(t, sections[2], SectionHeading, "C", &sections[0].Ref)
}

func TestParse_NestedHeadings(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# A\n## B\n### C\n## D")
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
	sections := Parse("doc1", "# Heading\n- item1\n- item2")
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "Heading", nil)
	assertSection(t, sections[1], SectionListItem, "item1", &sections[0].Ref)
	assertSection(t, sections[2], SectionListItem, "item2", &sections[0].Ref)
}

func TestParse_NestedLists(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "- parent\n  - child")
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionListItem, "parent", nil)
	assertSection(t, sections[1], SectionListItem, "child", &sections[0].Ref)
}

func TestParse_Checklists(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "- [ ] todo\n- [x] done")
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
	sections := Parse("doc1", md)
	if len(sections) != 2 {
		t.Fatalf("expected 2 table-row sections, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionTableRow, "Auth | Done", nil)
	assertSection(t, sections[1], SectionTableRow, "Pay | WIP", nil)
}

func TestParse_CodeBlock(t *testing.T) {
	t.Parallel()
	md := "```go\nfunc main(){}\n```"
	sections := Parse("doc1", md)
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
	sections := Parse("doc1", "")
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections, got %d", len(sections))
	}
}

func TestParse_SingleHeading(t *testing.T) {
	t.Parallel()
	sections := Parse("doc1", "# Title")
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	assertSection(t, sections[0], SectionHeading, "Title", nil)
}

func TestParse_MixedContent(t *testing.T) {
	t.Parallel()
	md := "# Overview\n- item\n- [ ] check\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n```py\nprint(1)\n```"
	sections := Parse("doc1", md)
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
	sections := Parse("doc1", "# Hello")
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
