package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lagz0ne/remmd/internal/app"
)

func TestNew_InMemory(t *testing.T) {
	t.Parallel()
	a, err := app.New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:): %v", err)
	}
	defer a.Close()
	if a.Docs == nil || a.Links == nil || a.Subscriptions == nil {
		t.Fatal("repos should not be nil")
	}
}

func TestNew_FileDB(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	a, err := app.New(dbPath)
	if err != nil {
		t.Fatalf("New(file): %v", err)
	}
	defer a.Close()
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("db file should exist: %v", err)
	}
}

func TestNew_CloseTwiceNoError(t *testing.T) {
	t.Parallel()
	a, err := app.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	a.Close()
	// second close should not panic
}
