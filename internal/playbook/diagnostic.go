package playbook

import "strings"

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

// Run evaluates all playbook rules against nodes without graph context.
func Run(pb *Playbook, nodes []Node) []Diagnostic {
	checker, err := NewChecker()
	if err != nil {
		return []Diagnostic{{Rule: "_internal", Severity: "error", Message: err.Error()}}
	}
	return runWithChecker(pb, nodes, checker)
}

// RunWithGraph evaluates rules with graph-aware CEL functions (edges_in, edges_out, etc.).
func RunWithGraph(pb *Playbook, nodes []Node, gc GraphContext) []Diagnostic {
	checker, err := NewCheckerWithGraph(gc)
	if err != nil {
		return []Diagnostic{{Rule: "_internal", Severity: "error", Message: err.Error()}}
	}
	return runWithChecker(pb, nodes, checker)
}

func runWithChecker(pb *Playbook, nodes []Node, checker *Checker) []Diagnostic {
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
	// Skip action-time guards (they reference principal/proposed which aren't available in static validation)
	if strings.Contains(rule.Expr, "principal.") || strings.Contains(rule.Expr, "proposed.") {
		return nil
	}
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
