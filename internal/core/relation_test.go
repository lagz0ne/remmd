package core

import "testing"

func TestNewRelation(t *testing.T) {
	t.Parallel()
	r := NewRelation("doc-1", "doc-2", "parent_of")
	if r.ID == "" {
		t.Error("ID should be auto-generated")
	}
	if r.FromDocID != "doc-1" {
		t.Errorf("FromDocID = %q, want %q", r.FromDocID, "doc-1")
	}
	if r.ToDocID != "doc-2" {
		t.Errorf("ToDocID = %q, want %q", r.ToDocID, "doc-2")
	}
	if r.RelationType != "parent_of" {
		t.Errorf("RelationType = %q, want %q", r.RelationType, "parent_of")
	}
	if r.Metadata != "{}" {
		t.Errorf("Metadata = %q, want %q", r.Metadata, "{}")
	}
	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
