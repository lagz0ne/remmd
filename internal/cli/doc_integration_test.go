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

func TestDocCreateExternal(t *testing.T) {
	t.Parallel()
	out, err := execCmd(t, "doc", "create", "Notion Page",
		"--external",
		"--system", "notion",
		"--external-id", "page-abc",
		"--hash", "sha256abc",
	)
	if err != nil {
		t.Fatalf("doc create external error: %v", err)
	}
	if !strings.Contains(out, "external section") {
		t.Errorf("output missing 'external section': %s", out)
	}
	if !strings.Contains(out, "@ext:notion/page-abc") {
		t.Errorf("output missing external ref '@ext:notion/page-abc': %s", out)
	}
}

func TestDocCreateExternalWithMetadata(t *testing.T) {
	t.Parallel()
	out, err := execCmd(t, "doc", "create", "With Meta",
		"--external",
		"--system", "figma",
		"--external-id", "frame-1",
		"--hash", "hash1",
		"--metadata", `{"file_key":"xyz"}`,
	)
	if err != nil {
		t.Fatalf("doc create external with metadata error: %v", err)
	}
	if !strings.Contains(out, "external section") {
		t.Errorf("output missing 'external section': %s", out)
	}
}

func TestDocCreateExternalMissingFlags(t *testing.T) {
	t.Parallel()
	// --external without required --system, --external-id, --hash should error
	_, err := execCmd(t, "doc", "create", "Bad", "--external")
	if err == nil {
		t.Fatal("expected error when --external is used without --system, --external-id, --hash")
	}
}

func TestShowExternalSection(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: create an external doc
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Ext Show",
			"--external",
			"--system", "test",
			"--external-id", "t1",
			"--hash", "abc123",
		},
	)
	if err != nil {
		t.Fatalf("create external doc error: %v", err)
	}
	_ = outputs

	// Step 2: show the external section by its ref
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"show", "@ext:test/t1"},
	)
	if err != nil {
		t.Fatalf("show external section error: %v", err)
	}
	showOut := outputs2[0]
	if !strings.Contains(showOut, "external") {
		t.Errorf("show output missing 'external': %s", showOut)
	}
	if !strings.Contains(showOut, "abc123") {
		t.Errorf("show output missing hash 'abc123': %s", showOut)
	}
	if !strings.Contains(showOut, "test") {
		t.Errorf("show output missing system name 'test': %s", showOut)
	}
}

func TestEditExternalHash(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: create external doc with hash "old"
	_, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Hash Edit",
			"--external",
			"--system", "test",
			"--external-id", "t1",
			"--hash", "old",
		},
	)
	if err != nil {
		t.Fatalf("create external doc error: %v", err)
	}

	// Step 2: edit the external section hash
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"edit", "@ext:test/t1", "--hash", "new"},
	)
	if err != nil {
		t.Fatalf("edit external hash error: %v", err)
	}
	editOut := outputs2[0]
	if !strings.Contains(editOut, "hash") {
		t.Errorf("edit output missing 'hash': %s", editOut)
	}
	if !strings.Contains(editOut, "updated") {
		t.Errorf("edit output missing 'updated': %s", editOut)
	}

	// Step 3: verify via show
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"show", "@ext:test/t1"},
	)
	if err != nil {
		t.Fatalf("show after edit error: %v", err)
	}
	showOut := outputs3[0]
	if !strings.Contains(showOut, "new") {
		t.Errorf("show output should contain new hash 'new': %s", showOut)
	}
}

func TestEditExternalRejectsContent(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: create external doc
	_, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "No Edit",
			"--external",
			"--system", "test",
			"--external-id", "t1",
			"--hash", "somehash",
		},
	)
	if err != nil {
		t.Fatalf("create external doc error: %v", err)
	}

	// Step 2: try to edit content — should fail for external sections
	_, err = execCmdChain2(t, tmpDB,
		[]string{"edit", "@ext:test/t1", "--content", "body text"},
	)
	if err == nil {
		t.Fatal("expected error when editing content of an external section")
	}
	if !strings.Contains(err.Error(), "external") {
		t.Errorf("error should mention 'external': %v", err)
	}
}

func TestLinkProposeWithExternalRef(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: create a native doc (produces @a1)
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Native Doc", "--content", "# API Spec"},
	)
	if err != nil {
		t.Fatalf("create native doc error: %v", err)
	}
	nativeRef := extractFirstRef(t, outputs[0])

	// Step 2: create an external doc (produces @ext:test/t1)
	_, err = execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "External Doc",
			"--external",
			"--system", "test",
			"--external-id", "t1",
			"--hash", "exthash",
		},
	)
	if err != nil {
		t.Fatalf("create external doc error: %v", err)
	}

	// Step 3: propose a link between native and external sections
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"link", "propose", nativeRef,
			"--implements", "@ext:test/t1",
			"--rationale", "impl covers external spec",
		},
	)
	if err != nil {
		t.Fatalf("link propose error: %v", err)
	}
	proposeOut := outputs3[0]
	if !strings.Contains(proposeOut, "opened") {
		t.Errorf("propose output missing 'opened': %s", proposeOut)
	}
	if !strings.Contains(proposeOut, "@ext:test/t1") {
		t.Errorf("propose output missing external ref: %s", proposeOut)
	}
}

func TestEdit_TransitionsLinksToStale(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: Create two docs with one section each so refs are distinct.
	// Doc A → @a1, Doc B → @b2
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A", "--content", "# Section A"},
		[]string{"doc", "create", "Doc B", "--content", "# Section B"},
	)
	if err != nil {
		t.Fatalf("create docs: %v", err)
	}
	refA := extractFirstRef(t, outputs[0])
	refB := extractFirstRef(t, outputs[1])

	// Step 2: Propose a link between the two sections
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"link", "propose", refA, "--agrees-with", refB, "--rationale", "same API"},
	)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	linkID := extractLinkIDFromPropose(t, outputs2[0])

	// Step 3: Approve → aligned
	_, err = execCmdChain2(t, tmpDB,
		[]string{"link", "approve", linkID},
	)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Verify aligned
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"link", "list"},
	)
	if err != nil {
		t.Fatalf("list after approve: %v", err)
	}
	if !strings.Contains(outputs3[0], "aligned") {
		t.Fatalf("link should be aligned after approve, got: %s", outputs3[0])
	}

	// Step 4: Edit section A → should trigger stale transition
	outputs4, err := execCmdChain2(t, tmpDB,
		[]string{"edit", refA, "--content", "Changed content"},
	)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	editOut := outputs4[0]
	if !strings.Contains(editOut, "content updated") {
		t.Errorf("edit output missing 'content updated': %s", editOut)
	}
	if !strings.Contains(editOut, "1 link(s) marked stale") {
		t.Errorf("edit output should report stale links: %s", editOut)
	}

	// Step 5: Verify link is now stale
	outputs5, err := execCmdChain2(t, tmpDB,
		[]string{"link", "list"},
	)
	if err != nil {
		t.Fatalf("list after edit: %v", err)
	}
	if !strings.Contains(outputs5[0], "stale") {
		t.Errorf("link should be stale after edit, got: %s", outputs5[0])
	}
}

func TestEdit_NoStaleTransition_WhenNotAligned(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: Create two docs
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A", "--content", "# Section A"},
		[]string{"doc", "create", "Doc B", "--content", "# Section B"},
	)
	if err != nil {
		t.Fatalf("create docs: %v", err)
	}
	refA := extractFirstRef(t, outputs[0])
	refB := extractFirstRef(t, outputs[1])

	// Step 2: Propose a link but do NOT approve — stays pending
	_, err = execCmdChain2(t, tmpDB,
		[]string{"link", "propose", refA, "--implements", refB, "--rationale", "impl"},
	)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}

	// Step 3: Edit section A — link should stay pending (not transition to stale)
	_, err = execCmdChain2(t, tmpDB,
		[]string{"edit", refA, "--content", "New content"},
	)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	// Step 4: Verify link is still pending
	outputs4, err := execCmdChain2(t, tmpDB,
		[]string{"link", "list"},
	)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(outputs4[0], "pending") {
		t.Errorf("link should remain pending, got: %s", outputs4[0])
	}
	if strings.Contains(outputs4[0], "stale") {
		t.Errorf("link should NOT be stale (was pending), got: %s", outputs4[0])
	}
}

func TestEdit_ExternalHashUpdate_TransitionsLinksToStale(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: Create a native doc and an external doc
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Native", "--content", "# Spec"},
	)
	if err != nil {
		t.Fatalf("create native: %v", err)
	}
	nativeRef := extractFirstRef(t, outputs[0])

	_, err = execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "External",
			"--external", "--system", "notion", "--external-id", "page1", "--hash", "hash1"},
	)
	if err != nil {
		t.Fatalf("create external: %v", err)
	}

	// Step 2: Propose and approve link
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"link", "propose", nativeRef, "--implements", "@ext:notion/page1", "--rationale", "covers"},
	)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	linkID := extractLinkIDFromPropose(t, outputs2[0])

	_, err = execCmdChain2(t, tmpDB,
		[]string{"link", "approve", linkID},
	)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Step 3: Update external hash → should transition to stale
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"edit", "@ext:notion/page1", "--hash", "hash2"},
	)
	if err != nil {
		t.Fatalf("edit external hash: %v", err)
	}
	if !strings.Contains(outputs3[0], "1 link(s) marked stale") {
		t.Errorf("external hash edit should report stale links: %s", outputs3[0])
	}

	// Step 4: Verify stale
	outputs4, err := execCmdChain2(t, tmpDB,
		[]string{"link", "list"},
	)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(outputs4[0], "stale") {
		t.Errorf("link should be stale after external hash update, got: %s", outputs4[0])
	}
}

func TestImpact_ShowsImpactedLinks(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Create two docs with distinct refs
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A", "--content", "# Section A"},
		[]string{"doc", "create", "Doc B", "--content", "# Section B"},
	)
	if err != nil {
		t.Fatalf("create docs: %v", err)
	}
	refA := extractFirstRef(t, outputs[0])
	refB := extractFirstRef(t, outputs[1])

	// Propose and approve a link
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"link", "propose", refA, "--agrees-with", refB, "--rationale", "aligned"},
	)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	linkID := extractLinkIDFromPropose(t, outputs2[0])

	_, err = execCmdChain2(t, tmpDB,
		[]string{"link", "approve", linkID},
	)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Run impact on refA — should show the link
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"impact", refA},
	)
	if err != nil {
		t.Fatalf("impact: %v", err)
	}
	impactOut := outputs3[0]
	if !strings.Contains(impactOut, "1 link(s) impacted") {
		t.Errorf("impact should show 1 impacted link, got: %s", impactOut)
	}
	if !strings.Contains(impactOut, "agrees_with") {
		t.Errorf("impact should show relationship type, got: %s", impactOut)
	}
}

func TestImpact_NoLinks(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Create a doc with no links
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Lonely Doc", "--content", "# Alone"},
	)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	ref := extractFirstRef(t, outputs[0])

	// Run impact — should show no impact
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"impact", ref},
	)
	if err != nil {
		t.Fatalf("impact: %v", err)
	}
	if !strings.Contains(outputs2[0], "no impact") {
		t.Errorf("impact should show no impact, got: %s", outputs2[0])
	}
}

func TestDocDelete(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Delete Me", "--content", "# Section One\n# Section Two"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	docID := extractDocID(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "delete", docID},
	)
	if err != nil {
		t.Fatalf("doc delete error: %v", err)
	}
	deleteOut := outputs2[0]
	if !strings.Contains(deleteOut, "deleted") {
		t.Errorf("output missing 'deleted': %s", deleteOut)
	}
	if !strings.Contains(deleteOut, "Delete Me") {
		t.Errorf("output missing title: %s", deleteOut)
	}
	if !strings.Contains(deleteOut, "2 sections removed") {
		t.Errorf("output missing section count: %s", deleteOut)
	}

	// Verify doc is gone from list
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "list"},
	)
	if err != nil {
		t.Fatalf("doc list error: %v", err)
	}
	if strings.Contains(outputs3[0], "Delete Me") {
		t.Errorf("deleted doc should not appear in list: %s", outputs3[0])
	}
}

func TestDocDelete_NotFound(t *testing.T) {
	t.Parallel()
	_, err := execCmd(t, "doc", "delete", "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for non-existent doc")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDocDelete_Multiple(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A"},
		[]string{"doc", "create", "Doc B"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	idA := extractDocID(t, outputs[0])
	idB := extractDocID(t, outputs[1])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "delete", idA, idB},
	)
	if err != nil {
		t.Fatalf("doc delete error: %v", err)
	}
	deleteOut := outputs2[0]
	if !strings.Contains(deleteOut, "Doc A") {
		t.Errorf("output missing 'Doc A': %s", deleteOut)
	}
	if !strings.Contains(deleteOut, "Doc B") {
		t.Errorf("output missing 'Doc B': %s", deleteOut)
	}
}

func TestDocArchive(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Archive Me"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	docID := extractDocID(t, outputs[0])

	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "archive", docID},
	)
	if err != nil {
		t.Fatalf("doc archive error: %v", err)
	}
	archiveOut := outputs2[0]
	if !strings.Contains(archiveOut, "archived") {
		t.Errorf("output missing 'archived': %s", archiveOut)
	}
	if !strings.Contains(archiveOut, "Archive Me") {
		t.Errorf("output missing title: %s", archiveOut)
	}
}

func TestDocArchive_NotFound(t *testing.T) {
	t.Parallel()
	_, err := execCmd(t, "doc", "archive", "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for non-existent doc")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDocList_HidesArchived(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Active Doc"},
		[]string{"doc", "create", "Archived Doc"},
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	archivedID := extractDocID(t, outputs[1])

	// Archive the second doc
	_, err = execCmdChain2(t, tmpDB,
		[]string{"doc", "archive", archivedID},
	)
	if err != nil {
		t.Fatalf("archive error: %v", err)
	}

	// Default list should hide archived
	outputs3, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "list"},
	)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	listOut := outputs3[0]
	if !strings.Contains(listOut, "Active Doc") {
		t.Errorf("list should show active doc: %s", listOut)
	}
	if strings.Contains(listOut, "Archived Doc") {
		t.Errorf("list should hide archived doc by default: %s", listOut)
	}

	// --all flag should show archived
	outputs4, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "list", "--all"},
	)
	if err != nil {
		t.Fatalf("list --all error: %v", err)
	}
	allOut := outputs4[0]
	if !strings.Contains(allOut, "Active Doc") {
		t.Errorf("list --all should show active doc: %s", allOut)
	}
	if !strings.Contains(allOut, "Archived Doc") {
		t.Errorf("list --all should show archived doc: %s", allOut)
	}
	if !strings.Contains(allOut, "[archived]") {
		t.Errorf("list --all should mark archived docs: %s", allOut)
	}
}

func TestDocCreate_GlobalRefUniqueness(t *testing.T) {
	t.Parallel()
	tmpDB := t.TempDir() + "/test.db"

	// Step 1: Create doc A with 2 sections → should get @a1, @b2
	// Step 2: Create doc B with 1 section  → should get @c3 (NOT @a1)
	// Step 3: Verify show @c3 resolves to doc B's section
	outputs, err := execCmdChain2(t, tmpDB,
		[]string{"doc", "create", "Doc A", "--content", "# Sec One\n# Sec Two"},
		[]string{"doc", "create", "Doc B", "--content", "# Sec Three"},
	)
	if err != nil {
		t.Fatalf("chain error: %v", err)
	}

	createA := outputs[0]
	createB := outputs[1]

	// Doc A should have @a1 and @b2
	if !strings.Contains(createA, "@a1") {
		t.Errorf("doc A missing @a1: %s", createA)
	}
	if !strings.Contains(createA, "@b2") {
		t.Errorf("doc A missing @b2: %s", createA)
	}

	// Doc B should have @c3 (global counter continues from doc A)
	if !strings.Contains(createB, "@c3") {
		t.Errorf("doc B should have @c3 (global ref), got: %s", createB)
	}

	// @c3 should NOT appear in doc A
	if strings.Contains(createA, "@c3") {
		t.Errorf("doc A should NOT have @c3: %s", createA)
	}

	// @a1 should NOT appear in doc B
	if strings.Contains(createB, "@a1") {
		t.Errorf("doc B should NOT have @a1 (collision!): %s", createB)
	}

	// Verify show @c3 resolves to doc B's section
	outputs2, err := execCmdChain2(t, tmpDB,
		[]string{"show", "@c3"},
	)
	if err != nil {
		t.Fatalf("show @c3 error: %v", err)
	}
	showOut := outputs2[0]
	if !strings.Contains(showOut, "@c3") {
		t.Errorf("show output missing @c3: %s", showOut)
	}
	if !strings.Contains(showOut, "Sec Three") {
		t.Errorf("show @c3 should resolve to 'Sec Three': %s", showOut)
	}
}
