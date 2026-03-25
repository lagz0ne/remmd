package core

// SchemaTemplate defines a required section kind for a document type.
type SchemaTemplate struct {
	DocType      string
	RequiredKind string
	MinCount     int
}

// TemplateViolation records a section kind requirement that was not met.
type TemplateViolation struct {
	DocType string
	Kind    string
	Want    int
	Got     int
}

// ValidateAgainstTemplates checks if sections meet all template requirements.
func ValidateAgainstTemplates(sections []*Section, templates []SchemaTemplate) []TemplateViolation {
	kindCounts := map[string]int{}
	for _, s := range sections {
		if s.Kind != "" {
			kindCounts[s.Kind]++
		}
	}

	var violations []TemplateViolation
	for _, t := range templates {
		got := kindCounts[t.RequiredKind]
		if got < t.MinCount {
			violations = append(violations, TemplateViolation{
				DocType: t.DocType,
				Kind:    t.RequiredKind,
				Want:    t.MinCount,
				Got:     got,
			})
		}
	}
	return violations
}
