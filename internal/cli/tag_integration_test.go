package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lagz0ne/remmd/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestTagSubscribe_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cmd := cli.NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", dbPath, "tag", "subscribe", "@a1", "--tag", "payment"})

	require.NoError(t, cmd.Execute())
	out := buf.String()
	require.Contains(t, out, "@a1")
	require.Contains(t, out, "payment")
	require.Contains(t, out, "subscribed")
}

func TestTagSubscribeAndList_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Subscribe
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "tag", "subscribe", "@a1", "--tag", "payment"})
	require.NoError(t, cmd1.Execute())

	// List
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "tag", "list"})
	require.NoError(t, cmd2.Execute())

	out := buf2.String()
	require.Contains(t, out, "@a1")
	require.Contains(t, out, "payment")
	require.NotContains(t, out, "(no active subscriptions)")
}

func TestTagSubscribeAndUnsubscribe_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Subscribe
	cmd1 := cli.NewRootCmd()
	buf1 := &bytes.Buffer{}
	cmd1.SetOut(buf1)
	cmd1.SetArgs([]string{"--db", dbPath, "tag", "subscribe", "@a1", "--tag", "auth"})
	require.NoError(t, cmd1.Execute())

	// Extract subscription ID from output: "subscribed @a1 to tag "auth" (id: <ULID>)"
	subID := extractSubIDFromSubscribe(t, buf1.String())

	// Unsubscribe
	cmd2 := cli.NewRootCmd()
	buf2 := &bytes.Buffer{}
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"--db", dbPath, "tag", "unsubscribe", subID})
	require.NoError(t, cmd2.Execute())
	require.Contains(t, buf2.String(), "unsubscribed")

	// List should be empty now
	cmd3 := cli.NewRootCmd()
	buf3 := &bytes.Buffer{}
	cmd3.SetOut(buf3)
	cmd3.SetArgs([]string{"--db", dbPath, "tag", "list"})
	require.NoError(t, cmd3.Execute())
	require.Contains(t, buf3.String(), "(no active subscriptions)")
}

func TestTagMultipleSubscriptions_Integration(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Subscribe two sections to different tags
	for _, tc := range []struct {
		ref string
		tag string
	}{
		{"@a1", "payment"},
		{"@b1", "auth"},
	} {
		cmd := cli.NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--db", dbPath, "tag", "subscribe", tc.ref, "--tag", tc.tag})
		require.NoError(t, cmd.Execute())
	}

	// List should show both
	cmd := cli.NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--db", dbPath, "tag", "list"})
	require.NoError(t, cmd.Execute())

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Equal(t, 2, len(lines), "expected 2 subscriptions, got:\n%s", out)
}

// extractSubIDFromSubscribe extracts the ULID from output:
// 'subscribed @a1 to tag "auth" (id: <ULID>)'
func extractSubIDFromSubscribe(t *testing.T, output string) string {
	t.Helper()
	prefix := "(id: "
	idx := strings.Index(output, prefix)
	if idx < 0 {
		t.Fatalf("could not find %q in output: %s", prefix, output)
	}
	rest := output[idx+len(prefix):]
	endIdx := strings.Index(rest, ")")
	if endIdx < 0 {
		t.Fatalf("could not find closing paren in: %s", output)
	}
	return rest[:endIdx]
}
