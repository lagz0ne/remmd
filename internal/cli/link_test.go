package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

// setupDocsForLinkTest creates two documents with sections @a1 and @b1
// so that link propose can resolve refs to section IDs.
func setupDocsForLinkTest(t *testing.T, dbFlag string) {
	t.Helper()
	// Create doc with @a1
	cmd1 := cli.NewRootCmd()
	cmd1.SetOut(new(bytes.Buffer))
	cmd1.SetErr(new(bytes.Buffer))
	cmd1.SetArgs([]string{"--db", dbFlag, "doc", "create", "Doc A", "--content", "# Section A1"})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("setup doc A: %v", err)
	}

	// Create doc with @b1 (will be @a1 in its own doc, so use different content structure)
	cmd2 := cli.NewRootCmd()
	cmd2.SetOut(new(bytes.Buffer))
	cmd2.SetErr(new(bytes.Buffer))
	cmd2.SetArgs([]string{"--db", dbFlag, "doc", "create", "Doc B", "--content", "# Section B1"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("setup doc B: %v", err)
	}
}

func TestLinkPropose(t *testing.T) {
	t.Parallel()
	dbPath := t.TempDir() + "/test.db"
	setupDocsForLinkTest(t, dbPath)

	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--implements", "@a1", "--rationale", "impl matches spec"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("link propose error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"@a1", "implements", "impl matches spec"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestLinkPropose_AgreesWith(t *testing.T) {
	t.Parallel()
	dbPath := t.TempDir() + "/test.db"
	setupDocsForLinkTest(t, dbPath)

	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--agrees-with", "@a1", "--rationale", "same API"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(buf.String(), "agrees_with") {
		t.Errorf("output missing agrees_with: %s", buf.String())
	}
}

func TestLinkPropose_RequiresRelationship(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--db", ":memory:", "link", "propose", "@a1", "--rationale", "reason"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when no relationship type provided")
	}
}

func TestLinkList_Empty(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "link", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(buf.String(), "(no links)") {
		t.Errorf("expected empty list message, got: %s", buf.String())
	}
}

func TestLinkList_StaleFilter_Empty(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "link", "list", "--stale"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	// Empty DB should show "(no links)"
	if !strings.Contains(buf.String(), "(no links)") {
		t.Errorf("output: %s", buf.String())
	}
}

func TestLinkList_MineFilter_Empty(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "link", "list", "--mine"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(buf.String(), "(no links)") {
		t.Errorf("output: %s", buf.String())
	}
}

func TestLinkReaffirm_RequiresArg(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--db", ":memory:", "link", "reaffirm"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when no link-id provided")
	}
}

func TestLinkReaffirm_All_Empty(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "link", "reaffirm", "--all"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(buf.String(), "reaffirmed") {
		t.Errorf("output: %s", buf.String())
	}
}
