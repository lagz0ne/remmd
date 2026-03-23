package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

func TestTagSubscribe(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "tag", "subscribe", "@a1", "--tag", "payment"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "@a1") || !strings.Contains(out, "payment") {
		t.Errorf("output: %s", out)
	}
}

func TestTagSubscribe_RequiresTag(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--db", ":memory:", "tag", "subscribe", "@a1"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --tag not provided")
	}
}

func TestTagUnsubscribe_NotFound(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--db", ":memory:", "tag", "unsubscribe", "nonexistent-id"})

	// Should error because no subscription exists with that ID
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for non-existent subscription")
	}
}

func TestTagList_Empty(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "tag", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(buf.String(), "(no active subscriptions)") {
		t.Errorf("expected empty list, got: %s", buf.String())
	}
}

func TestTagDismiss_NotFound(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--db", ":memory:", "tag", "dismiss", "nonexistent-fire"})

	// Should error because no fire exists with that ID
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for non-existent fire")
	}
}
