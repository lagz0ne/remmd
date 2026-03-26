package playbook

import (
	"fmt"
	"strings"
)

// FieldDef is a parsed field type annotation like "string!", "ref(container)!", "enum(a,b)".
type FieldDef struct {
	Type     string
	Required bool
	Target   string
	Targets  []string
	Values   []string
}

// ParseField parses a field type notation string.
func ParseField(notation string) (FieldDef, error) {
	notation = strings.TrimSpace(notation)
	if notation == "" {
		return FieldDef{}, fmt.Errorf("empty field notation")
	}

	var f FieldDef
	if strings.HasSuffix(notation, "!") {
		f.Required = true
		notation = notation[:len(notation)-1]
	}

	if idx := strings.Index(notation, "("); idx != -1 {
		if !strings.HasSuffix(notation, ")") {
			return FieldDef{}, fmt.Errorf("unclosed parenthesis in %q", notation)
		}
		f.Type = strings.TrimSpace(notation[:idx])
		params := notation[idx+1 : len(notation)-1]

		switch f.Type {
		case "enum":
			f.Values = splitTrim(params, ",")
		case "ref":
			f.Targets = splitTrim(params, "|")
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

func splitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}
