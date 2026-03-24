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

// --- External ref tests (RED — types/methods don't exist yet) ---

func TestParseRef_External(t *testing.T) {
	t.Parallel()
	ref, err := core.ParseRef("@ext:notion/page-abc")
	if err != nil {
		t.Fatalf("ParseRef(@ext:notion/page-abc) error: %v", err)
	}
	if ref.External != "ext:notion/page-abc" {
		t.Errorf("External = %q, want %q", ref.External, "ext:notion/page-abc")
	}
	if !ref.IsExternal() {
		t.Error("IsExternal() = false, want true")
	}
	if ref.System() != "notion" {
		t.Errorf("System() = %q, want %q", ref.System(), "notion")
	}
	if ref.ExternalID() != "page-abc" {
		t.Errorf("ExternalID() = %q, want %q", ref.ExternalID(), "page-abc")
	}
}

func TestParseRef_ExternalRoundTrip(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"@ext:notion/page-abc",
		"@ext:figma/frame-123",
		"@ext:jira/PROJ-456",
		"@ext:confluence/space/page-id",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			ref, err := core.ParseRef(input)
			if err != nil {
				t.Fatalf("ParseRef(%q) error: %v", input, err)
			}
			roundTripped, err := core.ParseRef(ref.String())
			if err != nil {
				t.Fatalf("round-trip ParseRef(%q) error: %v", ref.String(), err)
			}
			if roundTripped.String() != input {
				t.Errorf("round-trip mismatch: got %q, want %q", roundTripped.String(), input)
			}
		})
	}
}

func TestParseRef_ExternalInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"empty_after_ext", "@ext:"},
		{"no_slash", "@ext:no-slash"},
		{"no_system", "@ext:/no-system"},
		{"no_id", "@ext:sys/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := core.ParseRef(tt.input)
			if err == nil {
				t.Errorf("ParseRef(%q) should have returned error", tt.input)
			}
		})
	}
}

func TestNewExternalRef(t *testing.T) {
	t.Parallel()
	ref := core.NewExternalRef("figma", "frame-123")
	if ref.String() != "@ext:figma/frame-123" {
		t.Errorf("String() = %q, want %q", ref.String(), "@ext:figma/frame-123")
	}
	if !ref.IsExternal() {
		t.Error("IsExternal() = false, want true")
	}
	if ref.System() != "figma" {
		t.Errorf("System() = %q, want %q", ref.System(), "figma")
	}
	if ref.ExternalID() != "frame-123" {
		t.Errorf("ExternalID() = %q, want %q", ref.ExternalID(), "frame-123")
	}
}

func TestRef_IsExternal(t *testing.T) {
	t.Parallel()

	t.Run("native ref returns false", func(t *testing.T) {
		t.Parallel()
		ref := core.NewRef("doc1", 1)
		if ref.IsExternal() {
			t.Error("native ref: IsExternal() = true, want false")
		}
	})

	t.Run("external ref returns true", func(t *testing.T) {
		t.Parallel()
		ref := core.NewExternalRef("notion", "page-1")
		if !ref.IsExternal() {
			t.Error("external ref: IsExternal() = false, want true")
		}
	})
}

func TestRef_NativeUnchanged(t *testing.T) {
	t.Parallel()
	// Regression guard: native ParseRef still works after external ref support
	ref, err := core.ParseRef("@a1")
	if err != nil {
		t.Fatalf("ParseRef(@a1) error: %v", err)
	}
	if ref.String() != "@a1" {
		t.Errorf("String() = %q, want %q", ref.String(), "@a1")
	}
	if ref.IsExternal() {
		t.Error("native ref @a1: IsExternal() = true, want false")
	}
	if ref.Letters != "a" {
		t.Errorf("Letters = %q, want %q", ref.Letters, "a")
	}
	if ref.Seq != 1 {
		t.Errorf("Seq = %d, want %d", ref.Seq, 1)
	}
}
