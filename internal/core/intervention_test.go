package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestValidInterventionLevel_AllFourLevels(t *testing.T) {
	t.Parallel()
	for _, il := range []string{"watch", "notify", "urgent", "blocking"} {
		if !core.ValidInterventionLevel(il) {
			t.Errorf("ValidInterventionLevel(%q) = false, want true", il)
		}
	}
}

func TestValidInterventionLevel_RejectsUnknown(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "WATCH", "alert", "gate", "lock"} {
		if core.ValidInterventionLevel(bad) {
			t.Errorf("ValidInterventionLevel(%q) = true, want false", bad)
		}
	}
}
