package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lagz0ne/remmd/internal/app"
	"github.com/spf13/cobra"
)

type appKey struct{}

const cliPrincipal = "cli-user"

// GetApp retrieves the App from the command's context.
// Returns nil if no App has been initialized (e.g. help, health commands).
func GetApp(cmd *cobra.Command) *app.App {
	v := cmd.Context().Value(appKey{})
	if v == nil {
		return nil
	}
	return v.(*app.App)
}

// RequireApp retrieves the App from the command's context, returning an error if not initialized.
func RequireApp(cmd *cobra.Command) (*app.App, error) {
	if a := GetApp(cmd); a != nil {
		return a, nil
	}
	return nil, fmt.Errorf("database not initialized (use --db flag)")
}

func defaultDBPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".remmd")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "remmd.db")
}

// skipAppInit returns true for commands that don't need a database.
func skipAppInit(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "health", "help", "completion":
		return true
	}
	return false
}

// NewRootCmd creates the root cobra command for remmd.
func NewRootCmd() *cobra.Command {
	var (
		debug  bool
		dbPath string
	)

	cmd := &cobra.Command{
		Use:   "remmd",
		Short: "Canvas-oriented document platform with verification agreements",
		Long: `remmd is a document platform where verification agreements between
sections create a trust network. When content changes, linked sections
go SUSPECT, requiring human review.`,
		Example: `  remmd health          Check service health
  remmd --help          Show this help`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level := slog.LevelInfo
			if debug {
				level = slog.LevelDebug
			}
			handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
			slog.SetDefault(slog.New(handler))

			if skipAppInit(cmd) {
				return nil
			}

			application, err := app.New(dbPath)
			if err != nil {
				return err
			}
			ctx := context.WithValue(cmd.Context(), appKey{}, application)
			cmd.SetContext(ctx)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if a := GetApp(cmd); a != nil {
				return a.Close()
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath(), "Database file path")

	cmd.AddCommand(newHealthCmd())
	cmd.AddCommand(newDocCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newImpactCmd())
	cmd.AddCommand(newLinkGroupCmd())
	cmd.AddCommand(newTagCmd())

	return cmd
}

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check service health",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.DebugContext(cmd.Context(), "health check invoked", "op", "health.check")
			result := map[string]any{
				"status": "ok",
				"checks": map[string]any{},
			}
			out, err := json.Marshal(result)
			if err != nil {
				slog.ErrorContext(cmd.Context(), "health check failed",
					"op", "health.check",
					"error", err.Error(),
				)
				return fmt.Errorf("marshal health: %w", err)
			}
			slog.InfoContext(cmd.Context(), "health check completed",
				"op", "health.check",
				"outcome", "ok",
			)
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}
