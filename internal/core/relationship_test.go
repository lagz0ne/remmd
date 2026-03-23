package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestValidRelationshipType_AllFourTypes(t *testing.T) {
	t.Parallel()
	for _, rt := range []string{"agrees_with", "implements", "tests", "evidences"} {
		if !core.ValidRelationshipType(rt) {
			t.Errorf("ValidRelationshipType(%q) = false, want true", rt)
		}
	}
}

func TestValidRelationshipType_RejectsUnknown(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "depends_on", "IMPLEMENTS", "foo"} {
		if core.ValidRelationshipType(bad) {
			t.Errorf("ValidRelationshipType(%q) = true, want false", bad)
		}
	}
}
