package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

func TestDocCreateCmd(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "doc", "create", "My Doc"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("doc create error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "My Doc") {
		t.Errorf("output missing title: %s", out)
	}
	if !strings.Contains(out, "created") {
		t.Errorf("output missing 'created': %s", out)
	}
}

func TestDocCreateCmd_WithContent(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "doc", "create", "My Doc", "--content", "# Hello"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("doc create error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 sections") {
		t.Errorf("output missing section count: %s", out)
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("output missing section title: %s", out)
	}
}

func TestDocListCmd(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "doc", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("doc list error: %v", err)
	}
	if !strings.Contains(buf.String(), "(no documents)") {
		t.Errorf("expected empty list message: %s", buf.String())
	}
}

func TestDocCreateCmd_RequiresTitle(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "doc", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no title provided")
	}
}

func TestShowCmd_NotFound(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "show", "@a1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestEditCmd_NotFound(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "edit", "@a1", "--content", "Updated"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDeleteCmd_NotFound(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "delete", "@a1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}
