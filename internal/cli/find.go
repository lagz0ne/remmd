package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find [tag]",
		Short: "Find a section by tag",
		Long:  "Search all sections for a specific tag and display matching results.",
		Example: `  remmd find "c3:c3-101"
  remmd find "sft:screen:1"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			tag := args[0]

			sections, err := a.Docs.FindSectionsByTag(ctx, tag)
			if err != nil {
				return fmt.Errorf("find by tag: %w", err)
			}

			if len(sections) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(not found)")
				return nil
			}

			for _, s := range sections {
				if isJSON(cmd) {
					writeJSON(cmd.OutOrStdout(), map[string]any{
						"ref": s.Ref.String(), "doc_id": s.DocID,
						"kind": s.Kind, "title": s.Title,
					})
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", s.Ref.String(), s.DocID, s.Title)
				}
			}
			return nil
		},
	}
}
