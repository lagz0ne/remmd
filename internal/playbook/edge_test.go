package playbook

import "testing"

func TestParseEdge_Simple(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("component -> ref [1..*]")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.From) != 1 || e.From[0] != "component" {
		t.Fatalf("from: got %v", e.From)
	}
	if e.To != "ref" {
		t.Fatalf("to: got %q", e.To)
	}
	if e.MinCard != 1 || e.MaxCard != -1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_MultipleFrom(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("screen|region -> region [0..*]")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.From) != 2 || e.From[0] != "screen" || e.From[1] != "region" {
		t.Fatalf("from: got %v", e.From)
	}
}

func TestParseEdge_ExactCard(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("transition -> event [1..1]")
	if err != nil {
		t.Fatal(err)
	}
	if e.MinCard != 1 || e.MaxCard != 1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_OptionalCard(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("transition -> screen [0..1]")
	if err != nil {
		t.Fatal(err)
	}
	if e.MinCard != 0 || e.MaxCard != 1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_InvalidNoArrow(t *testing.T) {
	t.Parallel()
	_, err := ParseEdge("component ref")
	if err == nil {
		t.Fatal("expected error for missing arrow")
	}
}
