package core

import (
	"strings"
)

type headingEntry struct {
	idx   int
	level int
}

type listEntry struct {
	idx    int
	indent int
}

// Parse parses a markdown document into a flat list of Sections with parent
// references set based on nesting (heading levels, list indentation).
// startSeq is the first sequence number to assign. Pass 0 for default behavior
// (starts at 1, producing @a1, @b2, ...).
func Parse(docID string, markdown string, startSeq int) []Section {
	if strings.TrimSpace(markdown) == "" {
		return nil
	}

	if startSeq <= 0 {
		startSeq = 1
	}

	lines := strings.Split(markdown, "\n")
	var sections []Section
	seq := startSeq - 1

	var headingStack []headingEntry
	var listStack []listEntry

	currentHeadingIdx := -1
	var bodyLines []string
	tableState := 0 // 0=none, 1=header-seen, 2=separator-seen (data rows follow)
	inCodeBlock := false
	codeBlockLang := ""
	var codeBlockLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				codeBlockLines = nil
				continue
			}
			inCodeBlock = false
			seq++
			content := strings.Join(codeBlockLines, "\n")
			s := Section{
				ID:          NewID().String(),
				Ref:         NewRef(docID, seq),
				DocID:       docID,
				Type:        SectionCodeBlock,
				Title:       codeBlockLang,
				Content:     content,
				ContentHash: ContentHash(content),
				Order:       len(sections),
			}
			if currentHeadingIdx >= 0 {
				ref := sections[currentHeadingIdx].Ref
				s.ParentRef = &ref
			}
			sections = append(sections, s)
			tableState = 0
			listStack = nil
			continue
		}
		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
			continue
		}

		if trimmed == "" {
			tableState = 0
			listStack = nil
			if currentHeadingIdx >= 0 && len(bodyLines) > 0 {
				bodyLines = append(bodyLines, "")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			flushBody(sections, currentHeadingIdx, &bodyLines)
			level := 0
			for _, c := range trimmed {
				if c == '#' {
					level++
				} else {
					break
				}
			}
			title := strings.TrimSpace(trimmed[level:])

			for len(headingStack) > 0 && headingStack[len(headingStack)-1].level >= level {
				headingStack = headingStack[:len(headingStack)-1]
			}

			seq++
			s := Section{
				ID:          NewID().String(),
				Ref:         NewRef(docID, seq),
				DocID:       docID,
				Type:        SectionHeading,
				Title:       title,
				Content:     title,
				ContentHash: ContentHash(title),
				Order:       len(sections),
			}
			if len(headingStack) > 0 {
				ref := sections[headingStack[len(headingStack)-1].idx].Ref
				s.ParentRef = &ref
			}

			sections = append(sections, s)
			idx := len(sections) - 1
			headingStack = append(headingStack, headingEntry{idx: idx, level: level})
			currentHeadingIdx = idx
			tableState = 0
			listStack = nil
			continue
		}

		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			checked := strings.HasPrefix(trimmed, "- [x] ")
			title := trimmed[6:]
			indent := countIndent(line)

			seq++
			s := Section{
				ID:          NewID().String(),
				Ref:         NewRef(docID, seq),
				DocID:       docID,
				Type:        SectionChecklist,
				Title:       title,
				Content:     trimmed,
				ContentHash: ContentHash(trimmed),
				Checked:     checked,
				Order:       len(sections),
			}

			parent := findListParent(listStack, indent, sections)
			if parent == nil && currentHeadingIdx >= 0 {
				ref := sections[currentHeadingIdx].Ref
				parent = &ref
			}
			s.ParentRef = parent

			sections = append(sections, s)
			idx := len(sections) - 1
			listStack = append(listStack, listEntry{idx: idx, indent: indent})
			tableState = 0
			continue
		}

		if isListItem(trimmed) {
			title := extractListTitle(trimmed)
			indent := countIndent(line)

			for len(listStack) > 0 && listStack[len(listStack)-1].indent >= indent {
				listStack = listStack[:len(listStack)-1]
			}

			seq++
			s := Section{
				ID:          NewID().String(),
				Ref:         NewRef(docID, seq),
				DocID:       docID,
				Type:        SectionListItem,
				Title:       title,
				Content:     trimmed,
				ContentHash: ContentHash(trimmed),
				Order:       len(sections),
			}

			if len(listStack) > 0 {
				ref := sections[listStack[len(listStack)-1].idx].Ref
				s.ParentRef = &ref
			} else if currentHeadingIdx >= 0 {
				ref := sections[currentHeadingIdx].Ref
				s.ParentRef = &ref
			}

			sections = append(sections, s)
			idx := len(sections) - 1
			listStack = append(listStack, listEntry{idx: idx, indent: indent})
			tableState = 0
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			switch tableState {
			case 0:
				tableState = 1
				continue
			case 1:
				if isTableSeparator(trimmed) {
					tableState = 2
					continue
				}
				tableState = 0
			case 2:
				title := parseTableRow(trimmed)
				seq++
				s := Section{
					ID:          NewID().String(),
					Ref:         NewRef(docID, seq),
					DocID:       docID,
					Type:        SectionTableRow,
					Title:       title,
					Content:     trimmed,
					ContentHash: ContentHash(trimmed),
					Order:       len(sections),
				}
				if currentHeadingIdx >= 0 {
					ref := sections[currentHeadingIdx].Ref
					s.ParentRef = &ref
				}
				sections = append(sections, s)
			}
			continue
		}

		bodyLines = append(bodyLines, line)
		tableState = 0
	}

	flushBody(sections, currentHeadingIdx, &bodyLines)
	return sections
}

// flushBody appends accumulated body lines to the heading section's Content.
func flushBody(sections []Section, headingIdx int, bodyLines *[]string) {
	if headingIdx < 0 || len(*bodyLines) == 0 {
		*bodyLines = nil
		return
	}
	// Trim trailing empty lines
	body := *bodyLines
	for len(body) > 0 && body[len(body)-1] == "" {
		body = body[:len(body)-1]
	}
	if len(body) == 0 {
		*bodyLines = nil
		return
	}
	joined := strings.Join(body, "\n")
	sections[headingIdx].Content = sections[headingIdx].Title + "\n\n" + joined
	sections[headingIdx].ContentHash = ContentHash(sections[headingIdx].Content)
	*bodyLines = nil
}

// countIndent returns the number of leading spaces on a line.
func countIndent(line string) int {
	n := 0
	for _, c := range line {
		if c == ' ' {
			n++
		} else if c == '\t' {
			n += 4
		} else {
			break
		}
	}
	return n
}

// isListItem checks if a trimmed line starts with a list marker.
func isListItem(trimmed string) bool {
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return true
	}
	for i, c := range trimmed {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' && i > 0 && strings.HasPrefix(trimmed[i:], ". ") {
			return true
		}
		break
	}
	return false
}

// extractListTitle extracts the text after the list marker.
func extractListTitle(trimmed string) string {
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return trimmed[2:]
	}
	idx := strings.Index(trimmed, ". ")
	if idx > 0 {
		return trimmed[idx+2:]
	}
	return trimmed
}

// isTableSeparator checks if a line is a markdown table separator (|---|---|).
func isTableSeparator(trimmed string) bool {
	inner := strings.Trim(trimmed, "| ")
	for _, c := range inner {
		if c != '-' && c != '|' && c != ' ' && c != ':' {
			return false
		}
	}
	return len(inner) > 0
}

// parseTableRow extracts cell contents from a table row, joined by " | ".
func parseTableRow(trimmed string) string {
	inner := strings.Trim(trimmed, "| ")
	parts := strings.Split(inner, "|")
	var cells []string
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	return strings.Join(cells, " | ")
}

// RebaseRefs shifts all section refs by an offset so that the first section
// starts at newStart instead of its current seq. ParentRefs are updated to match.
func RebaseRefs(sections []Section, newStart int) {
	if len(sections) == 0 {
		return
	}
	offset := newStart - sections[0].Ref.Seq
	if offset == 0 {
		return
	}
	// Build old→new ref mapping
	oldToNew := make(map[string]Ref, len(sections))
	for i := range sections {
		oldRef := sections[i].Ref
		newSeq := oldRef.Seq + offset
		newRef := NewRef(sections[i].DocID, newSeq)
		oldToNew[oldRef.String()] = newRef
		sections[i].Ref = newRef
	}
	// Fix parent refs
	for i := range sections {
		if sections[i].ParentRef != nil {
			if nr, ok := oldToNew[sections[i].ParentRef.String()]; ok {
				sections[i].ParentRef = &nr
			}
		}
	}
}

// findListParent finds the nearest list parent with indent less than the given indent.
func findListParent(listStack []listEntry, indent int, sections []Section) *Ref {
	for i := len(listStack) - 1; i >= 0; i-- {
		if listStack[i].indent < indent {
			ref := sections[listStack[i].idx].Ref
			return &ref
		}
	}
	return nil
}
