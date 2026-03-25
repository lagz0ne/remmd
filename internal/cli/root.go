package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

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
	return discoverDBPath()
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
		debug      bool
		dbPath     string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "remmd",
		Short: "Document agreements through verified links",
		Long: `remmd — document agreements through verified links

Create documents, link sections across them, and track trust.
Content edits are immediate. Links require bilateral approval.
When linked content changes, the graph walks and counterparties review.`,
		Example: `  # Create a document with sections (auto @refs)
  remmd doc create "API Spec" --content "# Auth\nJWT bearer tokens\n# Errors\nStructured JSON"

  # View sections and their @refs
  remmd show @s1

  # Propose a link between sections across documents
  remmd link propose @s1 --implements @s2 --rationale "code implements auth spec"

  # Both sides approve → link becomes ALIGNED
  remmd link approve <link-id>

  # Edit content → see who's impacted → reaffirm
  remmd edit @s1 --content "Updated auth flow"
  remmd impact @s1
  remmd link reaffirm <link-id>

  # Subscribe to tag notifications
  remmd tag subscribe @s1 --tag payment`,
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

			slog.Debug("database path resolved", "path", dbPath)

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
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	contentGroup := &cobra.Group{ID: "content", Title: "Content:"}
	linksGroup := &cobra.Group{ID: "links", Title: "Links:"}
	subscriptionsGroup := &cobra.Group{ID: "subscriptions", Title: "Subscriptions:"}
	systemGroup := &cobra.Group{ID: "system", Title: "System:"}

	cmd.AddGroup(contentGroup, linksGroup, subscriptionsGroup, systemGroup)

	addToGroup := func(group *cobra.Group, c *cobra.Command) {
		c.GroupID = group.ID
		cmd.AddCommand(c)
	}

	addToGroup(contentGroup, newDocCmd())
	addToGroup(contentGroup, newShowCmd())
	addToGroup(contentGroup, newEditCmd())
	addToGroup(contentGroup, newDeleteCmd())
	addToGroup(contentGroup, newSearchCmd())
	addToGroup(contentGroup, newFindCmd())
	addToGroup(contentGroup, newImportCmd())

	addToGroup(linksGroup, newLinkGroupCmd())
	addToGroup(linksGroup, newImpactCmd())
	addToGroup(linksGroup, newRelationGroupCmd())

	addToGroup(subscriptionsGroup, newTagCmd())
	addToGroup(subscriptionsGroup, newTemplateGroupCmd())

	addToGroup(systemGroup, newServeCmd())
	addToGroup(systemGroup, newHealthCmd())
	addToGroup(systemGroup, newMigrateRefsCmd())

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
