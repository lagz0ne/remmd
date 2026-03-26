package playbook

import "testing"

func TestParseField_String(t *testing.T) {
	t.Parallel()
	f, err := ParseField("string")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "string" || f.Required {
		t.Fatalf("got type=%q required=%v", f.Type, f.Required)
	}
}

func TestParseField_StringRequired(t *testing.T) {
	t.Parallel()
	f, err := ParseField("string!")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "string" || !f.Required {
		t.Fatalf("got type=%q required=%v", f.Type, f.Required)
	}
}

func TestParseField_RefWithTarget(t *testing.T) {
	t.Parallel()
	f, err := ParseField("ref(container)!")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "ref" || f.Target != "container" || !f.Required {
		t.Fatalf("got type=%q target=%q required=%v", f.Type, f.Target, f.Required)
	}
}

func TestParseField_RefMultipleTargets(t *testing.T) {
	t.Parallel()
	f, err := ParseField("ref(screen | region)!")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "ref" || len(f.Targets) != 2 || !f.Required {
		t.Fatalf("got type=%q targets=%v required=%v", f.Type, f.Targets, f.Required)
	}
	if f.Targets[0] != "screen" || f.Targets[1] != "region" {
		t.Fatalf("targets mismatch: %v", f.Targets)
	}
}

func TestParseField_Enum(t *testing.T) {
	t.Parallel()
	f, err := ParseField("enum(active, provisioned, archived)")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "enum" || len(f.Values) != 3 {
		t.Fatalf("got type=%q values=%v", f.Type, f.Values)
	}
}

func TestParseField_List(t *testing.T) {
	t.Parallel()
	f, err := ParseField("list(ref)")
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != "list" || f.Target != "ref" {
		t.Fatalf("got type=%q target=%q", f.Type, f.Target)
	}
}

func TestParseField_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseField("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}
