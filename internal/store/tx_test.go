package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

func setupTxTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}
	return db
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	t.Parallel()
	db := setupTxTestDB(t)
	defer store.CloseDB(db)

	err := store.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, principal_id, sequence)
			VALUES ('evt-tx-1', 'agg-1', 'Document', 'DocumentCreated', '{}', 'user-1', 1)`)
		return err
	})
	if err != nil {
		t.Fatalf("WithTx error: %v", err)
	}

	// Row should be visible outside the transaction
	var id string
	err = db.QueryRow("SELECT id FROM events WHERE id = 'evt-tx-1'").Scan(&id)
	if err != nil {
		t.Fatalf("committed row not found: %v", err)
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	t.Parallel()
	db := setupTxTestDB(t)
	defer store.CloseDB(db)

	errBoom := errors.New("boom")
	err := store.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		_, _ = tx.Exec(`INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, principal_id, sequence)
			VALUES ('evt-tx-2', 'agg-1', 'Document', 'DocumentCreated', '{}', 'user-1', 1)`)
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got: %v", err)
	}

	// Row should NOT be visible
	var id string
	err = db.QueryRow("SELECT id FROM events WHERE id = 'evt-tx-2'").Scan(&id)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after rollback, got: %v (id=%s)", err, id)
	}
}

func TestWithTx_RollsBackOnPanic(t *testing.T) {
	t.Parallel()
	db := setupTxTestDB(t)
	defer store.CloseDB(db)

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic to propagate")
			}
		}()
		_ = store.WithTx(context.Background(), db, func(tx *sql.Tx) error {
			_, _ = tx.Exec(`INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, principal_id, sequence)
				VALUES ('evt-tx-3', 'agg-1', 'Document', 'DocumentCreated', '{}', 'user-1', 1)`)
			panic("kaboom")
		})
	}()

	// Row should NOT be visible after panic rollback
	var id string
	err := db.QueryRow("SELECT id FROM events WHERE id = 'evt-tx-3'").Scan(&id)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after panic rollback, got: %v (id=%s)", err, id)
	}
}
