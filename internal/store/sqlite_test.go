package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

func TestOpenDB_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB(:memory:) error: %v", err)
	}
	if db == nil {
		t.Fatal("OpenDB(:memory:) returned nil *sql.DB")
	}
	defer store.CloseDB(db)
}

func TestOpenDB_WALMode(t *testing.T) {
	t.Parallel()
	// WAL pragma returns "memory" for :memory: DBs, so use a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := store.OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	var mode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("PRAGMA journal_mode query error: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %q", mode)
	}
}

func TestOpenDB_ForeignKeysEnabled(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		t.Fatalf("PRAGMA foreign_keys query error: %v", err)
	}
	if fk != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", fk)
	}
}

func TestCloseDB(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	if err := store.CloseDB(db); err != nil {
		t.Fatalf("CloseDB error: %v", err)
	}
}

func TestOpenDB_TempFile_Cleanup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "cleanup.db")

	db, err := store.OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	store.CloseDB(db)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}
