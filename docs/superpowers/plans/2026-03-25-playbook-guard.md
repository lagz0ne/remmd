# Playbook Guard (Enforcement Layer) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the advisory `Run()` diagnostic with a `Guard` that gates mutations — evaluating ownership rules against a trusted principal with 4 CEL bindings (self, proposed, principal, action).

**Architecture:** Extend `Checker` to support 4 CEL variables. Add `Principal` (teams/roles as lists) and `Action` (verb/old/new) types. `Guard.Check()` evaluates rules using Old data as `self` (trusted) and New data as `proposed` (untrusted). Existing `Run()` stays for batch diagnostics; `Guard` is the mutation gate.

**Tech Stack:** Go, `github.com/google/cel-go` (already in go.mod)

---

## File Structure

```
internal/playbook/
├── principal.go      # Principal type (ID, Type, Teams, Roles)
├── principal_test.go  # Principal tests
├── action.go         # Action type (Verb, NodeType, NodeID, Old, New)
├── guard.go          # Guard type — Check(principal, action) → []Diagnostic
├── guard_test.go     # Guard tests — ownership, multi-team, verb gating
├── checker.go        # MODIFY — extend CEL env to 4 bindings
├── checker_test.go   # MODIFY — add tests for multi-binding eval
├── diagnostic.go     # KEEP — Run() stays as advisory batch check
├── diagnostic_test.go # KEEP — existing tests must pass
```

---

### Task 1: Principal Type

**Files:**
- Create: `internal/playbook/principal.go`
- Create: `internal/playbook/principal_test.go`

- [ ] **Step 1: Write failing tests**

```go
package playbook

import "testing"

func TestPrincipal_HasTeam(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend", "platform"}, Roles: []string{"engineer"}}
	if !p.HasTeam("backend") {
		t.Fatal("expected HasTeam(backend) = true")
	}
	if p.HasTeam("frontend") {
		t.Fatal("expected HasTeam(frontend) = false")
	}
}

func TestPrincipal_HasRole(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin", "reviewer"}}
	if !p.HasRole("admin") {
		t.Fatal("expected HasRole(admin) = true")
	}
	if p.HasRole("engineer") {
		t.Fatal("expected HasRole(engineer) = false")
	}
}

func TestPrincipal_ToMap(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	m := p.ToMap()
	if m["id"] != "agent-1" || m["type"] != "service" {
		t.Fatalf("unexpected map: %v", m)
	}
	teams, ok := m["teams"].([]any)
	if !ok || len(teams) != 1 || teams[0] != "backend" {
		t.Fatalf("teams: %v", m["teams"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestPrincipal_ -v`
Expected: FAIL — `Principal` not defined

- [ ] **Step 3: Implement Principal**

```go
package playbook

// Principal is the actor performing a mutation.
type Principal struct {
	ID    string
	Type  string   // "human" or "service"
	Teams []string
	Roles []string
}

func (p Principal) HasTeam(team string) bool {
	for _, t := range p.Teams {
		if t == team {
			return true
		}
	}
	return false
}

func (p Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ToMap converts Principal to a CEL-bindable map.
func (p Principal) ToMap() map[string]any {
	teams := make([]any, len(p.Teams))
	for i, t := range p.Teams {
		teams[i] = t
	}
	roles := make([]any, len(p.Roles))
	for i, r := range p.Roles {
		roles[i] = r
	}
	return map[string]any{
		"id":    p.ID,
		"type":  p.Type,
		"teams": teams,
		"roles": roles,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestPrincipal_ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/principal.go internal/playbook/principal_test.go
git commit -m "feat(playbook): Principal type with teams/roles as lists"
```

---

### Task 2: Action Type

**Files:**
- Create: `internal/playbook/action.go`

- [ ] **Step 1: Implement Action type**

```go
package playbook

// Action describes a mutation being checked by the Guard.
type Action struct {
	Verb     string         // "create", "update", "delete"
	NodeType string
	NodeID   string
	Old      map[string]any // current state (trusted, from store)
	New      map[string]any // proposed state (untrusted, from agent)
}

// ToMap converts Action to a CEL-bindable map.
func (a Action) ToMap() map[string]any {
	return map[string]any{
		"verb":      a.Verb,
		"node_type": a.NodeType,
		"node_id":   a.NodeID,
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/playbook/action.go
git commit -m "feat(playbook): Action type — verb, node type, old/new data"
```

---

### Task 3: Extend Checker to 4 CEL Bindings

**Files:**
- Modify: `internal/playbook/checker.go`
- Modify: `internal/playbook/checker_test.go`

- [ ] **Step 1: Write failing tests for multi-binding Eval**

Add to `checker_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run 'TestChecker_Principal|TestChecker_Action|TestChecker_Proposed|TestChecker_Backward' -v`
Expected: FAIL — `EvalContext` and `EvalWith` not defined

- [ ] **Step 3: Extend Checker implementation**

Replace `checker.go` with:

```go
package playbook

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

// Checker evaluates CEL expressions against node data.
type Checker struct {
	singleEnv *cel.Env // 1-binding env for backward compat
	guardEnv  *cel.Env // 4-binding env for Guard
}

// EvalContext holds all CEL bindings for Guard evaluation.
type EvalContext struct {
	Self      map[string]any // current state (trusted)
	Proposed  map[string]any // proposed state (untrusted)
	Principal map[string]any // actor identity
	Action    map[string]any // mutation metadata
}

func NewChecker() (*Checker, error) {
	mapType := cel.MapType(cel.StringType, cel.DynType)

	singleEnv, err := cel.NewEnv(
		cel.Variable("self", mapType),
	)
	if err != nil {
		return nil, fmt.Errorf("cel single env: %w", err)
	}

	guardEnv, err := cel.NewEnv(
		cel.Variable("self", mapType),
		cel.Variable("proposed", mapType),
		cel.Variable("principal", mapType),
		cel.Variable("action", mapType),
	)
	if err != nil {
		return nil, fmt.Errorf("cel guard env: %w", err)
	}

	return &Checker{singleEnv: singleEnv, guardEnv: guardEnv}, nil
}

// Eval evaluates expr with data bound as "self". Backward-compatible single-binding.
func (c *Checker) Eval(expr string, data map[string]any) (bool, error) {
	return c.eval(c.singleEnv, expr, map[string]any{"self": data})
}

// EvalWith evaluates expr with all 4 bindings.
func (c *Checker) EvalWith(expr string, ctx EvalContext) (bool, error) {
	proposed := ctx.Proposed
	if proposed == nil {
		proposed = map[string]any{}
	}
	principal := ctx.Principal
	if principal == nil {
		principal = map[string]any{}
	}
	action := ctx.Action
	if action == nil {
		action = map[string]any{}
	}
	return c.eval(c.guardEnv, expr, map[string]any{
		"self":      ctx.Self,
		"proposed":  proposed,
		"principal": principal,
		"action":    action,
	})
}

func (c *Checker) eval(env *cel.Env, expr string, bindings map[string]any) (bool, error) {
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("cel compile: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("cel program: %w", err)
	}

	out, _, err := prg.Eval(bindings)
	if err != nil {
		return false, fmt.Errorf("cel eval: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("cel expression must return bool, got %T", out.Value())
	}
	return result, nil
}
```

- [ ] **Step 4: Run ALL tests to verify they pass (including existing)**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -v`
Expected: all PASS — existing Eval tests unchanged, new EvalWith tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/checker.go internal/playbook/checker_test.go
git commit -m "feat(playbook): extend Checker to 4 CEL bindings (self, proposed, principal, action)"
```

---

### Task 4: Guard — Mutation Gate

**Files:**
- Create: `internal/playbook/guard.go`
- Create: `internal/playbook/guard_test.go`

- [ ] **Step 1: Write failing tests**

```go
package playbook

import "testing"

func TestGuard_OwnerMatchPasses(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend", "goal": "Root"},
		New:      map[string]any{"owner": "backend", "goal": "Updated root"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestGuard_OwnerMismatchBlocks(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-2", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend", "goal": "Root"},
		New:      map[string]any{"owner": "backend", "goal": "Hacked"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errors), errors)
	}
	if errors[0].Rule != "edit-scope" {
		t.Fatalf("expected edit-scope rule, got %q", errors[0].Rule)
	}
}

func TestGuard_AdminBypassesOwnership(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"edit-scope": {
						Name:     "edit-scope",
						Severity: "error",
						Expr:     "self.owner in principal.teams || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend"},
		New:      map[string]any{"owner": "backend", "goal": "Admin edit"},
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors (admin bypass), got %d: %v", len(errors), errors)
	}
}

func TestGuard_DeleteRequiresAdmin(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"delete-gate": {
						Name:     "delete-gate",
						Severity: "error",
						Expr:     "action.verb != 'delete' || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Engineer tries to delete → blocked
	engineer := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{Verb: "delete", NodeType: "component", NodeID: "cmd-root", Old: map[string]any{"owner": "backend"}}
	diags := g.Check(engineer, action)
	if len(filterErrorDiags(diags)) != 1 {
		t.Fatalf("expected 1 error for engineer delete, got %d", len(filterErrorDiags(diags)))
	}

	// Admin deletes → allowed
	admin := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin"}}
	diags = g.Check(admin, action)
	if len(filterErrorDiags(diags)) != 0 {
		t.Fatalf("expected 0 errors for admin delete, got %d", len(filterErrorDiags(diags)))
	}
}

func TestGuard_OwnerFieldProtected(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"no-owner-steal": {
						Name:     "no-owner-steal",
						Severity: "error",
						Expr:     "!has(proposed.owner) || proposed.owner == self.owner || principal.roles.exists(r, r == 'admin')",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Agent tries to change owner → blocked
	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb:     "update",
		NodeType: "component",
		NodeID:   "cmd-root",
		Old:      map[string]any{"owner": "backend"},
		New:      map[string]any{"owner": "frontend"}, // stealing ownership
	}

	diags := g.Check(principal, action)
	errors := filterErrorDiags(diags)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error for owner steal, got %d", len(errors))
	}
}

func TestGuard_GlobalRulesApply(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {Name: "component", Fields: map[string]FieldDef{}, Rules: map[string]*RuleDef{}},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"no-service-delete": {
				Name:     "no-service-delete",
				Severity: "error",
				Expr:     "action.verb != 'delete' || principal.type != 'service'",
			},
		},
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	principal := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{Verb: "delete", NodeType: "component", NodeID: "x", Old: map[string]any{}}
	diags := g.Check(principal, action)
	if len(filterErrorDiags(diags)) != 1 {
		t.Fatalf("expected global rule to block service delete")
	}
}

func filterErrorDiags(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity == "error" {
			out = append(out, d)
		}
	}
	return out
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestGuard_ -v`
Expected: FAIL — `NewGuard` not defined

- [ ] **Step 3: Implement Guard**

```go
package playbook

// Guard gates mutations by evaluating playbook rules against
// the acting principal, the current (trusted) node state, and the proposed changes.
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

// Check evaluates all applicable rules for the given mutation.
// self = action.Old (trusted from store), proposed = action.New (from agent).
func (g *Guard) Check(principal Principal, action Action) []Diagnostic {
	ctx := EvalContext{
		Self:      action.Old,
		Proposed:  action.New,
		Principal: principal.ToMap(),
		Action:    action.ToMap(),
	}
	if ctx.Self == nil {
		ctx.Self = map[string]any{}
	}

	var diags []Diagnostic

	// Global rules
	for _, rule := range g.pb.Rules {
		if d := g.evalGuardRule(rule, action, ctx); d != nil {
			diags = append(diags, *d)
		}
	}

	// Type-scoped rules
	td := g.pb.Type(action.NodeType)
	if td != nil {
		for _, rule := range td.Rules {
			if d := g.evalGuardRule(rule, action, ctx); d != nil {
				diags = append(diags, *d)
			}
		}
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
```

- [ ] **Step 4: Run ALL tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -v`
Expected: all PASS — Guard tests + existing tests

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/guard.go internal/playbook/guard_test.go
git commit -m "feat(playbook): Guard — mutation gate with 4-binding CEL evaluation"
```

---

### Task 5: Update Golden Example Fixtures

Add ownership rules to the c3 and sft playbook fixtures.

**Files:**
- Modify: `internal/playbook/testdata/c3.playbook.yaml`
- Modify: `internal/playbook/testdata/sft.playbook.yaml`

- [ ] **Step 1: Add ownership rules to c3 fixture**

Add to component type in `c3.playbook.yaml`:

```yaml
component:
  # ... existing fields ...
  owner: string!
  rules:
    # ... existing rules ...
    edit-scope:
      description: "Only owning team or admins can edit"
      severity: error
      expr: 'self.owner in principal.teams || principal.roles.exists(r, r == "admin")'
    no-owner-steal:
      description: "Only admins can change ownership"
      severity: error
      expr: '!has(proposed.owner) || proposed.owner == self.owner || principal.roles.exists(r, r == "admin")'
```

- [ ] **Step 2: Add ownership rules to sft fixture**

Add to screen type in `sft.playbook.yaml`:

```yaml
screen:
  # ... existing fields ...
  owner: string!
  rules:
    role-gate:
      description: "Only owning team can modify screens"
      severity: error
      expr: 'self.owner in principal.teams || principal.roles.exists(r, r == "admin")'
```

- [ ] **Step 3: Run all tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -v`
Expected: all PASS (parser tests still pass with new fields/rules)

- [ ] **Step 4: Commit**

```bash
git add internal/playbook/testdata/
git commit -m "feat(playbook): add ownership rules to c3 + sft golden fixtures"
```

---

### Task 6: Guard Integration Test

End-to-end: parse fixture → create Guard → check mutations → verify enforcement.

**Files:**
- Modify: `internal/playbook/guard_test.go` (add integration tests)

- [ ] **Step 1: Write integration test**

```go
func TestGuard_Integration_C3(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/c3.playbook.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pb, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGuard(pb)
	if err != nil {
		t.Fatal(err)
	}

	// Backend agent edits backend-owned component → allowed
	backend := Principal{ID: "agent-be", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	action := Action{
		Verb: "update", NodeType: "component", NodeID: "cmd-root",
		Old: map[string]any{"owner": "backend", "goal": "Root", "status": "active", "parent": "c3-1", "uses": []any{"ref-logging"}},
		New: map[string]any{"owner": "backend", "goal": "Updated root", "status": "active", "parent": "c3-1", "uses": []any{"ref-logging"}},
	}
	diags := g.Check(backend, action)
	for _, d := range diags {
		if d.Severity == "error" && d.Rule == "edit-scope" {
			t.Fatalf("backend agent should be allowed to edit backend-owned component: %v", d)
		}
	}

	// Frontend agent edits backend-owned component → blocked
	frontend := Principal{ID: "agent-fe", Type: "service", Teams: []string{"frontend"}, Roles: []string{"engineer"}}
	diags = g.Check(frontend, action)
	found := false
	for _, d := range diags {
		if d.Rule == "edit-scope" && d.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("frontend agent should be blocked from editing backend-owned component")
	}
}
```

- [ ] **Step 2: Run integration test**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestGuard_Integration -v`
Expected: PASS

- [ ] **Step 3: Run full project tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./... -count=1`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/playbook/guard_test.go
git commit -m "test(playbook): Guard integration test — parse c3 fixture, verify ownership enforcement"
```
