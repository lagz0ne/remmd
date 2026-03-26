package playbook

import "testing"

type mockGraph struct {
	edges map[string]map[string][]map[string]any // nodeID -> edgeType -> []sourceNodeData
	nodes map[string]map[string]bool             // nodeType -> nodeID -> exists
}

func (m *mockGraph) EdgesIn(nodeID, edgeType string) []map[string]any {
	if byType, ok := m.edges[nodeID]; ok {
		return byType[edgeType]
	}
	return nil
}

func (m *mockGraph) EdgesOut(nodeID, edgeType string) []map[string]any {
	return nil // not needed for these tests
}

func (m *mockGraph) NodeExists(nodeType, nodeID string) bool {
	if byID, ok := m.nodes[nodeType]; ok {
		return byID[nodeID]
	}
	return false
}

func TestCEL_EdgesIn_HasCitations(t *testing.T) {
	t.Parallel()
	g := &mockGraph{
		edges: map[string]map[string][]map[string]any{
			"ref-logging": {"cites": {
				{"id": "cmd-root", "goal": "Root command"},
				{"id": "graph", "goal": "Graph walker"},
			}},
		},
	}
	c, err := NewCheckerWithGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("size(edges_in('cites')) >= 1", map[string]any{
		"_node_id": "ref-logging",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass — ref-logging has 2 citations")
	}
}

func TestCEL_EdgesIn_Orphan(t *testing.T) {
	t.Parallel()
	g := &mockGraph{
		edges: map[string]map[string][]map[string]any{},
	}
	c, err := NewCheckerWithGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("size(edges_in('cites')) >= 1", map[string]any{
		"_node_id": "ref-jwt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail — ref-jwt has 0 citations")
	}
}

func TestCEL_Exists_Found(t *testing.T) {
	t.Parallel()
	g := &mockGraph{
		nodes: map[string]map[string]bool{
			"ref": {"ref-logging": true},
		},
	}
	c, err := NewCheckerWithGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("node_exists('ref', 'ref-logging')", map[string]any{
		"_node_id": "cmd-root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected true — ref-logging exists")
	}
}

func TestCEL_Exists_Missing(t *testing.T) {
	t.Parallel()
	g := &mockGraph{
		nodes: map[string]map[string]bool{},
	}
	c, err := NewCheckerWithGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("node_exists('ref', 'ref-jwt')", map[string]any{
		"_node_id": "cmd-root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected false — ref-jwt doesn't exist")
	}
}

func TestCEL_BackwardCompat_NoGraph(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("self.goal != ''", map[string]any{"goal": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass")
	}
}
