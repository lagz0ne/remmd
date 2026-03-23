package store_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

func TestMigrate_CreatesSchemaVersionTable(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// schema_version table should exist
	var count int
	err = db.QueryRow("SELECT count(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("schema_version table does not exist: %v", err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("first Migrate error: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("second Migrate error: %v", err)
	}
}

func TestMigrate_EventsTableWorks(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// INSERT should work
	_, err = db.Exec(`INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, principal_id, sequence)
		VALUES ('evt-1', 'agg-1', 'Document', 'DocumentCreated', '{}', 'user-1', 1)`)
	if err != nil {
		t.Fatalf("INSERT into events failed: %v", err)
	}

	// SELECT should return the row
	var id, aggID, aggType, evtType, payload, principal string
	var seq int
	err = db.QueryRow("SELECT id, aggregate_id, aggregate_type, event_type, payload, principal_id, sequence FROM events WHERE id = 'evt-1'").
		Scan(&id, &aggID, &aggType, &evtType, &payload, &principal, &seq)
	if err != nil {
		t.Fatalf("SELECT from events failed: %v", err)
	}
	if id != "evt-1" || aggID != "agg-1" || seq != 1 {
		t.Fatalf("unexpected row: id=%s aggID=%s seq=%d", id, aggID, seq)
	}
}

func TestMigrate_SchemaVersionTracked(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("query schema_version failed: %v", err)
	}
	if version < 1 {
		t.Fatalf("expected schema version >= 1, got %d", version)
	}
}
