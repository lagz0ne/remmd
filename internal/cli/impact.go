package cli

import (
	"fmt"

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
			ref := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "impact analysis for %s\n", ref)
			return nil
		},
	}
}
