package cli_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

// execCmd is a helper that creates a root cmd, sets args with --db :memory:,
// executes, and returns stdout output + error.
func execCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"--db", ":memory:"}, args...))
	err := cmd.Execute()
	return buf.String(), err
}

// execCmdChain runs multiple command sequences on the same in-memory DB.
// Since each NewRootCmd creates a separate :memory: DB, this helper uses
// a temp file to share state between commands.
func execCmdChain(t *testing.T, argSets ...[]string) ([]string, error) {
	t.Helper()
	tmpDB := t.TempDir() + "/test.db"
	var outputs []string
	for _, args := range argSets {
		cmd := cli.NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(append([]string{"--db", tmpDB}, args...))
		if err := cmd.Execute(); err != nil {
			return outputs, err
		}
		outputs = append(outputs, buf.String())
	}
	return outputs, nil
}

func TestDocCreate_Integration(t *testing.T) {
	t.Parallel()
	out, err := execCmd(t, "doc", "create", "Test Doc", "--content", "# Hello\n## World")
	if err != nil {
		t.Fatalf("doc create error: %v", err)
	}
	if !strings.Contains(out, "Test Doc") {
		t.Errorf("output missing title: %s", out)
	}
	if !strings.Contains(out, "created") {
		t.Errorf("output missing 'created': %s", out)
	}
	if !strings.Contains(out, "2 sections") {
		t.Errorf("expected 2 sections: %s", out)
	}
	// Should show section refs in tree
	if !strings.Contains(out, "@") {
		t.Errorf("output missing @refs in tree: %s", out)
	}
}

func TestDocCreate_NoContent(t *testing.T) {
	t.Parallel()
	out, err := execCmd(t, "doc", "create", "Empty Doc")
	if err != nil {
		t.Fatalf("doc create error: %v", err)
	}
	if !strings.Contains(out, "0 sections") {
		t.Errorf("expected 0 sections: %s", out)
	}
}

func TestDocList_AfterCreate(t *testing.T) {
	t.Parallel()
	outputs, err := execCmdChain(t,
		[]string{"doc", "create", "First Doc"},
		[]string{"doc", "create", "Second Doc"},
		[]string{"doc", "list"},
	)
	if err != nil {
		t.Fatalf("chain error: %v", err)
	}
	listOut := outputs[2]
	if !strings.Contains(listOut, "First Doc") {
		t.Errorf("list missing 'First Doc': %s", listOut)
	}
	if !strings.Contains(listOut, "Second Doc") {
		t.Errorf("list missing 'Second Doc': %s", listOut)
	}
}

func TestDocList_Empty(t *testing.T) {
	t.Parallel()
	out, err := execCmd(t, "doc", "list")
	if err != nil {
		t.Fatalf("doc list error: %v", err)
	}
	if !strings.Contains(out, "(no documents)") {
		t.Errorf("expected empty list message: %s", out)
	}
}

func TestShow_Section(t *testing.T) {
	t.Parallel()
	outputs, err := execCmdChain(t,
		[]string{"doc", "create", "Show Test", "--content", "# Overview\n## Details"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	// Extract a ref from the create output
	ref := extractFirstRef(t, outputs[0])

	// Show that section
	tmpDB := t.TempDir() + "/test.db"
	// Re-create with same DB
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Show Test", "--content", "# Overview\n## Details"},
		[]string{"show", ref},
	)
	if err != nil {
		t.Fatalf("show error: %v", err)
	}
	showOut := outputs2[1]
	if !strings.Contains(showOut, ref) {
		t.Errorf("show output missing ref %s: %s", ref, showOut)
	}
	if !strings.Contains(showOut, "heading") {
		t.Errorf("show output missing section type: %s", showOut)
	}
}

func TestShow_Document(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Create a doc and capture its ID
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc Show", "--content", "# Heading"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	// Extract doc ID from "doc <ID> "Doc Show" created"
	docID := extractDocID(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"show", docID},
	)
	if err != nil {
		t.Fatalf("show error: %v", err)
	}
	if !strings.Contains(outputs2[0], "Doc Show") {
		t.Errorf("show output missing title: %s", outputs2[0])
	}
}

func TestEdit_Content(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Edit Test", "--content", "# Section One"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	ref := extractFirstRef(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"edit", ref, "--content", "Updated content"},
	)
	if err != nil {
		t.Fatalf("edit error: %v", err)
	}
	if !strings.Contains(outputs2[0], "content updated") {
		t.Errorf("edit output missing 'content updated': %s", outputs2[0])
	}
}

func TestEdit_Tag(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Tag Test", "--content", "# Section"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	ref := extractFirstRef(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"edit", ref, "--tag", "api"},
	)
	if err != nil {
		t.Fatalf("edit error: %v", err)
	}
	if !strings.Contains(outputs2[0], "tagged") {
		t.Errorf("edit output missing 'tagged': %s", outputs2[0])
	}
	if !strings.Contains(outputs2[0], "api") {
		t.Errorf("edit output missing tag name: %s", outputs2[0])
	}
}

func TestDelete_Section(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Delete Test", "--content", "# ToDelete"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	ref := extractFirstRef(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"delete", ref, "--reason", "Obsolete"},
	)
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if !strings.Contains(outputs2[0], "deleted") {
		t.Errorf("delete output missing 'deleted': %s", outputs2[0])
	}
	if !strings.Contains(outputs2[0], "Obsolete") {
		t.Errorf("delete output missing reason: %s", outputs2[0])
	}
}

func TestDelete_WithReplacement(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Replace Test", "--content", "# First\n# Second"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	// Get first ref to delete
	ref := extractFirstRef(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"delete", ref, "--reason", "Merged", "--replacement", "@b2"},
	)
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if !strings.Contains(outputs2[0], "deleted") {
		t.Errorf("delete output missing 'deleted': %s", outputs2[0])
	}
	if !strings.Contains(outputs2[0], "@b2") {
		t.Errorf("delete output missing replacement: %s", outputs2[0])
	}
}

func TestCreateAndList_Integration(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A", "--content", "# One\n## Two\n### Three"},
		[]string{"doc", "list"},
	)
	if err != nil {
		t.Fatalf("chain error: %v", err)
	}

	createOut := outputs[0]
	if !strings.Contains(createOut, "3 sections") {
		t.Errorf("create should show 3 sections: %s", createOut)
	}

	listOut := outputs[1]
	if !strings.Contains(listOut, "Doc A") {
		t.Errorf("list missing doc: %s", listOut)
	}
	if !strings.Contains(listOut, "3 sections") {
		t.Errorf("list should show section count: %s", listOut)
	}
}

// execCmdChain2 is like execCmdChain but accepts the DB path explicitly.
func execCmdChain2(t *testing.T, dbPath string, argSets ...[]string) ([]string, error) {
	t.Helper()
	var outputs []string
	for _, args := range argSets {
		cmd := cli.NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(append([]string{"--db", dbPath}, args...))
		if err := cmd.Execute(); err != nil {
			return outputs, err
		}
		outputs = append(outputs, buf.String())
	}
	return outputs, nil
}

// extractFirstRef finds the first @ref (like @a1) in output text.
func extractFirstRef(t *testing.T, output string) string {
	t.Helper()
	re := regexp.MustCompile(`@[a-z]+\d+`)
	match := re.FindString(output)
	if match == "" {
		t.Fatalf("no @ref found in output: %s", output)
	}
	return match
}

// extractDocID extracts the document ID from create output like:
// doc <ULID> "title" created (N sections)
func extractDocID(t *testing.T, output string) string {
	t.Helper()
	// ULID is 26 uppercase alphanumeric chars
	re := regexp.MustCompile(`doc ([0-9A-Z]{26})`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 {
		t.Fatalf("no doc ID found in output: %s", output)
	}
	return match[1]
}
