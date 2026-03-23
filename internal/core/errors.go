package core

import "fmt"

// ErrNotFound indicates a missing entity lookup.
type ErrNotFound struct {
	Entity string
	ID     string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %q not found", e.Entity, e.ID)
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound)
	return ok
}

// ErrStaleContext indicates a version mismatch on a link.
type ErrStaleContext struct {
	LinkID   string
	Expected string
	Actual   string
}

func (e ErrStaleContext) Error() string {
	return fmt.Sprintf("stale context on link %q: expected version %q, got %q", e.LinkID, e.Expected, e.Actual)
}

func (e ErrStaleContext) Is(target error) bool {
	_, ok := target.(ErrStaleContext)
	return ok
}

// ErrUnauthorized indicates a principal type mismatch for an action.
type ErrUnauthorized struct {
	Action        string
	PrincipalType string
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("action %q requires human principal, got %s", e.Action, e.PrincipalType)
}

func (e ErrUnauthorized) Is(target error) bool {
	_, ok := target.(ErrUnauthorized)
	return ok
}

// ErrConflict indicates a state conflict on an entity.
type ErrConflict struct {
	Entity string
	ID     string
	Reason string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("conflict on %s %q: %s", e.Entity, e.ID, e.Reason)
}

func (e ErrConflict) Is(target error) bool {
	_, ok := target.(ErrConflict)
	return ok
}
