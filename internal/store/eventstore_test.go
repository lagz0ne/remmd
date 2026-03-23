package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE events (
		id TEXT PRIMARY KEY,
		aggregate_id TEXT NOT NULL,
		aggregate_type TEXT NOT NULL,
		event_type TEXT NOT NULL,
		payload TEXT NOT NULL,
		principal_id TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		sequence INTEGER NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create events table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewEventStore_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	es := store.NewEventStore(db)
	if es == nil {
		t.Fatal("NewEventStore returned nil")
	}
}

func TestEventStore_AppendAndReadStream(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	es := store.NewEventStore(db)
	ctx := context.Background()

	evt := core.NewEvent("agg-1", "Document", "DocumentCreated", json.RawMessage(`{"title":"hi"}`), "user-1")
	if err := es.Append(ctx, evt); err != nil {
		t.Fatalf("Append error: %v", err)
	}

	events, err := es.ReadStream(ctx, "agg-1")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != evt.ID {
		t.Errorf("ID = %q, want %q", events[0].ID, evt.ID)
	}
	if events[0].AggregateID != "agg-1" {
		t.Errorf("AggregateID = %q, want %q", events[0].AggregateID, "agg-1")
	}
	if events[0].EventType != "DocumentCreated" {
		t.Errorf("EventType = %q, want %q", events[0].EventType, "DocumentCreated")
	}
	if string(events[0].Payload) != `{"title":"hi"}` {
		t.Errorf("Payload = %s, want %s", events[0].Payload, `{"title":"hi"}`)
	}
	if events[0].Sequence != 1 {
		t.Errorf("Sequence = %d, want 1", events[0].Sequence)
	}
}

func TestEventStore_ReadStream_OrderedBySequence(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	es := store.NewEventStore(db)
	ctx := context.Background()

	types := []string{"Created", "Updated", "Published"}
	for _, et := range types {
		evt := core.NewEvent("agg-1", "Document", et, json.RawMessage(`{}`), "user-1")
		if err := es.Append(ctx, evt); err != nil {
			t.Fatalf("Append %s error: %v", et, err)
		}
	}

	events, err := es.ReadStream(ctx, "agg-1")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	for i, want := range types {
		if events[i].EventType != want {
			t.Errorf("events[%d].EventType = %q, want %q", i, events[i].EventType, want)
		}
		if events[i].Sequence != int64(i+1) {
			t.Errorf("events[%d].Sequence = %d, want %d", i, events[i].Sequence, i+1)
		}
	}
}

func TestEventStore_ReadStream_NonExistentAggregate(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	es := store.NewEventStore(db)
	ctx := context.Background()

	events, err := es.ReadStream(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("ReadStream error: %v", err)
	}
	if events == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestEventStore_SequenceAutoIncrementsPerAggregate(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	es := store.NewEventStore(db)
	ctx := context.Background()

	// Two events on agg-A
	for i := 0; i < 2; i++ {
		evt := core.NewEvent("agg-A", "Doc", "E", json.RawMessage(`{}`), "u")
		if err := es.Append(ctx, evt); err != nil {
			t.Fatalf("Append agg-A #%d error: %v", i, err)
		}
	}
	// Three events on agg-B
	for i := 0; i < 3; i++ {
		evt := core.NewEvent("agg-B", "Doc", "E", json.RawMessage(`{}`), "u")
		if err := es.Append(ctx, evt); err != nil {
			t.Fatalf("Append agg-B #%d error: %v", i, err)
		}
	}

	evtsA, _ := es.ReadStream(ctx, "agg-A")
	evtsB, _ := es.ReadStream(ctx, "agg-B")

	if len(evtsA) != 2 {
		t.Fatalf("agg-A: expected 2, got %d", len(evtsA))
	}
	if len(evtsB) != 3 {
		t.Fatalf("agg-B: expected 3, got %d", len(evtsB))
	}

	// agg-A sequences: 1, 2
	for i, want := range []int64{1, 2} {
		if evtsA[i].Sequence != want {
			t.Errorf("agg-A[%d].Sequence = %d, want %d", i, evtsA[i].Sequence, want)
		}
	}
	// agg-B sequences: 1, 2, 3
	for i, want := range []int64{1, 2, 3} {
		if evtsB[i].Sequence != want {
			t.Errorf("agg-B[%d].Sequence = %d, want %d", i, evtsB[i].Sequence, want)
		}
	}
}

func TestEventStore_AppendOnlyInvariant(t *testing.T) {
	t.Parallel()
	// Verify that EventStore has no Update or Delete methods.
	// This is a design constraint — the event store is append-only.
	esType := reflect.TypeOf(&store.EventStore{})

	for i := 0; i < esType.NumMethod(); i++ {
		name := esType.Method(i).Name
		if name == "Update" || name == "Delete" {
			t.Errorf("EventStore should not have %s method — append-only invariant violated", name)
		}
	}
}
