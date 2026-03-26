package playbook

type Guard struct {
	pb      *Playbook
	checker *Checker
}

func NewGuard(pb *Playbook) (*Guard, error) {
	checker, err := NewChecker()
	if err != nil {
		return nil, err
	}
	return &Guard{pb: pb, checker: checker}, nil
}

func (g *Guard) Check(principal Principal, action Action) []Diagnostic {
	ctx := EvalContext{
		Self:      action.Old,
		Proposed:  action.New,
		Principal: principal.ToMap(),
		Action:    action.ToMap(),
	}

	var diags []Diagnostic
	for _, rule := range g.pb.Rules {
		if d := g.evalGuardRule(rule, action, ctx); d != nil {
			diags = append(diags, *d)
		}
	}

	if td := g.pb.Type(action.NodeType); td != nil {
		for _, rule := range td.Rules {
			if d := g.evalGuardRule(rule, action, ctx); d != nil {
				diags = append(diags, *d)
			}
		}
	} else if action.NodeType != "" {
		diags = append(diags, Diagnostic{
			Rule:     "_unknown_type",
			NodeID:   action.NodeID,
			NodeType: action.NodeType,
			Severity: "warning",
			Message:  "type not defined in playbook",
		})
	}

	return diags
}

func (g *Guard) evalGuardRule(rule *RuleDef, action Action, ctx EvalContext) *Diagnostic {
	pass, err := g.checker.EvalWith(rule.Expr, ctx)
	if err != nil {
		return &Diagnostic{
			Rule:     rule.Name,
			NodeID:   action.NodeID,
			NodeType: action.NodeType,
			Severity: "error",
			Message:  "CEL error: " + err.Error(),
		}
	}
	if !pass {
		msg := rule.Description
		if msg == "" {
			msg = rule.Expr
		}
		return &Diagnostic{Rule: rule.Name, NodeID: action.NodeID, NodeType: action.NodeType, Severity: rule.Severity, Message: msg}
	}
	return nil
}
