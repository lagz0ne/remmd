package playbook

// Playbook is a parsed playbook — types, edges, and rules.
type Playbook struct {
	Types map[string]*TypeDef
	Edges map[string]*EdgeDef
	Rules map[string]*RuleDef // global rules
}

func (p *Playbook) Type(name string) *TypeDef { return p.Types[name] }
func (p *Playbook) Edge(name string) *EdgeDef { return p.Edges[name] }

// TypeDef defines a node type in the playbook.
type TypeDef struct {
	Name        string
	Description string
	Fields      map[string]FieldDef
	Sections    []SectionDef
	Rules       map[string]*RuleDef
}

func (td *TypeDef) RequiredSections() []SectionDef {
	var out []SectionDef
	for _, s := range td.Sections {
		if s.Required {
			out = append(out, s)
		}
	}
	return out
}

type SectionDef struct {
	Name     string
	Required bool
}

type RuleDef struct {
	Name        string
	Description string
	Severity    string // "error" or "warning"
	Expr        string
	Examples    []Example
}

type Example struct {
	Pass bool
	Data map[string]any
}
