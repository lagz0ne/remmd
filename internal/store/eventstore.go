package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
)

// EventStore provides append-only storage for domain events.
// No Update or Delete methods — this is by design.
type EventStore struct {
	db *sql.DB
}

// NewEventStore creates an EventStore backed by the given *sql.DB.
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

// Append inserts an event into the store, computing the next sequence number
// for the event's aggregate.
func (s *EventStore) Append(ctx context.Context, evt core.Event) error {
	start := time.Now()

	slog.DebugContext(ctx, "appending domain event",
		"op", "event.append",
		"event_type", evt.EventType,
		"aggregate_id", evt.AggregateID,
		"aggregate_type", evt.AggregateType,
		"principal_id", evt.PrincipalID,
	)

	if err := evt.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, principal_id, created_at, sequence)
		 VALUES (?, ?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sequence), 0) + 1 FROM events WHERE aggregate_id = ?))`,
		evt.ID, evt.AggregateID, evt.AggregateType, evt.EventType,
		string(evt.Payload), evt.PrincipalID, formatTime(evt.CreatedAt), evt.AggregateID,
	)
	if err != nil {
		slog.ErrorContext(ctx, "domain event append failed",
			"op", "event.append",
			"event_type", evt.EventType,
			"aggregate_id", evt.AggregateID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return fmt.Errorf("insert event: %w", err)
	}

	slog.InfoContext(ctx, "domain event appended",
		"op", "event.append",
		"event_type", evt.EventType,
		"aggregate_id", evt.AggregateID,
		"outcome", "ok",
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

// ReadStream returns all events for an aggregate, ordered by sequence.
// Returns an empty (non-nil) slice if the aggregate has no events.
func (s *EventStore) ReadStream(ctx context.Context, aggregateID string) ([]core.Event, error) {
	start := time.Now()

	slog.DebugContext(ctx, "reading event stream",
		"op", "event.read_stream",
		"aggregate_id", aggregateID,
	)

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, aggregate_id, aggregate_type, event_type, payload, principal_id, created_at, sequence
		 FROM events WHERE aggregate_id = ? ORDER BY sequence`,
		aggregateID,
	)
	if err != nil {
		slog.ErrorContext(ctx, "event stream read failed",
			"op", "event.read_stream",
			"aggregate_id", aggregateID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	events := make([]core.Event, 0)
	for rows.Next() {
		var evt core.Event
		var createdAt string
		var payload string
		if err := rows.Scan(
			&evt.ID, &evt.AggregateID, &evt.AggregateType, &evt.EventType,
			&payload, &evt.PrincipalID, &createdAt, &evt.Sequence,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		evt.Payload = []byte(payload)
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "event stream read failed",
			"op", "event.read_stream",
			"aggregate_id", aggregateID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, err
	}

	slog.InfoContext(ctx, "event stream read",
		"op", "event.read_stream",
		"aggregate_id", aggregateID,
		"outcome", "ok",
		"count", len(events),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return events, nil
}
