package playbook

import (
	"fmt"
	"sync"

	"github.com/google/cel-go/cel"
)

// Checker evaluates CEL expressions against node data. Not goroutine-safe —
// nodeID is mutated on each eval call. Create one Checker per goroutine if concurrent.
type Checker struct {
	env    *cel.Env
	graph  GraphContext
	nodeID string
	cache  sync.Map // expr string -> cel.Program
}

type EvalContext struct {
	Self      map[string]any
	Proposed  map[string]any
	Principal map[string]any
	Action    map[string]any
}

func NewChecker() (*Checker, error) {
	mapType := cel.MapType(cel.StringType, cel.DynType)
	env, err := cel.NewEnv(
		cel.Variable("self", mapType),
		cel.Variable("proposed", mapType),
		cel.Variable("principal", mapType),
		cel.Variable("action", mapType),
	)
	if err != nil {
		return nil, fmt.Errorf("cel env: %w", err)
	}
	return &Checker{env: env}, nil
}

func NewValidationChecker() (*Checker, error) {
	mapType := cel.MapType(cel.StringType, cel.DynType)
	listMapType := cel.ListType(mapType)

	env, err := cel.NewEnv(
		cel.Variable("self", mapType),
		cel.Variable("proposed", mapType),
		cel.Variable("principal", mapType),
		cel.Variable("action", mapType),

		cel.Function("edges_in",
			cel.Overload("edges_in_string", []*cel.Type{cel.StringType}, listMapType),
		),
		cel.Function("edges_out",
			cel.Overload("edges_out_string", []*cel.Type{cel.StringType}, listMapType),
		),
		cel.Function("node_exists",
			cel.Overload("node_exists_string_string", []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType),
		),
		cel.Function("exists",
			cel.Overload("exists_string_string", []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType),
		),
		cel.Function("parent_refs",
			cel.Overload("parent_refs_map", []*cel.Type{mapType}, listMapType),
		),
		cel.Function("depth",
			cel.Overload("depth_map_string", []*cel.Type{mapType, cel.StringType}, cel.IntType),
		),
		cel.Function("states",
			cel.Overload("states_map", []*cel.Type{mapType}, cel.ListType(cel.StringType)),
		),
		cel.Function("transitions_to",
			cel.Overload("transitions_to_string", []*cel.Type{cel.StringType}, cel.ListType(mapType)),
		),
		cel.Function("siblings",
			cel.Overload("siblings_map", []*cel.Type{mapType}, cel.ListType(mapType)),
		),
		cel.Function("count",
			cel.Overload("count_list", []*cel.Type{cel.ListType(cel.DynType)}, cel.IntType),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("cel env: %w", err)
	}
	return &Checker{env: env}, nil
}

func (c *Checker) Compile(expr string) error {
	_, issues := c.env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("cel compile: %w", issues.Err())
	}
	return nil
}

func (c *Checker) Eval(expr string, data map[string]any) (bool, error) {
	return c.EvalWith(expr, EvalContext{Self: data})
}

func (c *Checker) EvalWith(expr string, ctx EvalContext) (bool, error) {
	empty := map[string]any{}
	bindings := map[string]any{
		"self":      nilOr(ctx.Self, empty),
		"proposed":  nilOr(ctx.Proposed, empty),
		"principal": nilOr(ctx.Principal, empty),
		"action":    nilOr(ctx.Action, empty),
	}
	return c.eval(expr, bindings)
}

func (c *Checker) eval(expr string, bindings map[string]any) (bool, error) {
	if c.graph != nil {
		if selfMap, ok := bindings["self"].(map[string]any); ok {
			if nid, ok := selfMap["_node_id"].(string); ok {
				c.nodeID = nid
			}
		}
	}

	var prg cel.Program
	if cached, ok := c.cache.Load(expr); ok {
		prg = cached.(cel.Program)
	} else {
		ast, issues := c.env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return false, fmt.Errorf("cel compile: %w", issues.Err())
		}
		compiled, err := c.env.Program(ast)
		if err != nil {
			return false, fmt.Errorf("cel program: %w", err)
		}
		prg = compiled
		c.cache.Store(expr, prg)
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

func nilOr(m map[string]any, fallback map[string]any) map[string]any {
	if m == nil {
		return fallback
	}
	return m
}
