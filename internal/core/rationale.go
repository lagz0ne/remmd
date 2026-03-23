package core

import (
	"errors"
	"strings"
)

// Rationale captures the justification for a link: what is claimed, what scope
// it covers, and what is explicitly excluded.
type Rationale struct {
	Claim      string
	Scope      string
	Exclusions string
}

// Validate checks that required fields are non-empty.
func (r Rationale) Validate() error {
	var errs []error
	if strings.TrimSpace(r.Claim) == "" {
		errs = append(errs, errors.New("rationale claim is required"))
	}
	if strings.TrimSpace(r.Scope) == "" {
		errs = append(errs, errors.New("rationale scope is required"))
	}
	// Exclusions may be empty — means "nothing excluded".
	return errors.Join(errs...)
}
