package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

func TestImpactCmd_Exists(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "impact [ref]" || strings.HasPrefix(sub.Use, "impact") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("root command missing 'impact' subcommand")
	}
}

func TestImpactCmd_NonExistentRef(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "impact", "@a1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestImpactCmd_HelpIncludesExamples(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"impact", "--help"})

	_ = cmd.Execute()
	help := buf.String()
	if !strings.Contains(help, "impact") {
		t.Errorf("help missing 'impact' keyword: %s", help)
	}
	if !strings.Contains(help, "Example") {
		t.Errorf("help missing examples section: %s", help)
	}
}

func TestImpactCmd_RequiresRef(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--db", ":memory:", "impact"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no ref provided")
	}
}
