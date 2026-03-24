package core

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropy     = ulid.Monotonic(rand.Reader, 0)
	entropyLock sync.Mutex
)

// ID is a globally unique identifier backed by ULID.
type ID struct {
	val ulid.ULID
}

// NewID generates a new globally unique ID.
func NewID() ID {
	entropyLock.Lock()
	defer entropyLock.Unlock()
	return ID{val: ulid.MustNew(ulid.Timestamp(time.Now()), entropy)}
}

func (id ID) String() string {
	return id.val.String()
}

// Ref is a document-scoped, human-readable section reference.
// Format: @{letters}{number}, e.g. @a1, @b3, @aa27.
// External refs use format @ext:system/id.
type Ref struct {
	Letters  string
	Seq      int
	External string
}

// NewRef creates a Ref from a document ID and a 1-based sequence number.
// The letter component is base-26 (a-z) encoding of the sequence.
func NewRef(_ string, seq int) Ref {
	return Ref{Letters: toBase26(seq), Seq: seq}
}

func (r Ref) String() string {
	if r.External != "" {
		return "@" + r.External
	}
	return fmt.Sprintf("@%s%d", r.Letters, r.Seq)
}

// NewExternalRef creates a Ref pointing to an external system resource.
func NewExternalRef(system, externalID string) Ref {
	return Ref{External: fmt.Sprintf("ext:%s/%s", system, externalID)}
}

// IsExternal returns true if this ref points to an external system.
func (r Ref) IsExternal() bool {
	return r.External != ""
}

// System extracts the system name from an external ref (e.g. "notion" from "ext:notion/page-abc").
func (r Ref) System() string {
	if !r.IsExternal() {
		return ""
	}
	after := strings.TrimPrefix(r.External, "ext:")
	idx := strings.Index(after, "/")
	if idx < 0 {
		return ""
	}
	return after[:idx]
}

// ExternalID extracts the ID from an external ref (e.g. "page-abc" from "ext:notion/page-abc").
func (r Ref) ExternalID() string {
	if !r.IsExternal() {
		return ""
	}
	after := strings.TrimPrefix(r.External, "ext:")
	idx := strings.Index(after, "/")
	if idx < 0 {
		return ""
	}
	return after[idx+1:]
}

// ParseRef parses a string like "@a1" into a Ref.
// Returns error if format is invalid.
func ParseRef(s string) (Ref, error) {
	if !strings.HasPrefix(s, "@") || len(s) < 3 {
		return Ref{}, fmt.Errorf("invalid ref %q: must start with @ followed by letters and digits", s)
	}
	body := s[1:]

	if strings.HasPrefix(body, "ext:") {
		rest := body[4:]
		slashIdx := strings.Index(rest, "/")
		if slashIdx <= 0 || slashIdx == len(rest)-1 {
			return Ref{}, fmt.Errorf("invalid external ref %q: must be @ext:<system>/<id>", s)
		}
		return Ref{External: body}, nil
	}

	splitIdx := -1
	for i, c := range body {
		if c >= '0' && c <= '9' {
			splitIdx = i
			break
		}
	}
	if splitIdx <= 0 {
		return Ref{}, fmt.Errorf("invalid ref %q: must contain lowercase letters followed by digits", s)
	}

	letters := body[:splitIdx]
	digits := body[splitIdx:]

	for _, c := range letters {
		if c < 'a' || c > 'z' {
			return Ref{}, fmt.Errorf("invalid ref %q: letters must be lowercase a-z", s)
		}
	}

	var seq int
	for _, c := range digits {
		if c < '0' || c > '9' {
			return Ref{}, fmt.Errorf("invalid ref %q: trailing non-digit character", s)
		}
		seq = seq*10 + int(c-'0')
	}
	if seq == 0 {
		return Ref{}, fmt.Errorf("invalid ref %q: sequence must be > 0", s)
	}

	return Ref{Letters: letters, Seq: seq}, nil
}

// toBase26 converts a 1-based sequence to base-26 letters (a=1, z=26, aa=27).
func toBase26(n int) string {
	var buf []byte
	for n > 0 {
		n-- // shift to 0-based
		buf = append(buf, byte('a'+n%26))
		n /= 26
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
