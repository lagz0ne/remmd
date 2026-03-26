package playbook

import "testing"

func TestChecker_SimpleExprPass(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("self.goal != ''", map[string]any{
		"goal": "Build things",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass")
	}
}

func TestChecker_SimpleExprFail(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("self.goal != ''", map[string]any{
		"goal": "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail")
	}
}

func TestChecker_ListAll(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("self.uses.all(u, u != '')", map[string]any{
		"uses": []any{"ref-logging", "ref-error"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass")
	}
}

func TestChecker_ListAllFails(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("self.uses.all(u, u != '')", map[string]any{
		"uses": []any{"ref-logging", ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail")
	}
}

func TestChecker_SizeFunction(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("size(self.cited_by) >= 1", map[string]any{
		"cited_by": []any{"cmd-root"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass")
	}
}

func TestChecker_SizeEmpty(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Eval("size(self.cited_by) >= 1", map[string]any{
		"cited_by": []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail")
	}
}

func TestChecker_InvalidExpr(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Eval("this is not cel +++", map[string]any{})
	if err == nil {
		t.Fatal("expected error for invalid CEL")
	}
}

func TestChecker_PrincipalTeamCheck(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	ctx := EvalContext{
		Self:      map[string]any{"owner": "backend"},
		Principal: map[string]any{"teams": []any{"backend", "platform"}},
	}
	result, err := c.EvalWith("self.owner in principal.teams", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass — backend is in principal.teams")
	}
}

func TestChecker_PrincipalTeamDenied(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	ctx := EvalContext{
		Self:      map[string]any{"owner": "backend"},
		Principal: map[string]any{"teams": []any{"frontend"}},
	}
	result, err := c.EvalWith("self.owner in principal.teams", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail — frontend not owner")
	}
}

func TestChecker_ActionVerbGating(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	ctx := EvalContext{
		Self:      map[string]any{"owner": "backend"},
		Principal: map[string]any{"roles": []any{"engineer"}},
		Action:    map[string]any{"verb": "delete"},
	}
	result, err := c.EvalWith("action.verb != 'delete' || principal.roles.exists(r, r == 'admin')", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail — engineer cannot delete")
	}
}

func TestChecker_ProposedOwnerChange(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	ctx := EvalContext{
		Self:      map[string]any{"owner": "backend"},
		Proposed:  map[string]any{"owner": "frontend"},
		Principal: map[string]any{"roles": []any{"engineer"}},
		Action:    map[string]any{"verb": "update"},
	}
	// Only admins can change owner
	result, err := c.EvalWith("!has(proposed.owner) || proposed.owner == self.owner || principal.roles.exists(r, r == 'admin')", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Fatal("expected fail — engineer cannot change owner")
	}
}

func TestChecker_BackwardCompatible(t *testing.T) {
	t.Parallel()
	c, err := NewChecker()
	if err != nil {
		t.Fatal(err)
	}
	// Old-style single-binding Eval still works
	result, err := c.Eval("self.goal != ''", map[string]any{"goal": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("expected pass")
	}
}
