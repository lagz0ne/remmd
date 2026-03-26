package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

func TestPlaybookCheck_ValidFile(t *testing.T) {
	t.Parallel()
	fixturePath := filepath.Join("..", "playbook", "testdata", "c3.playbook.yaml")
	if _, err := os.Stat(fixturePath); err != nil {
		t.Skipf("fixture not found: %s", fixturePath)
	}

	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"playbook", "check", fixturePath})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected success, got: %v\noutput: %s", err, buf.String())
	}
	output := buf.String()
	if !strings.Contains(output, "types") || !strings.Contains(output, "edges") {
		t.Fatalf("expected summary with types and edges, got: %s", output)
	}
	if !strings.Contains(output, "rules") {
		t.Fatalf("expected summary with rules, got: %s", output)
	}
}

func TestPlaybookCheck_MissingFile(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"playbook", "check", "nonexistent.yaml"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPlaybookCheck_InvalidYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad.yaml")
	os.WriteFile(bad, []byte("not: [valid: yaml: {{"), 0644)

	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"playbook", "check", bad})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestPlaybookCheck_BrokenEdgeRef(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	f := filepath.Join(tmp, "broken-edge.yaml")
	// Edge references "ghost" type which doesn't exist
	os.WriteFile(f, []byte(`
component:
  goal: string!
broken-edge: "ghost -> component [1..*]"
`), 0644)

	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"playbook", "check", f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for broken edge reference")
	}
	output := buf.String()
	if !strings.Contains(output, "ghost") {
		t.Fatalf("expected error mentioning 'ghost', got: %s", output)
	}
}

func TestPlaybookCheck_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"playbook", "check"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no file argument provided")
	}
}
