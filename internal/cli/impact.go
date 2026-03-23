package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newImpactCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "impact [ref]",
		Short: "Show blast radius of a section change",
		Long:  "Analyze which agreements are impacted when a section changes. Shows counterparties who need to review.",
		Example: `  remmd impact @a1
  remmd impact @b2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "impact analysis for %s\n", ref)
			return nil
		},
	}
}
