package core_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestErrNotFound_Error(t *testing.T) {
	t.Parallel()
	err := core.ErrNotFound{Entity: "document", ID: "abc"}
	got := err.Error()
	want := `document "abc" not found`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrStaleContext_Error(t *testing.T) {
	t.Parallel()
	err := core.ErrStaleContext{
		LinkID:   "link-42",
		Expected: "v3",
		Actual:   "v5",
	}
	got := err.Error()
	want := `stale context on link "link-42": expected version "v3", got "v5"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrUnauthorized_Error(t *testing.T) {
	t.Parallel()
	err := core.ErrUnauthorized{Action: "approve", PrincipalType: "service"}
	got := err.Error()
	want := `action "approve" requires human principal, got service`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrConflict_Error(t *testing.T) {
	t.Parallel()
	err := core.ErrConflict{Entity: "agreement", ID: "agr-1", Reason: "already verified"}
	got := err.Error()
	want := `conflict on agreement "agr-1": already verified`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrors_Is(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		target error
	}{
		{
			name:   "ErrNotFound wrapped",
			err:    fmt.Errorf("loading: %w", core.ErrNotFound{Entity: "doc", ID: "x"}),
			target: core.ErrNotFound{},
		},
		{
			name:   "ErrStaleContext wrapped",
			err:    fmt.Errorf("cmd: %w", core.ErrStaleContext{LinkID: "l1", Expected: "a", Actual: "b"}),
			target: core.ErrStaleContext{},
		},
		{
			name:   "ErrUnauthorized wrapped",
			err:    fmt.Errorf("handler: %w", core.ErrUnauthorized{Action: "delete", PrincipalType: "bot"}),
			target: core.ErrUnauthorized{},
		},
		{
			name:   "ErrConflict wrapped",
			err:    fmt.Errorf("store: %w", core.ErrConflict{Entity: "doc", ID: "d1", Reason: "version mismatch"}),
			target: core.ErrConflict{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !errors.Is(tt.err, tt.target) {
				t.Errorf("errors.Is(%v, %T{}) = false, want true", tt.err, tt.target)
			}
		})
	}
}

func TestErrors_As(t *testing.T) {
	t.Parallel()

	t.Run("ErrNotFound", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("loading: %w", core.ErrNotFound{Entity: "document", ID: "abc"})
		var target core.ErrNotFound
		if !errors.As(wrapped, &target) {
			t.Fatal("errors.As failed")
		}
		if target.Entity != "document" || target.ID != "abc" {
			t.Errorf("got entity=%q id=%q", target.Entity, target.ID)
		}
	})

	t.Run("ErrStaleContext", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("cmd: %w", core.ErrStaleContext{LinkID: "l1", Expected: "v1", Actual: "v2"})
		var target core.ErrStaleContext
		if !errors.As(wrapped, &target) {
			t.Fatal("errors.As failed")
		}
		if target.LinkID != "l1" {
			t.Errorf("got linkID=%q", target.LinkID)
		}
	})
}
