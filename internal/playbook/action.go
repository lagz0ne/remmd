package playbook

type Action struct {
	Verb     string         // "create", "update", "delete"
	NodeType string
	NodeID   string
	Old      map[string]any // current state (trusted, from store)
	New      map[string]any // proposed state (untrusted, from agent)
}

func (a Action) ToMap() map[string]any {
	return map[string]any{
		"verb":      a.Verb,
		"node_type": a.NodeType,
		"node_id":   a.NodeID,
	}
}
