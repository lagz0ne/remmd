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
type Ref struct {
	Letters string
	Seq     int
}

// NewRef creates a Ref from a document ID and a 1-based sequence number.
// The letter component is base-26 (a-z) encoding of the sequence.
func NewRef(_ string, seq int) Ref {
	return Ref{Letters: toBase26(seq), Seq: seq}
}

func (r Ref) String() string {
	return fmt.Sprintf("@%s%d", r.Letters, r.Seq)
}

// ParseRef parses a string like "@a1" into a Ref.
// Returns error if format is invalid.
func ParseRef(s string) (Ref, error) {
	if !strings.HasPrefix(s, "@") || len(s) < 3 {
		return Ref{}, fmt.Errorf("invalid ref %q: must start with @ followed by letters and digits", s)
	}
	body := s[1:]

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
