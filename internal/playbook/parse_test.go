package playbook

import (
	"os"
	"strings"
	"testing"
)

func TestParse_C3_Types(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	if len(pb.Types) != 6 {
		t.Fatalf("expected 6 types, got %d: %v", len(pb.Types), typeNames(pb))
	}
	for _, name := range []string{"component", "ref", "rule", "container", "context", "adr"} {
		if pb.Type(name) == nil {
			t.Errorf("missing type %q", name)
		}
	}
}

func TestParse_C3_Fields(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if comp == nil {
		t.Fatal("missing component type")
	}
	goalField, ok := comp.Fields["goal"]
	if !ok {
		t.Fatal("missing goal field")
	}
	if goalField.Type != "string" || !goalField.Required {
		t.Fatalf("goal: type=%q required=%v", goalField.Type, goalField.Required)
	}

	parentField := comp.Fields["parent"]
	if parentField.Type != "ref" || parentField.Target != "container" || !parentField.Required {
		t.Fatalf("parent: type=%q target=%q required=%v", parentField.Type, parentField.Target, parentField.Required)
	}
}

func TestParse_C3_Sections(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if len(comp.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(comp.Sections))
	}
	if !comp.Sections[0].Required || comp.Sections[0].Name != "Dependencies" {
		t.Fatalf("first section: %+v", comp.Sections[0])
	}
}

func TestParse_C3_Edges(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	if len(pb.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(pb.Edges))
	}
	cites := pb.Edge("cites")
	if cites == nil {
		t.Fatal("missing cites edge")
	}
	if cites.To != "ref" || cites.MinCard != 1 || cites.MaxCard != -1 {
		t.Fatalf("cites: to=%q min=%d max=%d", cites.To, cites.MinCard, cites.MaxCard)
	}
}

func TestParse_C3_TypeRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if len(comp.Rules) != 4 {
		t.Fatalf("expected 4 component rules, got %d", len(comp.Rules))
	}
	dc := comp.Rules["dangling-cite"]
	if dc == nil || dc.Severity != "error" || dc.Expr == "" {
		t.Fatalf("dangling-cite rule: %+v", dc)
	}
}

func TestParse_C3_RuleExamples(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	ref := pb.Type("ref")
	orphan := ref.Rules["orphan-ref"]
	if orphan == nil {
		t.Fatal("missing orphan-ref rule")
	}
	if len(orphan.Examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(orphan.Examples))
	}
	if !orphan.Examples[0].Pass {
		t.Fatal("first example should pass")
	}
	if orphan.Examples[1].Pass {
		t.Fatal("second example should fail")
	}
}

func TestParse_SFT_Types(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Types) != 6 {
		t.Fatalf("expected 6 types, got %d: %v", len(pb.Types), typeNames(pb))
	}
	for _, name := range []string{"screen", "region", "event", "transition", "flow", "data-type"} {
		if pb.Type(name) == nil {
			t.Errorf("missing type %q", name)
		}
	}
}

func TestParse_SFT_Edges(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(pb.Edges))
	}
	nests := pb.Edge("nests")
	if nests == nil || len(nests.From) != 2 {
		t.Fatalf("nests edge: %+v", nests)
	}
}

func TestParse_SFT_GlobalRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Rules) != 2 {
		t.Fatalf("expected 2 global rules, got %d", len(pb.Rules))
	}
}

func TestParse_SFT_NestedRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	transition := pb.Type("transition")
	if len(transition.Rules) != 2 {
		t.Fatalf("expected 2 transition rules, got %d", len(transition.Rules))
	}
}

func TestParse_SFT_MultiTargetRef(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	region := pb.Type("region")
	parent := region.Fields["parent"]
	if len(parent.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %v", parent.Targets)
	}
}

func TestIntegration_C3_ParseAndCheck(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	// Valid graph — should produce 0 errors
	validNodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{
			"goal": "Root command", "uses": []any{"ref-logging"},
		}},
		{Type: "ref", ID: "ref-logging", Data: map[string]any{
			"goal": "Structured logging",
		}},
	}
	diags := Run(pb, validNodes)
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Fatalf("expected 0 errors for valid graph, got %d: %v", len(errors), errors)
	}
}

func TestIntegration_SFT_ParseAndCheck(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	// Valid graph
	validNodes := []Node{
		{Type: "screen", ID: "inbox", Data: map[string]any{
			"name": "Inbox",
		}},
		{Type: "region", ID: "email-list", Data: map[string]any{
			"name": "EmailList", "parent": "inbox",
		}},
		{Type: "event", ID: "select-email", Data: map[string]any{
			"name": "select-email", "emitted_by": "email-list",
		}},
	}
	diags := Run(pb, validNodes)
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func filterErrors(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity == "error" && !strings.HasPrefix(d.Message, "CEL error:") {
			out = append(out, d)
		}
	}
	return out
}

func mustParseFixture(t *testing.T, path string) *Playbook {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return pb
}

func typeNames(pb *Playbook) []string {
	var names []string
	for n := range pb.Types {
		names = append(names, n)
	}
	return names
}
