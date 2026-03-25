package cli

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/remmd/internal/serve"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the dev server (embedded NATS + Vite)",
		Long:  "Starts remmd with an embedded NATS message bus and proxies the Vite frontend. All data flows through NATS request-reply.",
		RunE: func(cmd *cobra.Command, args []string) error {
			application, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			viteDir := resolveViteDir()
			if viteDir == "" {
				return fmt.Errorf("cannot find view/ directory; set REMMD_PACKAGE_DIR or run from the project root")
			}

			if port == 0 {
				port = derivePort()
			}

			srv, err := serve.New(application, viteDir, serve.WithPort(port))
			if err != nil {
				return err
			}
			return srv.Start(cmd.Context())
		},
	}

	cmd.Flags().IntVar(&port, "port", 0, "HTTP listen port (default: derived from project name + git branch)")

	return cmd
}

// derivePort generates a deterministic port from the project directory name
// and git branch. Same project+branch always gets the same port.
// Range: 4000-9999 (~6000 slots, enough to avoid collisions).
func derivePort() int {
	wd, _ := os.Getwd()
	key := filepath.Base(wd)
	if branch := gitBranch(); branch != "" {
		key += ":" + branch
	}

	h := sha256.Sum256([]byte(key))
	n := binary.BigEndian.Uint32(h[:4])
	port := int(n%6000) + 4000

	slog.Info("port derived", "project", filepath.Base(wd), "branch", gitBranch(), "port", port)
	return port
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// resolveViteDir finds the frontend view directory.
// Priority: REMMD_PACKAGE_DIR env → relative to binary → working directory.
func resolveViteDir() string {
	if dir := os.Getenv("REMMD_PACKAGE_DIR"); dir != "" {
		return filepath.Join(dir, "view")
	}

	// Try relative to the binary
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "view")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	// Try working directory
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, "view")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}
