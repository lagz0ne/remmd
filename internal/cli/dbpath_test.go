package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverDBPathFrom_FindsLocal(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	dbDir := filepath.Join(tmp, ".remmd")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatal(err)
	}
	dbFile := filepath.Join(dbDir, "remmd.db")
	if err := os.WriteFile(dbFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	got := discoverDBPathFrom(tmp)
	if got != dbFile {
		t.Fatalf("expected %q, got %q", dbFile, got)
	}
}

func TestDiscoverDBPathFrom_WalksUp(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	dbDir := filepath.Join(tmp, ".remmd")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatal(err)
	}
	dbFile := filepath.Join(dbDir, "remmd.db")
	if err := os.WriteFile(dbFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a nested subdirectory to start from
	nested := filepath.Join(tmp, "sub", "deep")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	got := discoverDBPathFrom(nested)
	if got != dbFile {
		t.Fatalf("expected %q, got %q", dbFile, got)
	}
}

func TestDiscoverDBPathFrom_FallsToGlobal(t *testing.T) {
	t.Parallel()

	// Use a temp dir with no .remmd/ anywhere in the hierarchy
	tmp := t.TempDir()

	got := discoverDBPathFrom(tmp)
	expected := globalDBPath()
	if got != expected {
		t.Fatalf("expected global path %q, got %q", expected, got)
	}
}

func TestDiscoverDBPathFrom_PrefersClosest(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// Create .remmd/remmd.db at root level
	rootDB := filepath.Join(tmp, ".remmd")
	if err := os.MkdirAll(rootDB, 0755); err != nil {
		t.Fatal(err)
	}
	rootDBFile := filepath.Join(rootDB, "remmd.db")
	if err := os.WriteFile(rootDBFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Create .remmd/remmd.db at nested level
	nested := filepath.Join(tmp, "project")
	nestedDB := filepath.Join(nested, ".remmd")
	if err := os.MkdirAll(nestedDB, 0755); err != nil {
		t.Fatal(err)
	}
	nestedDBFile := filepath.Join(nestedDB, "remmd.db")
	if err := os.WriteFile(nestedDBFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Start from nested — should find the closest one
	got := discoverDBPathFrom(nested)
	if got != nestedDBFile {
		t.Fatalf("expected closest %q, got %q", nestedDBFile, got)
	}
}
