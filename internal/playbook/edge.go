package playbook

import (
	"fmt"
	"strconv"
	"strings"
)

// EdgeDef is a parsed edge type from arrow notation like "component -> ref [1..*]".
type EdgeDef struct {
	Name    string
	From    []string
	To      string
	MinCard int // -1 = unbounded
	MaxCard int
}

// ParseEdge parses an arrow notation string.
func ParseEdge(notation string) (EdgeDef, error) {
	notation = strings.TrimSpace(notation)

	parts := strings.SplitN(notation, " -> ", 2)
	if len(parts) != 2 {
		return EdgeDef{}, fmt.Errorf("expected 'A -> B [card]', got %q", notation)
	}

	from := splitTrim(strings.TrimSpace(parts[0]), "|")

	right := strings.TrimSpace(parts[1])
	bracketIdx := strings.Index(right, "[")
	if bracketIdx == -1 {
		return EdgeDef{}, fmt.Errorf("missing cardinality in %q", notation)
	}

	to := strings.TrimSpace(right[:bracketIdx])
	cardStr := right[bracketIdx:]
	if !strings.HasSuffix(cardStr, "]") {
		return EdgeDef{}, fmt.Errorf("bad cardinality format in %q", cardStr)
	}
	cardStr = cardStr[1 : len(cardStr)-1]

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
