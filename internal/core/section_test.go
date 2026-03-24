package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestContentType_Constants(t *testing.T) {
	t.Parallel()
	if core.ContentNative != "native" {
		t.Errorf("ContentNative = %q, want %q", core.ContentNative, "native")
	}
	if core.ContentExternal != "external" {
		t.Errorf("ContentExternal = %q, want %q", core.ContentExternal, "external")
	}
}

func TestSection_ValidateExternal_Valid(t *testing.T) {
	t.Parallel()
	ref := core.NewExternalRef("notion", "page-abc")
	s := core.Section{
		ID:          "sec-1",
		Ref:         ref,
		DocID:       "doc-1",
		Type:        core.SectionHeading,
		Title:       "External heading",
		ContentType: core.ContentExternal,
		ContentHash: "sha256:abc123",
	}
	if err := s.ValidateExternal(); err != nil {
		t.Errorf("ValidateExternal() = %v, want nil", err)
	}
}

func TestSection_ValidateExternal_MissingHash(t *testing.T) {
	t.Parallel()
	ref := core.NewExternalRef("notion", "page-abc")
	s := core.Section{
		ID:          "sec-1",
		Ref:         ref,
		DocID:       "doc-1",
		Type:        core.SectionHeading,
		ContentType: core.ContentExternal,
		ContentHash: "",
	}
	if err := s.ValidateExternal(); err == nil {
		t.Error("ValidateExternal() = nil, want error for missing hash")
	}
}

func TestSection_ValidateExternal_HasBody(t *testing.T) {
	t.Parallel()
	ref := core.NewExternalRef("notion", "page-abc")
	s := core.Section{
		ID:          "sec-1",
		Ref:         ref,
		DocID:       "doc-1",
		Type:        core.SectionHeading,
		ContentType: core.ContentExternal,
		Content:     "this should not be here",
		ContentHash: "sha256:abc123",
	}
	if err := s.ValidateExternal(); err == nil {
		t.Error("ValidateExternal() = nil, want error for non-empty body on external section")
	}
}

func TestSection_ValidateExternal_NativeRefOnExternal(t *testing.T) {
	t.Parallel()
	nativeRef := core.NewRef("doc-1", 1)
	s := core.Section{
		ID:          "sec-1",
		Ref:         nativeRef,
		DocID:       "doc-1",
		Type:        core.SectionHeading,
		ContentType: core.ContentExternal,
		ContentHash: "sha256:abc123",
	}
	if err := s.ValidateExternal(); err == nil {
		t.Error("ValidateExternal() = nil, want error for native ref on external section")
	}
}

func TestSection_ValidateExternal_NativeSection(t *testing.T) {
	t.Parallel()
	nativeRef := core.NewRef("doc-1", 1)
	s := core.Section{
		ID:          "sec-1",
		Ref:         nativeRef,
		DocID:       "doc-1",
		Type:        core.SectionHeading,
		ContentType: core.ContentNative,
		Content:     "some content",
		ContentHash: "sha256:def456",
	}
	if err := s.ValidateExternal(); err != nil {
		t.Errorf("ValidateExternal() = %v, want nil for native section", err)
	}
}
