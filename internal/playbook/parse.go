package playbook

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse parses playbook YAML into a Playbook using shape-based classification:
// string with "->" = edge, map with "expr" = rule, other map = type.
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
			if strings.Contains(v, "->") {
				edge, err := ParseEdge(v)
				if err != nil {
					return nil, fmt.Errorf("edge %q: %w", key, err)
				}
				edge.Name = key
				pb.Edges[key] = &edge
			}
		case map[string]any:
			if _, hasExpr := v["expr"]; hasExpr {
				rule, err := parseRule(key, v)
				if err != nil {
					return nil, fmt.Errorf("rule %q: %w", key, err)
				}
				pb.Rules[key] = rule
			} else {
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
			var err error
			if td.Sections, err = parseSections(val); err != nil {
				return nil, err
			}
		case "rules":
			var err error
			if td.Rules, err = parseRulesMap(val); err != nil {
				return nil, err
			}
		default:
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
		var err error
		if r.Examples, err = parseExamples(v); err != nil {
			return nil, fmt.Errorf("examples: %w", err)
		}
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
