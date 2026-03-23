package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
)

func TestNewRootCmd_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}
}

func TestNewRootCmd_UseName(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	if cmd.Use != "remmd" {
		t.Fatalf("expected Use = %q, got %q", "remmd", cmd.Use)
	}
}

func TestNewRootCmd_HasHealthSubcommand(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "health" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("root command missing 'health' subcommand")
	}
}

func TestHealthSubcommand_OutputsJSON(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"health"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("health command returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("health output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if _, ok := result["status"]; !ok {
		t.Fatalf("health JSON missing 'status' field: %s", buf.String())
	}
	if result["status"] != "ok" {
		t.Fatalf("expected status = %q, got %q", "ok", result["status"])
	}
}

func TestRootCmd_HelpIncludesExamples(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	_ = cmd.Execute()

	help := buf.String()
	if !strings.Contains(help, "Examples:") && !strings.Contains(help, "examples") && !strings.Contains(help, "Example") {
		t.Fatalf("help output missing usage examples:\n%s", help)
	}
}
