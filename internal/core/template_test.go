package core

import "testing"

func TestValidateAgainstTemplates_AllMet(t *testing.T) {
	t.Parallel()
	sections := []*Section{
		{Kind: "goal"},
		{Kind: "dependencies"},
	}
	templates := []SchemaTemplate{
		{DocType: "component", RequiredKind: "goal", MinCount: 1},
		{DocType: "component", RequiredKind: "dependencies", MinCount: 1},
	}
	violations := ValidateAgainstTemplates(sections, templates)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestValidateAgainstTemplates_MissingKind(t *testing.T) {
	t.Parallel()
	sections := []*Section{
		{Kind: "goal"},
	}
	templates := []SchemaTemplate{
		{DocType: "component", RequiredKind: "goal", MinCount: 1},
		{DocType: "component", RequiredKind: "dependencies", MinCount: 1},
	}
	violations := ValidateAgainstTemplates(sections, templates)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Kind != "dependencies" {
		t.Errorf("violation kind = %q, want %q", violations[0].Kind, "dependencies")
	}
	if violations[0].Want != 1 || violations[0].Got != 0 {
		t.Errorf("violation want=%d got=%d, expected want=1 got=0", violations[0].Want, violations[0].Got)
	}
}

func TestValidateAgainstTemplates_InsufficientCount(t *testing.T) {
	t.Parallel()
	sections := []*Section{
		{Kind: "test"},
	}
	templates := []SchemaTemplate{
		{DocType: "spec", RequiredKind: "test", MinCount: 3},
	}
	violations := ValidateAgainstTemplates(sections, templates)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Want != 3 || violations[0].Got != 1 {
		t.Errorf("violation want=%d got=%d, expected want=3 got=1", violations[0].Want, violations[0].Got)
	}
}

func TestValidateAgainstTemplates_EmptySections(t *testing.T) {
	t.Parallel()
	templates := []SchemaTemplate{
		{DocType: "spec", RequiredKind: "goal", MinCount: 1},
		{DocType: "spec", RequiredKind: "scope", MinCount: 1},
	}
	violations := ValidateAgainstTemplates(nil, templates)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
}
