# Playbook Parser & Checker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Parse playbook YAML (6-keyword syntax) into a typed graph model, then evaluate CEL rules against document instances to produce pass/fail diagnostics.

**Architecture:** New `internal/playbook` package. Three layers: parser (YAML → model), model (types/edges/rules), checker (model + data → diagnostics). CEL via `google/cel-go`. Golden-example test fixtures use real c3 and sft domain playbooks.

**Tech Stack:** Go, `gopkg.in/yaml.v3`, `github.com/google/cel-go`

---

## File Structure

```
internal/playbook/
├── model.go          # Domain types: TypeDef, FieldDef, EdgeDef, RuleDef, Playbook
├── model_test.go     # Unit tests for model construction/validation
├── parse.go          # YAML → Playbook (shape-based parsing)
├── parse_test.go     # Parser tests against c3.playbook.yaml + sft.playbook.yaml
├── field.go          # Field type notation parser: "string!", "ref(container)!", "enum(a,b)"
├── field_test.go     # Field notation unit tests
├── edge.go           # Edge arrow notation parser: "A -> B [1..*]"
├── edge_test.go      # Edge notation unit tests
├── checker.go        # CEL rule evaluation engine
├── checker_test.go   # Checker tests with real graph instances
├── testdata/
│   ├── c3.playbook.yaml   # Real c3-design playbook (golden example)
│   └── sft.playbook.yaml  # Real sft playbook (golden example)
```

---

### Task 1: Field Type Notation Parser

Parse `string!`, `ref(container)!`, `enum(active, archived)`, `list(ref)` into typed structs.

**Files:**
- Create: `internal/playbook/field.go`
- Create: `internal/playbook/field_test.go`

- [ ] **Step 1: Write failing tests for field notation**

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParseField -v`
Expected: FAIL — `ParseField` not defined

- [ ] **Step 3: Implement FieldDef and ParseField**

```go
package playbook

import (
	"fmt"
	"strings"
)

// FieldDef represents a parsed field type annotation.
type FieldDef struct {
	Type     string   // "string", "enum", "ref", "list"
	Required bool     // true if notation ends with !
	Target   string   // for ref(X) or list(X) — single target
	Targets  []string // for ref(X | Y) — multiple targets
	Values   []string // for enum(a, b, c)
}

// ParseField parses a field type notation string like "string!", "ref(container)!",
// "enum(active, archived)", "list(ref)".
func ParseField(notation string) (FieldDef, error) {
	notation = strings.TrimSpace(notation)
	if notation == "" {
		return FieldDef{}, fmt.Errorf("empty field notation")
	}

	var f FieldDef

	// Check required suffix
	if strings.HasSuffix(notation, "!") {
		f.Required = true
		notation = notation[:len(notation)-1]
	}

	// Check for parameterized type: type(params)
	if idx := strings.Index(notation, "("); idx != -1 {
		if !strings.HasSuffix(notation, ")") {
			return FieldDef{}, fmt.Errorf("unclosed parenthesis in %q", notation)
		}
		f.Type = strings.TrimSpace(notation[:idx])
		params := notation[idx+1 : len(notation)-1]

		switch f.Type {
		case "enum":
			for _, v := range strings.Split(params, ",") {
				f.Values = append(f.Values, strings.TrimSpace(v))
			}
		case "ref":
			parts := strings.Split(params, "|")
			for _, p := range parts {
				f.Targets = append(f.Targets, strings.TrimSpace(p))
			}
			if len(f.Targets) == 1 {
				f.Target = f.Targets[0]
			}
		case "list":
			f.Target = strings.TrimSpace(params)
		default:
			return FieldDef{}, fmt.Errorf("unknown parameterized type %q", f.Type)
		}
	} else {
		f.Type = notation
	}

	return f, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParseField -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/field.go internal/playbook/field_test.go
git commit -m "feat(playbook): field type notation parser — string!, ref(x)!, enum(a,b), list(x)"
```

---

### Task 2: Edge Arrow Notation Parser

Parse `"component -> ref [1..*]"` into typed EdgeDef structs.

**Files:**
- Create: `internal/playbook/edge.go`
- Create: `internal/playbook/edge_test.go`

- [ ] **Step 1: Write failing tests for edge notation**

```go
package playbook

import "testing"

func TestParseEdge_Simple(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("component -> ref [1..*]")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.From) != 1 || e.From[0] != "component" {
		t.Fatalf("from: got %v", e.From)
	}
	if e.To != "ref" {
		t.Fatalf("to: got %q", e.To)
	}
	if e.MinCard != 1 || e.MaxCard != -1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_MultipleFrom(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("screen|region -> region [0..*]")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.From) != 2 || e.From[0] != "screen" || e.From[1] != "region" {
		t.Fatalf("from: got %v", e.From)
	}
}

func TestParseEdge_ExactCard(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("transition -> event [1..1]")
	if err != nil {
		t.Fatal(err)
	}
	if e.MinCard != 1 || e.MaxCard != 1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_OptionalCard(t *testing.T) {
	t.Parallel()
	e, err := ParseEdge("transition -> screen [0..1]")
	if err != nil {
		t.Fatal(err)
	}
	if e.MinCard != 0 || e.MaxCard != 1 {
		t.Fatalf("cardinality: got %d..%d", e.MinCard, e.MaxCard)
	}
}

func TestParseEdge_InvalidNoArrow(t *testing.T) {
	t.Parallel()
	_, err := ParseEdge("component ref")
	if err == nil {
		t.Fatal("expected error for missing arrow")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParseEdge -v`
Expected: FAIL — `ParseEdge` not defined

- [ ] **Step 3: Implement EdgeDef and ParseEdge**

```go
package playbook

import (
	"fmt"
	"strconv"
	"strings"
)

// EdgeDef represents a parsed edge type.
type EdgeDef struct {
	Name    string
	From    []string // source type(s)
	To      string   // target type
	MinCard int      // minimum cardinality
	MaxCard int      // maximum cardinality (-1 = unbounded)
}

// ParseEdge parses an arrow notation string like "component -> ref [1..*]".
func ParseEdge(notation string) (EdgeDef, error) {
	notation = strings.TrimSpace(notation)

	// Split on " -> "
	parts := strings.SplitN(notation, " -> ", 2)
	if len(parts) != 2 {
		return EdgeDef{}, fmt.Errorf("expected 'A -> B [card]', got %q", notation)
	}

	// Parse from (may have | for multiple types)
	fromStr := strings.TrimSpace(parts[0])
	var from []string
	for _, f := range strings.Split(fromStr, "|") {
		from = append(from, strings.TrimSpace(f))
	}

	// Parse to and cardinality
	right := strings.TrimSpace(parts[1])
	bracketIdx := strings.Index(right, "[")
	if bracketIdx == -1 {
		return EdgeDef{}, fmt.Errorf("missing cardinality in %q", notation)
	}

	to := strings.TrimSpace(right[:bracketIdx])
	cardStr := right[bracketIdx:]
	if !strings.HasPrefix(cardStr, "[") || !strings.HasSuffix(cardStr, "]") {
		return EdgeDef{}, fmt.Errorf("bad cardinality format in %q", cardStr)
	}
	cardStr = cardStr[1 : len(cardStr)-1] // strip [ ]

	cardParts := strings.SplitN(cardStr, "..", 2)
	if len(cardParts) != 2 {
		return EdgeDef{}, fmt.Errorf("cardinality must be 'min..max', got %q", cardStr)
	}

	minCard, err := strconv.Atoi(strings.TrimSpace(cardParts[0]))
	if err != nil {
		return EdgeDef{}, fmt.Errorf("bad min cardinality %q: %w", cardParts[0], err)
	}

	maxStr := strings.TrimSpace(cardParts[1])
	maxCard := -1
	if maxStr != "*" {
		maxCard, err = strconv.Atoi(maxStr)
		if err != nil {
			return EdgeDef{}, fmt.Errorf("bad max cardinality %q: %w", maxStr, err)
		}
	}

	return EdgeDef{
		From:    from,
		To:      to,
		MinCard: minCard,
		MaxCard: maxCard,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParseEdge -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/edge.go internal/playbook/edge_test.go
git commit -m "feat(playbook): edge arrow notation parser — A -> B [min..max]"
```

---

### Task 3: Playbook Model

Core domain types that hold a parsed playbook in memory.

**Files:**
- Create: `internal/playbook/model.go`
- Create: `internal/playbook/model_test.go`

- [ ] **Step 1: Write failing tests for model construction**

```go
package playbook

import "testing"

func TestPlaybook_TypeByName(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {Name: "component", Description: "test"},
		},
	}
	td := pb.Type("component")
	if td == nil || td.Name != "component" {
		t.Fatal("expected component type")
	}
	if pb.Type("missing") != nil {
		t.Fatal("expected nil for missing type")
	}
}

func TestPlaybook_EdgeByName(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Edges: map[string]*EdgeDef{
			"cites": {Name: "cites", To: "ref"},
		},
	}
	if pb.Edge("cites") == nil {
		t.Fatal("expected cites edge")
	}
}

func TestPlaybook_GlobalRules(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Rules: map[string]*RuleDef{
			"field-required": {Name: "field-required", Severity: "error"},
		},
	}
	if len(pb.Rules) != 1 {
		t.Fatalf("expected 1 global rule, got %d", len(pb.Rules))
	}
}

func TestTypeDef_HasField(t *testing.T) {
	t.Parallel()
	td := &TypeDef{
		Fields: map[string]FieldDef{
			"goal": {Type: "string", Required: true},
		},
	}
	if _, ok := td.Fields["goal"]; !ok {
		t.Fatal("expected goal field")
	}
}

func TestTypeDef_RequiredSections(t *testing.T) {
	t.Parallel()
	td := &TypeDef{
		Sections: []SectionDef{
			{Name: "Dependencies", Required: true},
			{Name: "Related Refs", Required: false},
		},
	}
	req := td.RequiredSections()
	if len(req) != 1 || req[0].Name != "Dependencies" {
		t.Fatalf("expected [Dependencies], got %v", req)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run 'TestPlaybook_|TestTypeDef_' -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Implement model types**

```go
package playbook

// Playbook is a parsed playbook — types, edges, and rules.
type Playbook struct {
	Types map[string]*TypeDef  // node types (keyed by name)
	Edges map[string]*EdgeDef  // edge types (keyed by name)
	Rules map[string]*RuleDef  // global rules (apply to all types)
}

// Type returns the TypeDef for the given name, or nil.
func (p *Playbook) Type(name string) *TypeDef {
	return p.Types[name]
}

// Edge returns the EdgeDef for the given name, or nil.
func (p *Playbook) Edge(name string) *EdgeDef {
	return p.Edges[name]
}

// TypeDef defines a node type in the playbook.
type TypeDef struct {
	Name        string
	Description string
	Fields      map[string]FieldDef
	Sections    []SectionDef
	Rules       map[string]*RuleDef // type-scoped rules
}

// RequiredSections returns sections marked as required.
func (td *TypeDef) RequiredSections() []SectionDef {
	var out []SectionDef
	for _, s := range td.Sections {
		if s.Required {
			out = append(out, s)
		}
	}
	return out
}

// SectionDef defines a named content section.
type SectionDef struct {
	Name     string
	Required bool
}

// RuleDef defines a constraint rule.
type RuleDef struct {
	Name        string
	Description string
	Severity    string     // "error" or "warning"
	Expr        string     // CEL expression
	Examples    []Example  // golden examples
}

// Example is a test case for a rule: expected result + data.
type Example struct {
	Pass bool
	Data map[string]any
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run 'TestPlaybook_|TestTypeDef_' -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/model.go internal/playbook/model_test.go
git commit -m "feat(playbook): domain model — Playbook, TypeDef, EdgeDef, RuleDef"
```

---

### Task 4: Golden Example Test Fixtures

Create real c3 and sft playbook YAML files for parser tests.

**Files:**
- Create: `internal/playbook/testdata/c3.playbook.yaml`
- Create: `internal/playbook/testdata/sft.playbook.yaml`

- [ ] **Step 1: Write c3 playbook fixture**

```yaml
# playbook/1.0
# C3 Architecture Documentation — constraints for component graphs

component:
  description: "Functional unit within a container"
  goal: string!
  status: enum(active, provisioned, archived)
  parent: ref(container)!
  uses: list(ref)
  sections: [Dependencies!, "Related Refs", "Related Rules"]
  rules:
    dangling-cite:
      description: "Every citation must resolve to an existing ref"
      severity: error
      expr: 'self.uses.all(u, exists("ref", u))'
    scope-crosscheck:
      description: "If ref scopes a container, child components should cite explicitly"
      severity: warning
      expr: 'parent_refs(self).all(r, r in self.uses)'

ref:
  description: "Cross-cutting architectural pattern or decision"
  goal: string!
  status: enum(active, archived)
  sections: [Choice!, Why!, How]
  rules:
    orphan-ref:
      description: "Every ref must be cited by at least one component"
      severity: error
      expr: 'size(edges_in("cites")) >= 1'
      example:
        - [true, { "ref-logging": { cited_by: ["cmd-root", "graph"] } }]
        - [false, { "ref-jwt": { cited_by: [] } }]

rule:
  description: "Enforceable coding standard with golden example"
  goal: string!
  status: enum(active, archived)
  sections: [Rule!, "Golden Example!", "Not This", Scope, Override]

container:
  description: "Deployment boundary (process or library)"
  goal: string!
  boundary: enum(process, library)

context:
  description: "System-level boundary"
  goal: string!

adr:
  description: "Architectural decision record"
  goal: string!
  status: enum(proposed, accepted, implemented, superseded)

cites: "component -> ref [1..*]"
governs: "rule -> component [0..*]"
contains: "container -> component [1..*]"
parent-of: "context -> container [1..*]"
```

- [ ] **Step 2: Write sft playbook fixture**

```yaml
# playbook/1.0
# SFT Behavioral Spec — constraints for screen/region/event graphs

screen:
  description: "Top-level viewport grouping"
  name: string!
  description_text: string

region:
  description: "Scoped UI area with own content and events"
  name: string!
  parent: ref(screen | region)!
  rules:
    nesting-depth:
      description: "Regions must not nest deeper than 2 levels"
      severity: warning
      expr: 'depth(self, "nests") <= 2'

event:
  description: "User or system action emitted by a region"
  name: string!
  emitted_by: ref(region)!
  annotation: string
  rules:
    orphan-emit:
      description: "Every emitted event must have at least one handler"
      severity: error
      expr: 'size(edges_in("handles")) >= 1'
      example:
        - [true, { "select-email": { handlers: ["inbox.browsing->detail"] } }]
        - [false, { "notify-sent": { handlers: [] } }]

transition:
  description: "State change triggered by event"
  on: ref(event)!
  from: string!
  to: string
  action: string
  guard: string
  rules:
    unreachable-state:
      description: "Every non-initial state must be reachable as a transition target"
      severity: error
      expr: 'states(self.owner).all(s, s == initial || size(transitions_to(s)) > 0)'
    duplicate-transition:
      description: "No two transitions with same event+from_state in same owner"
      severity: error
      expr: 'count(siblings(self, on: self.on, from: self.from)) == 1'

flow:
  description: "Named user journey across screens"
  name: string!
  sequence: string!
  rules:
    invalid-flow-ref:
      description: "Flow references must resolve to known screens, regions, or events"
      severity: error
      expr: 'self.steps.all(s, exists(s.type, s.target))'

data-type:
  description: "Domain type with fields"
  name: string!

nests: "screen|region -> region [0..*]"
handles: "transition -> event [1..1]"
emits: "region -> event [0..*]"
navigates: "transition -> screen [0..1]"

# Global rules
field-required:
  description: "Required fields must have non-empty values"
  severity: error
  expr: 'self.fields.filter(f, f.required).all(f, len(f.value) > 0)'

enum-valid:
  description: "Enum fields must contain allowed values"
  severity: error
  expr: 'self.fields.filter(f, f.type == "enum").all(f, f.value in f.values)'
```

- [ ] **Step 3: Commit**

```bash
git add internal/playbook/testdata/
git commit -m "feat(playbook): golden example fixtures — c3 + sft playbooks"
```

---

### Task 5: YAML Parser — Shape-Based Parsing

Parse playbook YAML into the Playbook model using shape-based classification.

**Files:**
- Create: `internal/playbook/parse.go`
- Create: `internal/playbook/parse_test.go`

- [ ] **Step 1: Write failing tests — parse c3 fixture**

```go
package playbook

import (
	"os"
	"testing"
)

func TestParse_C3_Types(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	if len(pb.Types) != 6 {
		t.Fatalf("expected 6 types, got %d: %v", len(pb.Types), typeNames(pb))
	}
	for _, name := range []string{"component", "ref", "rule", "container", "context", "adr"} {
		if pb.Type(name) == nil {
			t.Errorf("missing type %q", name)
		}
	}
}

func TestParse_C3_Fields(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if comp == nil {
		t.Fatal("missing component type")
	}
	goalField, ok := comp.Fields["goal"]
	if !ok {
		t.Fatal("missing goal field")
	}
	if goalField.Type != "string" || !goalField.Required {
		t.Fatalf("goal: type=%q required=%v", goalField.Type, goalField.Required)
	}

	parentField := comp.Fields["parent"]
	if parentField.Type != "ref" || parentField.Target != "container" || !parentField.Required {
		t.Fatalf("parent: type=%q target=%q required=%v", parentField.Type, parentField.Target, parentField.Required)
	}
}

func TestParse_C3_Sections(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if len(comp.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(comp.Sections))
	}
	if !comp.Sections[0].Required || comp.Sections[0].Name != "Dependencies" {
		t.Fatalf("first section: %+v", comp.Sections[0])
	}
}

func TestParse_C3_Edges(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	if len(pb.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(pb.Edges))
	}
	cites := pb.Edge("cites")
	if cites == nil {
		t.Fatal("missing cites edge")
	}
	if cites.To != "ref" || cites.MinCard != 1 || cites.MaxCard != -1 {
		t.Fatalf("cites: to=%q min=%d max=%d", cites.To, cites.MinCard, cites.MaxCard)
	}
}

func TestParse_C3_TypeRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	comp := pb.Type("component")
	if len(comp.Rules) != 2 {
		t.Fatalf("expected 2 component rules, got %d", len(comp.Rules))
	}
	dc := comp.Rules["dangling-cite"]
	if dc == nil || dc.Severity != "error" || dc.Expr == "" {
		t.Fatalf("dangling-cite rule: %+v", dc)
	}
}

func TestParse_C3_RuleExamples(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	ref := pb.Type("ref")
	orphan := ref.Rules["orphan-ref"]
	if orphan == nil {
		t.Fatal("missing orphan-ref rule")
	}
	if len(orphan.Examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(orphan.Examples))
	}
	if !orphan.Examples[0].Pass {
		t.Fatal("first example should pass")
	}
	if orphan.Examples[1].Pass {
		t.Fatal("second example should fail")
	}
}

func mustParseFixture(t *testing.T, path string) *Playbook {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return pb
}

func typeNames(pb *Playbook) []string {
	var names []string
	for n := range pb.Types {
		names = append(names, n)
	}
	return names
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParse_C3 -v`
Expected: FAIL — `Parse` not defined

- [ ] **Step 3: Implement shape-based YAML parser**

```go
package playbook

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Reserved keys — not fields.
var reservedKeys = map[string]bool{
	"description": true,
	"sections":    true,
	"rules":       true,
	"severity":    true,
	"expr":        true,
	"example":     true,
}

// Parse parses playbook YAML bytes into a Playbook.
func Parse(data []byte) (*Playbook, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}

	pb := &Playbook{
		Types: make(map[string]*TypeDef),
		Edges: make(map[string]*EdgeDef),
		Rules: make(map[string]*RuleDef),
	}

	for key, val := range raw {
		switch v := val.(type) {
		case string:
			// String with "->" = edge
			if strings.Contains(v, "->") {
				edge, err := ParseEdge(v)
				if err != nil {
					return nil, fmt.Errorf("edge %q: %w", key, err)
				}
				edge.Name = key
				pb.Edges[key] = &edge
			}
		case map[string]any:
			// Has expr: = rule
			if _, hasExpr := v["expr"]; hasExpr {
				rule, err := parseRule(key, v)
				if err != nil {
					return nil, fmt.Errorf("rule %q: %w", key, err)
				}
				pb.Rules[key] = rule
			} else {
				// Otherwise = type definition
				td, err := parseTypeDef(key, v)
				if err != nil {
					return nil, fmt.Errorf("type %q: %w", key, err)
				}
				pb.Types[key] = td
			}
		}
	}

	return pb, nil
}

func parseTypeDef(name string, raw map[string]any) (*TypeDef, error) {
	td := &TypeDef{
		Name:   name,
		Fields: make(map[string]FieldDef),
		Rules:  make(map[string]*RuleDef),
	}

	for key, val := range raw {
		switch key {
		case "description":
			td.Description = str(val)
		case "sections":
			secs, err := parseSections(val)
			if err != nil {
				return nil, err
			}
			td.Sections = secs
		case "rules":
			rules, err := parseRulesMap(val)
			if err != nil {
				return nil, err
			}
			td.Rules = rules
		default:
			// Everything else is a field
			field, err := ParseField(str(val))
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", key, err)
			}
			td.Fields[key] = field
		}
	}

	return td, nil
}

func parseSections(val any) ([]SectionDef, error) {
	list, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("sections must be a list")
	}
	var out []SectionDef
	for _, item := range list {
		s := str(item)
		required := strings.HasSuffix(s, "!")
		if required {
			s = s[:len(s)-1]
		}
		out = append(out, SectionDef{Name: s, Required: required})
	}
	return out, nil
}

func parseRulesMap(val any) (map[string]*RuleDef, error) {
	m, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("rules must be a map")
	}
	rules := make(map[string]*RuleDef)
	for name, rv := range m {
		rm, ok := rv.(map[string]any)
		if !ok {
			continue
		}
		rule, err := parseRule(name, rm)
		if err != nil {
			return nil, err
		}
		rules[name] = rule
	}
	return rules, nil
}

func parseRule(name string, raw map[string]any) (*RuleDef, error) {
	r := &RuleDef{
		Name:     name,
		Severity: "error", // default
	}
	if v, ok := raw["description"]; ok {
		r.Description = str(v)
	}
	if v, ok := raw["severity"]; ok {
		r.Severity = str(v)
	}
	if v, ok := raw["expr"]; ok {
		r.Expr = str(v)
	}
	if v, ok := raw["example"]; ok {
		examples, err := parseExamples(v)
		if err != nil {
			return nil, fmt.Errorf("examples: %w", err)
		}
		r.Examples = examples
	}
	return r, nil
}

func parseExamples(val any) ([]Example, error) {
	list, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("example must be a list")
	}
	var out []Example
	for _, item := range list {
		tuple, ok := item.([]any)
		if !ok || len(tuple) != 2 {
			return nil, fmt.Errorf("each example must be [bool, data]")
		}
		pass, ok := tuple[0].(bool)
		if !ok {
			return nil, fmt.Errorf("first element of example must be bool")
		}
		data, ok := tuple[1].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("second element of example must be a map")
		}
		out = append(out, Example{Pass: pass, Data: data})
	}
	return out, nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParse_C3 -v`
Expected: all PASS

- [ ] **Step 5: Write failing tests — parse sft fixture**

```go
func TestParse_SFT_Types(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Types) != 6 {
		t.Fatalf("expected 6 types, got %d: %v", len(pb.Types), typeNames(pb))
	}
	for _, name := range []string{"screen", "region", "event", "transition", "flow", "data-type"} {
		if pb.Type(name) == nil {
			t.Errorf("missing type %q", name)
		}
	}
}

func TestParse_SFT_Edges(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(pb.Edges))
	}
	nests := pb.Edge("nests")
	if nests == nil || len(nests.From) != 2 {
		t.Fatalf("nests edge: %+v", nests)
	}
}

func TestParse_SFT_GlobalRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	if len(pb.Rules) != 2 {
		t.Fatalf("expected 2 global rules, got %d", len(pb.Rules))
	}
}

func TestParse_SFT_NestedRules(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	transition := pb.Type("transition")
	if len(transition.Rules) != 2 {
		t.Fatalf("expected 2 transition rules, got %d", len(transition.Rules))
	}
}

func TestParse_SFT_MultiTargetRef(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	region := pb.Type("region")
	parent := region.Fields["parent"]
	if len(parent.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %v", parent.Targets)
	}
}
```

- [ ] **Step 6: Run all parse tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestParse_ -v`
Expected: all PASS (both c3 and sft)

- [ ] **Step 7: Commit**

```bash
git add internal/playbook/parse.go internal/playbook/parse_test.go
git commit -m "feat(playbook): shape-based YAML parser — types, edges, rules from flat YAML"
```

---

### Task 6: CEL Checker — Rule Evaluation Engine

Evaluate CEL rules against document graph instances. Add `google/cel-go` dependency.

**Files:**
- Create: `internal/playbook/checker.go`
- Create: `internal/playbook/checker_test.go`

- [ ] **Step 1: Add cel-go dependency**

```bash
cd /home/lagz0ne/dev/remmd && go get github.com/google/cel-go@latest
```

- [ ] **Step 2: Write failing tests for the checker**

```go
package playbook

import "testing"

// Node is a graph instance node for testing.
// In production this maps to remmd's Document/Section model.

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
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestChecker_ -v`
Expected: FAIL — `NewChecker` not defined

- [ ] **Step 4: Implement CEL checker**

```go
package playbook

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

// Checker evaluates CEL expressions against node data.
type Checker struct{}

// NewChecker creates a new Checker.
func NewChecker() (*Checker, error) {
	return &Checker{}, nil
}

// Eval evaluates a CEL expression against a data map.
// The data is bound as "self" in the expression.
// Returns true if the expression evaluates to true.
func (c *Checker) Eval(expr string, data map[string]any) (bool, error) {
	env, err := cel.NewEnv(
		cel.Variable("self", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return false, fmt.Errorf("cel env: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("cel compile: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("cel program: %w", err)
	}

	out, _, err := prg.Eval(map[string]any{
		"self": data,
	})
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

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestChecker_ -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/playbook/checker.go internal/playbook/checker_test.go go.mod go.sum
git commit -m "feat(playbook): CEL checker — evaluate rule expressions against node data"
```

---

### Task 7: Diagnostic Runner — Check Playbook Against Data

Combine parser + checker to run all rules in a playbook against a graph of nodes and produce diagnostic output.

**Files:**
- Create: `internal/playbook/diagnostic.go`
- Create: `internal/playbook/diagnostic_test.go`

- [ ] **Step 1: Write failing tests for diagnostic runner**

```go
package playbook

import "testing"

func TestDiagnostic_PassingGraph(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name: "component",
				Fields: map[string]FieldDef{
					"goal": {Type: "string", Required: true},
				},
				Rules: map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"has-goal": {
				Name:     "has-goal",
				Severity: "error",
				Expr:     "self.goal != ''",
			},
		},
	}

	nodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{"goal": "Root command"}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 0 {
		t.Fatalf("expected 0 diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestDiagnostic_FailingRule(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules:  map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{
			"has-goal": {
				Name:     "has-goal",
				Severity: "error",
				Expr:     "self.goal != ''",
			},
		},
	}

	nodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{"goal": ""}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != "error" || diags[0].Rule != "has-goal" || diags[0].NodeID != "cmd-root" {
		t.Fatalf("unexpected diagnostic: %+v", diags[0])
	}
}

func TestDiagnostic_TypeScopedRule(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"ref": {
				Name:   "ref",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"has-choice": {
						Name:     "has-choice",
						Severity: "error",
						Expr:     "self.choice != ''",
					},
				},
			},
			"component": {
				Name:   "component",
				Fields: map[string]FieldDef{},
				Rules:  map[string]*RuleDef{},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}

	nodes := []Node{
		{Type: "ref", ID: "ref-jwt", Data: map[string]any{"choice": ""}},
		{Type: "component", ID: "cmd-root", Data: map[string]any{}}, // should NOT be checked by ref rule
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic (ref only), got %d: %v", len(diags), diags)
	}
	if diags[0].NodeID != "ref-jwt" {
		t.Fatalf("expected ref-jwt, got %q", diags[0].NodeID)
	}
}

func TestDiagnostic_WarningSeverity(t *testing.T) {
	t.Parallel()
	pb := &Playbook{
		Types: map[string]*TypeDef{
			"region": {
				Name:   "region",
				Fields: map[string]FieldDef{},
				Rules: map[string]*RuleDef{
					"naming": {
						Name:     "naming",
						Severity: "warning",
						Expr:     "size(self.name) > 2",
					},
				},
			},
		},
		Edges: map[string]*EdgeDef{},
		Rules: map[string]*RuleDef{},
	}

	nodes := []Node{
		{Type: "region", ID: "r1", Data: map[string]any{"name": "AB"}},
	}

	diags := Run(pb, nodes)
	if len(diags) != 1 || diags[0].Severity != "warning" {
		t.Fatalf("expected 1 warning, got %v", diags)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestDiagnostic_ -v`
Expected: FAIL — `Node`, `Run`, `Diagnostic` not defined

- [ ] **Step 3: Implement diagnostic runner**

```go
package playbook

// Node represents a graph instance node to check against a playbook.
type Node struct {
	Type string
	ID   string
	Data map[string]any
}

// Diagnostic is a single rule violation or warning.
type Diagnostic struct {
	Rule     string // rule name
	NodeID   string // which node failed
	NodeType string // node's type
	Severity string // "error" or "warning"
	Message  string // human-readable
}

// Run evaluates all playbook rules against a set of nodes.
// Global rules run against every node. Type-scoped rules run only
// against nodes of that type.
func Run(pb *Playbook, nodes []Node) []Diagnostic {
	checker, err := NewChecker()
	if err != nil {
		return []Diagnostic{{Rule: "_internal", Severity: "error", Message: err.Error()}}
	}

	var diags []Diagnostic

	for _, node := range nodes {
		// Global rules — apply to all nodes
		for _, rule := range pb.Rules {
			if d := evalRule(checker, rule, node); d != nil {
				diags = append(diags, *d)
			}
		}

		// Type-scoped rules — apply only to matching type
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
		return &Diagnostic{
			Rule:     rule.Name,
			NodeID:   node.ID,
			NodeType: node.Type,
			Severity: rule.Severity,
			Message:  msg,
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestDiagnostic_ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/playbook/diagnostic.go internal/playbook/diagnostic_test.go
git commit -m "feat(playbook): diagnostic runner — evaluate rules, produce pass/fail diagnostics"
```

---

### Task 8: Integration Test — Full Pipeline

End-to-end: parse fixture → create graph → run diagnostics → verify output.

**Files:**
- Modify: `internal/playbook/parse_test.go` (add integration tests)

- [ ] **Step 1: Write failing integration test**

```go
func TestIntegration_C3_ParseAndCheck(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/c3.playbook.yaml")

	// Valid graph — should produce 0 errors
	validNodes := []Node{
		{Type: "component", ID: "cmd-root", Data: map[string]any{
			"goal": "Root command", "uses": []any{"ref-logging"},
		}},
		{Type: "ref", ID: "ref-logging", Data: map[string]any{
			"goal": "Structured logging",
		}},
	}
	diags := Run(pb, validNodes)
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Fatalf("expected 0 errors for valid graph, got %d: %v", len(errors), errors)
	}
}

func TestIntegration_SFT_ParseAndCheck(t *testing.T) {
	t.Parallel()
	pb := mustParseFixture(t, "testdata/sft.playbook.yaml")

	// Valid graph
	validNodes := []Node{
		{Type: "screen", ID: "inbox", Data: map[string]any{
			"name": "Inbox",
		}},
		{Type: "region", ID: "email-list", Data: map[string]any{
			"name": "EmailList", "parent": "inbox",
		}},
		{Type: "event", ID: "select-email", Data: map[string]any{
			"name": "select-email", "emitted_by": "email-list",
		}},
	}
	diags := Run(pb, validNodes)
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func filterErrors(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity == "error" {
			out = append(out, d)
		}
	}
	return out
}
```

- [ ] **Step 2: Run integration tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -run TestIntegration_ -v`
Expected: all PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/playbook/ -v`
Expected: all tests PASS

- [ ] **Step 4: Commit**

```bash
git add internal/playbook/parse_test.go
git commit -m "test(playbook): integration tests — full pipeline parse→check with c3+sft fixtures"
```

---

### Task 9: Run Full Project Tests

Verify nothing is broken.

- [ ] **Step 1: Run all project tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./... -count=1`
Expected: all PASS

- [ ] **Step 2: Final commit if any fixups needed**

```bash
git add -A && git commit -m "chore: fixups from full test run"
```
