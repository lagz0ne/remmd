package playbook

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type GraphContext interface {
	EdgesIn(nodeID string, edgeType string) []map[string]any
	EdgesOut(nodeID string, edgeType string) []map[string]any
	NodeExists(nodeType string, nodeID string) bool
}

func NewCheckerWithGraph(gc GraphContext) (*Checker, error) {
	c := &Checker{graph: gc}

	mapType := cel.MapType(cel.StringType, cel.DynType)
	listMapType := cel.ListType(mapType)

	env, err := cel.NewEnv(
		cel.Variable("self", mapType),
		cel.Variable("proposed", mapType),
		cel.Variable("principal", mapType),
		cel.Variable("action", mapType),

		cel.Function("edges_in",
			cel.Overload("edges_in_string",
				[]*cel.Type{cel.StringType}, listMapType,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					return edgesToCEL(gc.EdgesIn(c.nodeID, arg.Value().(string)))
				}),
			),
		),
		cel.Function("edges_out",
			cel.Overload("edges_out_string",
				[]*cel.Type{cel.StringType}, listMapType,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					return edgesToCEL(gc.EdgesOut(c.nodeID, arg.Value().(string)))
				}),
			),
		),
		cel.Function("node_exists",
			cel.Overload("node_exists_string_string",
				[]*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return types.Bool(gc.NodeExists(lhs.Value().(string), rhs.Value().(string)))
				}),
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("cel env: %w", err)
	}
	c.env = env
	return c, nil
}

func edgesToCEL(edges []map[string]any) ref.Val {
	if edges == nil {
		edges = []map[string]any{}
	}
	result := make([]any, len(edges))
	for i, e := range edges {
		result[i] = e
	}
	return types.DefaultTypeAdapter.NativeToValue(result)
}
