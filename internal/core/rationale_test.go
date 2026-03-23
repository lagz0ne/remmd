package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestRationale_Validate_Valid(t *testing.T) {
	t.Parallel()
	r := core.Rationale{
		Claim:      "API contract matches implementation",
		Scope:      "All public endpoints",
		Exclusions: "",
	}
	if err := r.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid rationale: %v", err)
	}
}

func TestRationale_Validate_WithExclusions(t *testing.T) {
	t.Parallel()
	r := core.Rationale{
		Claim:      "Test coverage",
		Scope:      "Unit tests",
		Exclusions: "integration tests excluded",
	}
	if err := r.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid rationale with exclusions: %v", err)
	}
}

func TestRationale_Validate_EmptyClaim(t *testing.T) {
	t.Parallel()
	r := core.Rationale{Claim: "", Scope: "all", Exclusions: ""}
	if err := r.Validate(); err == nil {
		t.Error("Validate() should reject empty claim")
	}
}

func TestRationale_Validate_WhitespaceClaim(t *testing.T) {
	t.Parallel()
	r := core.Rationale{Claim: "   ", Scope: "all", Exclusions: ""}
	if err := r.Validate(); err == nil {
		t.Error("Validate() should reject whitespace-only claim")
	}
}

func TestRationale_Validate_EmptyScope(t *testing.T) {
	t.Parallel()
	r := core.Rationale{Claim: "valid claim", Scope: "", Exclusions: ""}
	if err := r.Validate(); err == nil {
		t.Error("Validate() should reject empty scope")
	}
}

func TestRationale_Validate_BothEmpty(t *testing.T) {
	t.Parallel()
	r := core.Rationale{Claim: "", Scope: "", Exclusions: ""}
	err := r.Validate()
	if err == nil {
		t.Error("Validate() should reject when both claim and scope are empty")
	}
}
