package playbook

import (
	"os"
	"testing"
)

func TestGuard_OwnerMatchPasses(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend", "goal": "Root"},
		New:      map[string]any{"owner": "backend", "goal": "Updated root"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestGuard_OwnerMismatchBlocks(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-2", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend", "goal": "Root"},
		New:      map[string]any{"owner": "backend", "goal": "Hacked"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errors), errors)
	}
	if errors[0].Rule != "edit-scope" {
		t.Fatalf("expected edit-scope rule, got %q", errors[0].Rule)
	}
}

func TestGuard_AdminBypassesOwnership(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend"},
		New:      map[string]any{"owner": "backend", "goal": "Admin edit"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors (admin bypass), got %d: %v", len(errors), errors)
	}
}

func TestGuard_DeleteRequiresAdmin(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"delete-gate": {
						Name:     "delete-gate",
						Severity: "error",
						Expr:     "action.verb != 'delete' || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Engineer tries to delete → blocked
	engineer := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{Verb: "delete", NodeType: "component", NodeID: "cmd-root", Old: map[string]any{"owner": "backend"}}
	diags := g.Check(engineer, action)
	if len(filterErrorDiags(diags)) != 1 {
		t.Fatalf("expected 1 error for engineer delete, got %d", len(filterErrorDiags(diags)))
	}

	// Admin deletes → allowed
	admin := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin"}}
	diags = g.Check(admin, action)
	if len(filterErrorDiags(diags)) != 0 {
		t.Fatalf("expected 0 errors for admin delete, got %d", len(filterErrorDiags(diags)))
	}
}

func TestGuard_OwnerFieldProtected(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"no-owner-steal": {
						Name:     "no-owner-steal",
						Severity: "error",
						Expr:     "!has(proposed.owner) || proposed.owner == self.owner || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Agent tries to change owner → blocked
	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend"},
		New:      map[string]any{"owner": "frontend"}, // stealing ownership
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error for owner steal, got %d", len(errors))
	}
}

func TestGuard_GlobalRulesApply(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {Name: "component", Fields: map[string]FieldDef{}, Rules: map[string]*RuleDef{}},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"no-service-delete": {
				Name:     "no-service-delete",
				Severity: "error",
				Expr:     "action.verb != 'delete' || principal.type != 'service'",
			},
		},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{Verb: "delete", NodeType: "component", NodeID: "x", Old: map[string]any{}}
	diags := g.Check(principal, action)
	if len(filterErrorDiags(diags)) != 1 {
		t.Fatalf("expected global rule to block service delete")
	}
}

func TestGuard_Integration_C3(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/c3.playbook.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pb, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Backend agent edits backend-owned component → allowed
	backend := Principal{ID: "agent-be", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb: "update", NodeType: "component", NodeID: "cmd-root",
		Old: map[string]any{"owner": "backend", "goal": "Root", "status": "active", "parent": "c3-1", "uses": []any{"ref-logging"}},
		New: map[string]any{"owner": "backend", "goal": "Updated root", "status": "active", "parent": "c3-1", "uses": []any{"ref-logging"}},
	}
	diags := g.Check(backend, action)
	for _, d := range diags {
		if d.Severity == "error" && d.Rule == "edit-scope" {
			t.Fatalf("backend agent should be allowed to edit backend-owned component: %v", d)
		}
	}

	// Frontend agent edits backend-owned component → blocked
	frontend := Principal{ID: "agent-fe", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	diags = g.Check(frontend, action)
	found := false
	for _, d := range diags {
		if d.Rule == "edit-scope" && d.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("frontend agent should be blocked from editing backend-owned component")
	}
}

func TestGuard_UnknownTypeWarning(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {Name: "component", Fields: map[string]FieldDef{}, Rules: map[string]*RuleDef{}},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}
	principal := Principal{ID: "a", Type: "human", Teams: []string{"x"}}
	action := Action{Verb: "update", NodeType: "nonexistent", NodeID: "n1", Old: map[string]any{}}
	diags := g.Check(principal, action)
	found := false
	for _, d := range diags {
		if d.Rule == "_unknown_type" && d.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected _unknown_type warning for unrecognized node type")
	}
}

func filterErrorDiags(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity == "error" {
			out = append(out, d)
		}
	}
	return out
}
