package core

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// AgreementSnapshot captures the content hashes of both sides of a link
// at a point in time, used for optimistic concurrency on review actions.
type AgreementSnapshot struct {
	LinkID             string
	LeftContentHashes  []string
	RightContentHashes []string
}

// Hash returns a SHA-256 hex digest of the concatenated left and right
// content hashes, separated by a boundary marker to prevent collisions.
func (s AgreementSnapshot) Hash() string {
	var b strings.Builder
	for _, h := range s.LeftContentHashes {
		b.WriteString(h)
		b.WriteByte(':')
	}
	b.WriteString("|")
	for _, h := range s.RightContentHashes {
		b.WriteString(h)
		b.WriteByte(':')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", sum)
}
