package playbook

import "testing"

func TestDiagnostic_PassingGraph(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name: "component",
				Fields: map[string]FieldDef{
					"goal": {Type: "string", Required: true},
				},
				Rules: map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"has-goal": {
				Name:     "has-goal",
				Severity: "error",
				Expr:     "self.goal != ''",
			},
		},
	}

	nodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{"goal": "Root command"}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 0 {
		t.Fatalf("expected 0 diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestDiagnostic_FailingRule(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules:  map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"has-goal": {
				Name:     "has-goal",
				Severity: "error",
				Expr:     "self.goal != ''",
			},
		},
	}

	nodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{"goal": ""}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != "error" || diags[0].Rule != "has-goal" || diags[0].NodeID != "cmd-root" {
		t.Fatalf("unexpected diagnostic: %+v", diags[0])
	}
}

func TestDiagnostic_TypeScopedRule(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"ref": {
				Name:   "ref",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"has-choice": {
						Name:     "has-choice",
						Severity: "error",
						Expr:     "self.choice != ''",
					},
				},
			},
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules:  map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}

	nodes := []Node{
		{Type: "ref", ID: "ref-jwt", Data: map[string]any{"choice": ""}},
		{Type: "component", ID: "cmd-root", Data: map[string]any{}}, // should NOT be checked by ref rule
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic (ref only), got %d: %v", len(diags), diags)
	}
	if diags[0].NodeID != "ref-jwt" {
		t.Fatalf("expected ref-jwt, got %q", diags[0].NodeID)
	}
}

func TestDiagnostic_WarningSeverity(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"region": {
				Name:   "region",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"naming": {
						Name:     "naming",
						Severity: "warning",
						Expr:     "size(self.name) > 2",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}

	nodes := []Node{
		{Type: "region", ID: "r1", Data: map[string]any{"name": "AB"}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 || diags[0].Severity != "warning" {
		t.Fatalf("expected 1 warning, got %v", diags)
	}
}
