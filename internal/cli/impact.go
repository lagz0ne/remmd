package cli

import (
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newImpactCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "impact [ref]",
		Short: "Show blast radius before changing content",
		Long:  "Walk the graph from a section to find all impacted links and counterparties. Run this before editing to see who will need to review.",
		Example: `  remmd impact @s1              # who's affected if I change this?
  remmd impact @ext:jira/ENG-42  # external section impact`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			ref := args[0]

			sec, _, err := findSectionByRef(ctx, a.Docs, ref)
			if err != nil {
				return fmt.Errorf("section not found: %s", ref)
			}

			walker := newWalker(a.Links)
			impacted, err := walker.WalkFromSection(ctx, sec.ID)
			if err != nil {
				return fmt.Errorf("graph walk: %w", err)
			}

			if isJSON(cmd) {
				if impacted == nil {
					impacted = []core.ImpactedLink{}
				}
				return writeJSON(cmd.OutOrStdout(), impacted)
			}

			br := core.NewBlastRadius(sec.ID, impacted)
			fmt.Fprintln(cmd.OutOrStdout(), br.Summary())

			for _, g := range br.Groups {
				for _, il := range g.Links {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s (changed: %s)\n",
						il.State, il.RelationshipType, il.ChangedSide)
				}
			}

			return nil
		},
	}
}
