package cli

import (
	"os"
	"path/filepath"
)

// discoverDBPath walks up from the current working directory looking for
// .remmd/remmd.db. Returns the first match, or falls back to ~/.remmd/remmd.db.
func discoverDBPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return globalDBPath()
	}
	return discoverDBPathFrom(dir)
}

// discoverDBPathFrom walks up from startDir looking for .remmd/remmd.db.
// Returns the first match, or falls back to ~/.remmd/remmd.db.
func discoverDBPathFrom(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".remmd", "remmd.db")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}
	return globalDBPath()
}

// globalDBPath returns the global fallback database path at ~/.remmd/remmd.db,
// creating the directory if it doesn't exist.
func globalDBPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".remmd")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "remmd.db")
}
