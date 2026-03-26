package serve

import (
	"context"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/playbook"
)

func TestSubjectPart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		subject string
		index   int
		want    string
		wantErr bool
	}{
		{"remmd.q.documents.d1.sections", 3, "d1", false},
		{"remmd.q.section.a1", 3, "a1", false},
		{"remmd.q.documents", 0, "remmd", false},
		{"remmd.q", 3, "", true},
		{"a", 1, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got, err := subjectPart(tt.subject, tt.index)
			if (err != nil) != tt.wantErr {
				t.Fatalf("subjectPart(%q, %d) error = %v, wantErr %v", tt.subject, tt.index, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("subjectPart(%q, %d) = %q, want %q", tt.subject, tt.index, got, tt.want)
			}
		})
	}
}

func TestDeriveLinkInfoState_Priority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		states []string
		want   string
	}{
		{"no links", nil, ""},
		{"single aligned", []string{"aligned"}, "aligned"},
		{"stale + broken = broken", []string{"stale", "broken"}, "broken"},
		{"archived only = empty", []string{"archived"}, ""},
		{"archived + pending = pending", []string{"archived", "pending"}, "pending"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var links []*core.LinkInfo
			for _, s := range tt.states {
				links = append(links, &core.LinkInfo{State: s})
			}
			got := deriveLinkInfoState(links)
			if got != tt.want {
				t.Errorf("deriveLinkInfoState(%v) = %q, want %q", tt.states, got, tt.want)
			}
		})
	}
}

func TestBuildGraphResponse(t *testing.T) {
	t.Parallel()

	// Helper: section-to-doc resolver that uses a static map.
	sectionDocMap := map[string]string{
		"s1": "doc-1",
		"s2": "doc-1",
		"s3": "doc-2",
		"s4": "doc-2",
		"s5": "doc-3",
	}
	resolver := func(_ context.Context, sectionID string) (string, error) {
		docID, ok := sectionDocMap[sectionID]
		if !ok {
			return "", core.ErrNotFound{Entity: "section", ID: sectionID}
		}
		return docID, nil
	}

	t.Run("empty graph", func(t *testing.T) {
		resp := buildGraphResponse(context.Background(), nil, nil, resolver)
		if len(resp.Nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(resp.Nodes))
		}
		if len(resp.Edges) != 0 {
			t.Errorf("expected 0 edges, got %d", len(resp.Edges))
		}
	})

	t.Run("docs only no links", func(t *testing.T) {
		docs := []*core.Document{
			{ID: "doc-1", Title: "Design", Status: core.DocumentActive, Source: "native"},
			{ID: "doc-2", Title: "Spec", Status: core.DocumentArchived, Source: "git"},
		}
		resp := buildGraphResponse(context.Background(), docs, nil, resolver)
		if len(resp.Nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(resp.Nodes))
		}
		if resp.Nodes[0].ID != "doc-1" || resp.Nodes[0].Title != "Design" {
			t.Errorf("node[0] = %+v", resp.Nodes[0])
		}
		if resp.Nodes[1].Status != "archived" {
			t.Errorf("node[1].Status = %q, want %q", resp.Nodes[1].Status, "archived")
		}
		if len(resp.Edges) != 0 {
			t.Errorf("expected 0 edges, got %d", len(resp.Edges))
		}
	})

	t.Run("link resolves to source and target docs", func(t *testing.T) {
		docs := []*core.Document{
			{ID: "doc-1", Title: "Design", Status: core.DocumentActive},
			{ID: "doc-2", Title: "Impl", Status: core.DocumentActive},
		}
		links := []*core.Link{
			{
				ID:               "link-1",
				LeftSectionIDs:   []string{"s1"},
				RightSectionIDs:  []string{"s3"},
				RelationshipType: core.RelImplements,
				State:            core.LinkAligned,
			},
		}
		resp := buildGraphResponse(context.Background(), docs, links, resolver)
		if len(resp.Edges) != 1 {
			t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
		}
		edge := resp.Edges[0]
		if edge.ID != "link-1" {
			t.Errorf("edge.ID = %q, want %q", edge.ID, "link-1")
		}
		if edge.SourceDocID != "doc-1" {
			t.Errorf("edge.SourceDocID = %q, want %q", edge.SourceDocID, "doc-1")
		}
		if edge.TargetDocID != "doc-2" {
			t.Errorf("edge.TargetDocID = %q, want %q", edge.TargetDocID, "doc-2")
		}
		if edge.State != "aligned" {
			t.Errorf("edge.State = %q, want %q", edge.State, "aligned")
		}
		if edge.RelationshipType != "implements" {
			t.Errorf("edge.RelationshipType = %q, want %q", edge.RelationshipType, "implements")
		}
	})

	t.Run("link with unresolvable section is skipped", func(t *testing.T) {
		docs := []*core.Document{
			{ID: "doc-1", Title: "Design", Status: core.DocumentActive},
		}
		links := []*core.Link{
			{
				ID:               "link-bad",
				LeftSectionIDs:   []string{"s1"},
				RightSectionIDs:  []string{"s-missing"},
				RelationshipType: core.RelAgreesWith,
				State:            core.LinkBroken,
			},
		}
		resp := buildGraphResponse(context.Background(), docs, links, resolver)
		if len(resp.Edges) != 0 {
			t.Errorf("expected broken link with missing section to be skipped, got %d edges", len(resp.Edges))
		}
	})

	t.Run("multiple sections same doc picks first", func(t *testing.T) {
		docs := []*core.Document{
			{ID: "doc-1", Title: "A", Status: core.DocumentActive},
			{ID: "doc-2", Title: "B", Status: core.DocumentActive},
		}
		links := []*core.Link{
			{
				ID:               "link-multi",
				LeftSectionIDs:   []string{"s1", "s2"}, // both in doc-1
				RightSectionIDs:  []string{"s3", "s4"}, // both in doc-2
				RelationshipType: core.RelTests,
				State:            core.LinkPending,
			},
		}
		resp := buildGraphResponse(context.Background(), docs, links, resolver)
		if len(resp.Edges) != 1 {
			t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
		}
		edge := resp.Edges[0]
		if edge.SourceDocID != "doc-1" {
			t.Errorf("edge.SourceDocID = %q, want %q", edge.SourceDocID, "doc-1")
		}
		if edge.TargetDocID != "doc-2" {
			t.Errorf("edge.TargetDocID = %q, want %q", edge.TargetDocID, "doc-2")
		}
		if len(edge.LeftSectionIDs) != 2 || len(edge.RightSectionIDs) != 2 {
			t.Errorf("section IDs not preserved: left=%v right=%v", edge.LeftSectionIDs, edge.RightSectionIDs)
		}
	})
}

func TestBuildPlaybookResponse(t *testing.T) {
	t.Parallel()

	pb := &playbook.Playbook{
		Types: map[string]*playbook.TypeDef{
			"component": {
				Name:        "component",
				Description: "A software component",
				Fields: map[string]playbook.FieldDef{
					"name":   {Type: "string", Required: true},
					"status": {Type: "enum", Values: []string{"active", "deprecated"}, Required: false},
				},
				Sections: []playbook.SectionDef{
					{Name: "overview", Required: true},
					{Name: "notes", Required: false},
				},
				Rules: map[string]*playbook.RuleDef{
					"name_present": {
						Name:     "name_present",
						Severity: "error",
						Expr:     `has(data.name)`,
					},
				},
			},
		},
		Edges: map[string]*playbook.EdgeDef{
			"depends_on": {
				Name:    "depends_on",
				From:    []string{"component"},
				To:      "component",
				MinCard: 0,
				MaxCard: -1,
			},
		},
		Rules: map[string]*playbook.RuleDef{
			"global_check": {
				Name:     "global_check",
				Severity: "warning",
				Expr:     `has(data.name)`,
			},
		},
	}

	resp := buildPlaybookResponse(pb)

	// Check types
	if len(resp.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(resp.Types))
	}
	ct := resp.Types[0]
	if ct.Name != "component" {
		t.Errorf("type name = %q, want %q", ct.Name, "component")
	}
	if ct.Description != "A software component" {
		t.Errorf("type description = %q", ct.Description)
	}
	if len(ct.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(ct.Fields))
	}
	if len(ct.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(ct.Sections))
	}
	if len(ct.Rules) != 1 {
		t.Errorf("expected 1 type rule, got %d", len(ct.Rules))
	}

	// Check edges
	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
	}
	edge := resp.Edges[0]
	if edge.Name != "depends_on" {
		t.Errorf("edge name = %q", edge.Name)
	}
	if edge.Notation != "component -> component [0..*]" {
		t.Errorf("edge notation = %q, want %q", edge.Notation, "component -> component [0..*]")
	}

	// Check global rules
	if len(resp.Rules) != 1 {
		t.Fatalf("expected 1 global rule, got %d", len(resp.Rules))
	}
	if resp.Rules[0].Name != "global_check" {
		t.Errorf("global rule name = %q", resp.Rules[0].Name)
	}
}

func TestBuildValidationResponse(t *testing.T) {
	t.Parallel()

	diags := []playbook.Diagnostic{
		{Rule: "name_present", NodeID: "n1", NodeType: "component", Severity: "error", Message: "missing name"},
	}
	resp := buildValidationResponse(diags)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
	if len(resp.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(resp.Diagnostics))
	}
	d := resp.Diagnostics[0]
	if d.Rule != "name_present" || d.NodeID != "n1" || d.Severity != "error" {
		t.Errorf("diagnostic = %+v", d)
	}
}

func TestDeriveLinkState_Priority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		states []core.LinkState
		want   string
	}{
		{"no links", nil, ""},
		{"single aligned", []core.LinkState{core.LinkAligned}, "aligned"},
		{"single pending", []core.LinkState{core.LinkPending}, "pending"},
		{"single stale", []core.LinkState{core.LinkStale}, "stale"},
		{"single broken", []core.LinkState{core.LinkBroken}, "broken"},
		{"aligned + pending = pending", []core.LinkState{core.LinkAligned, core.LinkPending}, "pending"},
		{"aligned + stale = stale", []core.LinkState{core.LinkAligned, core.LinkStale}, "stale"},
		{"pending + stale = stale", []core.LinkState{core.LinkPending, core.LinkStale}, "stale"},
		{"stale + broken = broken", []core.LinkState{core.LinkStale, core.LinkBroken}, "broken"},
		{"all states = broken", []core.LinkState{core.LinkAligned, core.LinkPending, core.LinkStale, core.LinkBroken}, "broken"},
		{"archived only = empty", []core.LinkState{core.LinkArchived}, ""},
		{"archived + aligned = aligned", []core.LinkState{core.LinkArchived, core.LinkAligned}, "aligned"},
		{"archived + broken = broken", []core.LinkState{core.LinkArchived, core.LinkBroken}, "broken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var links []*core.Link
			for _, s := range tt.states {
				links = append(links, &core.Link{State: s})
			}
			got := deriveLinkState(links)
			if got != tt.want {
				t.Errorf("deriveLinkState(%v) = %q, want %q", tt.states, got, tt.want)
			}
		})
	}
}
