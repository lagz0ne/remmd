package playbook

import "testing"

func TestPlaybook_TypeByName(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {Name: "component", Description: "test"},
		},
	}
	td := pb.Type("component")
	if td == nil || td.Name != "component" {
		t.Fatal("expected component type")
	}
	if pb.Type("missing") != nil {
		t.Fatal("expected nil for missing type")
	}
}

func TestPlaybook_EdgeByName(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Edges: map[string]*EdgeDef{
			"cites": {Name: "cites", To: "ref"},
		},
	}
	if pb.Edge("cites") == nil {
		t.Fatal("expected cites edge")
	}
}

func TestPlaybook_GlobalRules(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Rules: map[string]*RuleDef{
			"field-required": {Name: "field-required", Severity: "error"},
		},
	}
	if len(pb.Rules) != 1 {
		t.Fatalf("expected 1 global rule, got %d", len(pb.Rules))
	}
}

func TestTypeDef_HasField(t *testing.T) {
	t.Parallel()
	td := &TypeDef{
		Fields: map[string]FieldDef{
			"goal": {Type: "string", Required: true},
		},
	}
	if _, ok := td.Fields["goal"]; !ok {
		t.Fatal("expected goal field")
	}
}

func TestTypeDef_RequiredSections(t *testing.T) {
	t.Parallel()
	td := &TypeDef{
		Sections: []SectionDef{
			{Name: "Dependencies", Required: true},
			{Name: "Related Refs", Required: false},
		},
	}
	req := td.RequiredSections()
	if len(req) != 1 || req[0].Name != "Dependencies" {
		t.Fatalf("expected [Dependencies], got %v", req)
	}
}
