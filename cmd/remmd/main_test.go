package main_test

import (
	"os"
	"os/exec"
	"testing"
)

func TestBinaryBuilds(t *testing.T) {
	t.Parallel()
	tmpBin := t.TempDir() + "/remmd-test"
	cmd := exec.Command("go", "build", "-o", tmpBin, "./")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary build failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(tmpBin); err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
}

func TestBinaryHelp_ExitsZero(t *testing.T) {
	t.Parallel()
	tmpBin := t.TempDir() + "/remmd-test"
	build := exec.Command("go", "build", "-o", tmpBin, "./")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cmd := exec.Command(tmpBin, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--help exited non-zero: %v\n%s", err, out)
	}
}
