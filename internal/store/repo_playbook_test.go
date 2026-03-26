package store_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

var c3PlaybookYAML = []byte(`
component:
  description: "Functional unit"
  goal: string!
  status: enum(active, archived)
  parent: ref(container)!
  sections: [Dependencies!]
  rules:
    dangling-cite:
      description: "Every citation must resolve"
      severity: error
      expr: 'self.uses.all(u, exists("ref", u))'
      example:
        - [true, {"ok": {"cited": ["a"]}}]
        - [false, {"bad": {"cited": []}}]

ref:
  description: "Cross-cutting pattern"
  goal: string!

cites: "component -> ref [1..*]"
governs: "component -> ref [0..*]"

no-orphan:
  description: "Global rule"
  severity: warning
  expr: 'size(edges_in("cites")) >= 1'
`)

func TestPlaybookStore_ImportAndLatest(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPlaybookStore(db)
	ctx := context.Background()

	ver, isNew, err := ps.Import(ctx, "c3", c3PlaybookYAML)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if ver != 1 {
		t.Errorf("version = %d, want 1", ver)
	}
	if !isNew {
		t.Error("expected isNew=true")
	}

	pb, loadVer, err := ps.Latest(ctx, "c3")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if loadVer != 1 {
		t.Errorf("loaded version = %d, want 1", loadVer)
	}

	if len(pb.Types) != 2 {
		t.Fatalf("types = %d, want 2", len(pb.Types))
	}
	comp := pb.Types["component"]
	if comp == nil {
		t.Fatal("missing type 'component'")
	}
	if comp.Description != "Functional unit" {
		t.Errorf("component description = %q", comp.Description)
	}
	if len(comp.Fields) != 3 {
		t.Errorf("component fields = %d, want 3 (goal, status, parent)", len(comp.Fields))
	}
	goalField, ok := comp.Fields["goal"]
	if !ok {
		t.Fatal("missing field 'goal'")
	}
	if goalField.Type != "string" || !goalField.Required {
		t.Errorf("goal field: type=%q required=%v", goalField.Type, goalField.Required)
	}
	statusField := comp.Fields["status"]
	if statusField.Type != "enum" {
		t.Errorf("status type = %q, want enum", statusField.Type)
	}
	if len(statusField.Values) != 2 {
		t.Errorf("status values = %v, want [active, archived]", statusField.Values)
	}
	parentField := comp.Fields["parent"]
	if parentField.Type != "ref" || parentField.Target != "container" {
		t.Errorf("parent field: type=%q target=%q", parentField.Type, parentField.Target)
	}

	if len(comp.Sections) != 1 {
		t.Fatalf("component sections = %d, want 1", len(comp.Sections))
	}
	if comp.Sections[0].Name != "Dependencies" || !comp.Sections[0].Required {
		t.Errorf("section[0] = %+v", comp.Sections[0])
	}

	if len(comp.Rules) != 1 {
		t.Fatalf("component rules = %d, want 1", len(comp.Rules))
	}
	dcRule := comp.Rules["dangling-cite"]
	if dcRule == nil {
		t.Fatal("missing rule 'dangling-cite'")
	}
	if dcRule.Severity != "error" {
		t.Errorf("dangling-cite severity = %q", dcRule.Severity)
	}
	if len(dcRule.Examples) != 2 {
		t.Errorf("dangling-cite examples = %d, want 2", len(dcRule.Examples))
	}
	if !dcRule.Examples[0].Pass || dcRule.Examples[1].Pass {
		t.Errorf("example pass flags wrong")
	}

	if len(pb.Edges) != 2 {
		t.Fatalf("edges = %d, want 2", len(pb.Edges))
	}
	citesEdge := pb.Edges["cites"]
	if citesEdge == nil {
		t.Fatal("missing edge 'cites'")
	}
	if citesEdge.To != "ref" || citesEdge.MinCard != 1 || citesEdge.MaxCard != -1 {
		t.Errorf("cites edge = %+v", citesEdge)
	}

	if len(pb.Rules) != 1 {
		t.Fatalf("global rules = %d, want 1", len(pb.Rules))
	}
	if pb.Rules["no-orphan"] == nil {
		t.Error("missing global rule 'no-orphan'")
	}
}

func TestPlaybookStore_ImportIdempotent(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPlaybookStore(db)
	ctx := context.Background()

	ver1, isNew1, err := ps.Import(ctx, "c3", c3PlaybookYAML)
	if err != nil {
		t.Fatalf("Import 1: %v", err)
	}
	if ver1 != 1 || !isNew1 {
		t.Fatalf("first import: ver=%d isNew=%v", ver1, isNew1)
	}

	ver2, isNew2, err := ps.Import(ctx, "c3", c3PlaybookYAML)
	if err != nil {
		t.Fatalf("Import 2: %v", err)
	}
	if ver2 != 1 {
		t.Errorf("second import version = %d, want 1", ver2)
	}
	if isNew2 {
		t.Error("second import should not be new")
	}
}

func TestPlaybookStore_ImportNewVersion(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPlaybookStore(db)
	ctx := context.Background()

	ver1, _, err := ps.Import(ctx, "c3", c3PlaybookYAML)
	if err != nil {
		t.Fatalf("Import v1: %v", err)
	}
	if ver1 != 1 {
		t.Fatalf("v1 = %d", ver1)
	}

	modified := strings.Replace(string(c3PlaybookYAML), "goal: string!", "goal: string!\n  extra: string", 1)
	ver2, isNew2, err := ps.Import(ctx, "c3", []byte(modified))
	if err != nil {
		t.Fatalf("Import v2: %v", err)
	}
	if ver2 != 2 {
		t.Errorf("v2 = %d, want 2", ver2)
	}
	if !isNew2 {
		t.Error("v2 should be new")
	}

	pb, loadVer, err := ps.Latest(ctx, "c3")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if loadVer != 2 {
		t.Errorf("latest version = %d, want 2", loadVer)
	}
	comp := pb.Types["component"]
	if comp == nil {
		t.Fatal("missing component type in v2")
	}
	if _, ok := comp.Fields["extra"]; !ok {
		t.Error("v2 should have 'extra' field")
	}
}

func TestPlaybookStore_LatestNotFound(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPlaybookStore(db)
	ctx := context.Background()

	pb, ver, err := ps.Latest(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Latest should not error for missing: %v", err)
	}
	if pb != nil {
		t.Error("expected nil playbook for nonexistent name")
	}
	if ver != 0 {
		t.Errorf("version = %d, want 0", ver)
	}
}

func TestPlaybookStore_LatestVersion(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPlaybookStore(db)
	ctx := context.Background()

	ver, hash, err := ps.LatestVersion(ctx, "c3")
	if err != nil {
		t.Fatalf("LatestVersion before import: %v", err)
	}
	if ver != 0 || hash != "" {
		t.Errorf("before import: ver=%d hash=%q", ver, hash)
	}

	ps.Import(ctx, "c3", c3PlaybookYAML)
	ver, hash, err = ps.LatestVersion(ctx, "c3")
	if err != nil {
		t.Fatalf("LatestVersion after import: %v", err)
	}
	if ver != 1 {
		t.Errorf("after import: ver=%d, want 1", ver)
	}
	if hash == "" {
		t.Error("hash should be non-empty after import")
	}
}
