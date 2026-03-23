package core

import "fmt"

// LinkState represents the current trust state of a link.
type LinkState string

const (
	LinkPending  LinkState = "pending"
	LinkAligned  LinkState = "aligned"
	LinkStale    LinkState = "stale"
	LinkBroken   LinkState = "broken"
	LinkArchived LinkState = "archived"
)

// validTransitions maps each state to the set of states it can transition to.
var validTransitions = map[LinkState]map[LinkState]bool{
	LinkPending:  {LinkAligned: true, LinkArchived: true, LinkBroken: true},
	LinkAligned:  {LinkStale: true, LinkArchived: true, LinkBroken: true},
	LinkStale:    {LinkAligned: true, LinkArchived: true, LinkBroken: true},
	LinkBroken:   {LinkArchived: true},
	LinkArchived: {},
}

// Transition validates a state transition from -> to.
// Returns nil if valid, or an error describing the invalid transition.
func Transition(from, to LinkState) error {
	targets, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("unknown link state %q", from)
	}
	if !targets[to] {
		return fmt.Errorf("invalid transition: %s -> %s", from, to)
	}
	return nil
}
