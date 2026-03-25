package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

			srv, err := serve.New(application, viteDir, serve.WithPort(port))
			if err != nil {
				return err
			}
			return srv.Start(cmd.Context())
		},
	}

	cmd.Flags().IntVar(&port, "port", 4312, "HTTP listen port")

	return cmd
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
