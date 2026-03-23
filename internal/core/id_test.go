package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestNewID_Unique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := core.NewID()
		s := id.String()
		if seen[s] {
			t.Fatalf("duplicate ID at iteration %d: %s", i, s)
		}
		seen[s] = true
	}
}

func TestNewID_NonEmpty(t *testing.T) {
	t.Parallel()
	id := core.NewID()
	if id.String() == "" {
		t.Fatal("NewID returned empty string")
	}
}

func TestNewRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		docID string
		seq   int
		want  string
	}{
		{"doc1", 1, "@a1"},
		{"doc1", 2, "@b2"},
		{"doc1", 3, "@c3"},
		{"doc1", 26, "@z26"},
		{"doc1", 27, "@aa27"},
		{"doc1", 28, "@ab28"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			ref := core.NewRef(tt.docID, tt.seq)
			if ref.String() != tt.want {
				t.Errorf("NewRef(%q, %d).String() = %q, want %q", tt.docID, tt.seq, ref.String(), tt.want)
			}
		})
	}
}

func TestParseRef_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"@a1", "@a1"},
		{"@b3", "@b3"},
		{"@z26", "@z26"},
		{"@aa27", "@aa27"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			ref, err := core.ParseRef(tt.input)
			if err != nil {
				t.Fatalf("ParseRef(%q) error: %v", tt.input, err)
			}
			if ref.String() != tt.want {
				t.Errorf("ParseRef(%q).String() = %q, want %q", tt.input, ref.String(), tt.want)
			}
		})
	}
}

func TestParseRef_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
	}{
		{"invalid"},
		{""},
		{"@"},
		{"@1"},
		{"@A1"},
		{"hello"},
		{"@abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			_, err := core.ParseRef(tt.input)
			if err == nil {
				t.Errorf("ParseRef(%q) should have returned error", tt.input)
			}
		})
	}
}

func TestRef_RoundTrip(t *testing.T) {
	t.Parallel()
	ref := core.NewRef("doc1", 5)
	parsed, err := core.ParseRef(ref.String())
	if err != nil {
		t.Fatalf("round-trip failed: %v", err)
	}
	if parsed.String() != ref.String() {
		t.Errorf("round-trip mismatch: %q != %q", parsed.String(), ref.String())
	}
}
