package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestTransition_PendingToAligned(t *testing.T) {
	t.Parallel()
	if err := core.Transition(core.LinkPending, core.LinkAligned); err != nil {
		t.Errorf("pending -> aligned should be valid: %v", err)
	}
}

func TestTransition_AlignedToStale(t *testing.T) {
	t.Parallel()
	if err := core.Transition(core.LinkAligned, core.LinkStale); err != nil {
		t.Errorf("aligned -> stale should be valid: %v", err)
	}
}

func TestTransition_StaleToAligned(t *testing.T) {
	t.Parallel()
	if err := core.Transition(core.LinkStale, core.LinkAligned); err != nil {
		t.Errorf("stale -> aligned should be valid: %v", err)
	}
}

func TestTransition_AnyToArchived(t *testing.T) {
	t.Parallel()
	for _, from := range []core.LinkState{core.LinkPending, core.LinkAligned, core.LinkStale, core.LinkBroken} {
		if err := core.Transition(from, core.LinkArchived); err != nil {
			t.Errorf("%s -> archived should be valid: %v", from, err)
		}
	}
}

func TestTransition_AnyToBroken(t *testing.T) {
	t.Parallel()
	for _, from := range []core.LinkState{core.LinkPending, core.LinkAligned, core.LinkStale} {
		if err := core.Transition(from, core.LinkBroken); err != nil {
			t.Errorf("%s -> broken should be valid: %v", from, err)
		}
	}
}

func TestTransition_AlignedToPending_Invalid(t *testing.T) {
	t.Parallel()
	if err := core.Transition(core.LinkAligned, core.LinkPending); err == nil {
		t.Error("aligned -> pending should be invalid")
	}
}

func TestTransition_BrokenToAligned_Invalid(t *testing.T) {
	t.Parallel()
	if err := core.Transition(core.LinkBroken, core.LinkAligned); err == nil {
		t.Error("broken -> aligned should be invalid")
	}
}

func TestTransition_ArchivedToAnything_Invalid(t *testing.T) {
	t.Parallel()
	for _, to := range []core.LinkState{core.LinkPending, core.LinkAligned, core.LinkStale, core.LinkBroken} {
		if err := core.Transition(core.LinkArchived, to); err == nil {
			t.Errorf("archived -> %s should be invalid", to)
		}
	}
}
