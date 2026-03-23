package core

import (
	"encoding/json"
	"errors"
	"time"
)

// Event represents a single domain event in the event store.
type Event struct {
	ID            string          `json:"id"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	PrincipalID   string          `json:"principal_id"`
	CreatedAt     time.Time       `json:"created_at"`
	Sequence      int64           `json:"sequence"`
}

// NewEvent creates a new Event with an auto-generated ULID and current timestamp.
// Sequence is left at 0 — it is assigned by the event store on append.
func NewEvent(aggregateID, aggregateType, eventType string, payload json.RawMessage, principalID string) Event {
	return Event{
		ID:            NewID().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payload,
		PrincipalID:   principalID,
		CreatedAt:     time.Now(),
	}
}

// Validate checks required fields on the event.
func (e Event) Validate() error {
	var errs []error
	if e.AggregateID == "" {
		errs = append(errs, errors.New("aggregate_id is required"))
	}
	if e.EventType == "" {
		errs = append(errs, errors.New("event_type is required"))
	}
	if e.PrincipalID == "" {
		errs = append(errs, errors.New("principal_id is required"))
	}
	return errors.Join(errs...)
}
