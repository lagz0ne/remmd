package core

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// ContentHash returns the SHA-256 hex digest of whitespace-normalized content.
func ContentHash(content string) string {
	normalized := normalizeWhitespace(content)
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)
}

// normalizeWhitespace strips trailing spaces per line, collapses consecutive
// blank lines to a single blank line, and normalizes CRLF to LF.
func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	var result []string
	prevBlank := false
	for _, line := range lines {
		blank := line == ""
		if blank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = blank
	}

	return strings.Join(result, "\n")
}
