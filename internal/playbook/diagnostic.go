package playbook

type Node struct {
	Type string
	ID   string
	Data map[string]any
}

type Diagnostic struct {
	Rule     string
	NodeID   string
	NodeType string
	Severity string
	Message  string
}

// Run evaluates all playbook rules against nodes.
// Global rules apply to all nodes; type-scoped rules only to matching types.
func Run(pb *Playbook, nodes []Node) []Diagnostic {
	checker, err := NewChecker()
	if err != nil {
		return []Diagnostic{{Rule: "_internal", Severity: "error", Message: err.Error()}}
	}

	var diags []Diagnostic

	for _, node := range nodes {
		for _, rule := range pb.Rules {
			if d := evalRule(checker, rule, node); d != nil {
				diags = append(diags, *d)
			}
		}

		td := pb.Type(node.Type)
		if td == nil {
			continue
		}
		for _, rule := range td.Rules {
			if d := evalRule(checker, rule, node); d != nil {
				diags = append(diags, *d)
			}
		}
	}

	return diags
}

func evalRule(checker *Checker, rule *RuleDef, node Node) *Diagnostic {
	pass, err := checker.Eval(rule.Expr, node.Data)
	if err != nil {
		return &Diagnostic{
			Rule:     rule.Name,
			NodeID:   node.ID,
			NodeType: node.Type,
			Severity: "error",
			Message:  "CEL error: " + err.Error(),
		}
	}
	if !pass {
		msg := rule.Description
		if msg == "" {
			msg = rule.Expr
		}
		return &Diagnostic{Rule: rule.Name, NodeID: node.ID, NodeType: node.Type, Severity: rule.Severity, Message: msg}
	}
	return nil
}
