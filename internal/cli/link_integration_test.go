package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
	"github.com/stretchr/testify/require"
)

// setupDocsForIntegration creates two documents with sections @a1 and @b1
// (each in its own doc) so that link propose can resolve refs.
// Since findSectionByRef scans all docs and returns first match, we use
// distinct refs: doc1 has only @a1, doc2 has @a1 too but we search unique.
// For clarity, doc A has "# Spec" (section @a1) and doc B has a different
// heading that also produces @a1.
func setupDocsForIntegration(t *testing.T, dbPath string) {
	t.Helper()
	// Doc with section @a1
	cmd1 := cli.NewRootCmd()
	cmd1.SetOut(new(bytes.Buffer))
	cmd1.SetErr(new(bytes.Buffer))
	cmd1.SetArgs([]string{"--db", dbPath, "doc", "create", "Spec Doc", "--content", "# API Spec"})
	require.NoError(t, cmd1.Execute())

	// Doc with section @a1 (different doc, same ref — but CLI picks first match)
	cmd2 := cli.NewRootCmd()
	cmd2.SetOut(new(bytes.Buffer))
	cmd2.SetErr(new(bytes.Buffer))
	cmd2.SetArgs([]string{"--db", dbPath, "doc", "create", "Impl Doc", "--content", "# Implementation"})
	require.NoError(t, cmd2.Execute())
}

func TestLinkPropose_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	cmd := cli.NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	// Both docs have @a1, CLI resolves to first match for each — the link stores section IDs
	cmd.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--implements", "@a1", "--rationale", "impl covers spec"})

	require.NoError(t, cmd.Execute())
	out := buf.String()
	require.Contains(t, out, "opened")
	require.Contains(t, out, "@a1")
	require.Contains(t, out, "implements")
}

func TestLinkProposeAndList_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	// Step 1: propose a link
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--implements", "@a1", "--rationale", "auth spec"})
	require.NoError(t, cmd1.Execute())

	// Step 2: list links -- should show the proposed link
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "link", "list"})
	require.NoError(t, cmd2.Execute())

	out := buf2.String()
	require.Contains(t, out, "implements")
	require.Contains(t, out, "pending")
	require.NotContains(t, out, "(no links)")
}

func TestLinkProposeApproveList_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	// Step 1: propose
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--tests", "@a1", "--rationale", "unit test"})
	require.NoError(t, cmd1.Execute())

	// Extract full link ID from propose output: "link <ULID> opened: ..."
	linkID := extractLinkIDFromPropose(t, buf1.String())

	// Step 2: approve
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "link", "approve", linkID})
	require.NoError(t, cmd2.Execute())
	require.Contains(t, buf2.String(), "approved")

	// Step 3: verify state changed to aligned
	cmd3 := cli.NewRootCmd()
	buf3 := &bytes.Buffer{}
	cmd3.SetOut(buf3)
	cmd3.SetArgs([]string{"--db", dbPath, "link", "list"})
	require.NoError(t, cmd3.Execute())
	require.Contains(t, buf3.String(), "aligned")
}

func TestLinkComment_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	// Propose
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--evidences", "@a1", "--rationale", "data"})
	require.NoError(t, cmd1.Execute())

	linkID := extractLinkIDFromPropose(t, buf1.String())

	// Comment
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "link", "comment", linkID, "Looks good"})
	require.NoError(t, cmd2.Execute())
	require.Contains(t, buf2.String(), "commented on")
}

func TestLinkReaffirmAll_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	// Propose and approve a link
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", "--implements", "@a1", "--rationale", "spec"})
	require.NoError(t, cmd1.Execute())

	linkID := extractLinkIDFromPropose(t, buf1.String())

	// Approve (pending -> aligned)
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "link", "approve", linkID})
	require.NoError(t, cmd2.Execute())

	// Reaffirm --all (no stale links, should succeed with 0 count)
	cmd3 := cli.NewRootCmd()
	buf3 := &bytes.Buffer{}
	cmd3.SetOut(buf3)
	cmd3.SetArgs([]string{"--db", dbPath, "link", "reaffirm", "--all"})
	require.NoError(t, cmd3.Execute())
	require.Contains(t, buf3.String(), "reaffirmed 0 stale links")
}

func TestLinkMultiplePropose_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupDocsForIntegration(t, dbPath)

	// Propose two links (both use @a1 since that's the ref in both docs)
	for _, rel := range []struct {
		flag  string
		value string
	}{
		{"--implements", "@a1"},
		{"--agrees-with", "@a1"},
	} {
		cmd := cli.NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--db", dbPath, "link", "propose", "@a1", rel.flag, rel.value, "--rationale", "test"})
		require.NoError(t, cmd.Execute())
	}

	// List should show both
	cmd := cli.NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", dbPath, "link", "list"})
	require.NoError(t, cmd.Execute())

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Equal(t, 2, len(lines), "expected 2 links, got output:\n%s", out)
}

// extractLinkIDFromPropose extracts the full ULID from propose output:
// "link <ULID> opened: @a1 --implements-> @a1"
func extractLinkIDFromPropose(t *testing.T, output string) string {
	t.Helper()
	prefix := "link "
	idx := strings.Index(output, prefix)
	if idx < 0 {
		t.Fatalf("could not find %q in output: %s", prefix, output)
	}
	rest := output[idx+len(prefix):]
	spaceIdx := strings.Index(rest, " ")
	if spaceIdx < 0 {
		t.Fatalf("could not find space after link ID in: %s", output)
	}
	return rest[:spaceIdx]
}
