package cli

import (
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newTemplateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage document schema templates",
	}

	cmd.AddCommand(newTemplateSetCmd())
	cmd.AddCommand(newTemplateCheckCmd())

	return cmd
}

func newTemplateSetCmd() *cobra.Command {
	var min int
	cmd := &cobra.Command{
		Use:   "set [doc-type] [required-kind]",
		Short: "Set a schema template requirement",
		Example: `  remmd template set spec overview --min 1
  remmd template set design decision --min 2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			t := core.SchemaTemplate{
				DocType:      args[0],
				RequiredKind: args[1],
				MinCount:     max(min, 1),
			}
			if err := a.Templates.SetTemplate(cmd.Context(), t); err != nil {
				return fmt.Errorf("set template: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "template set: %s requires >=%d %q section(s)\n", args[0], t.MinCount, args[1])
			return nil
		},
	}
	cmd.Flags().IntVar(&min, "min", 1, "Minimum count of required section kind")
	return cmd
}

func newTemplateCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [doc-id]",
		Short: "Check a document against its schema templates",
		Example: `  remmd template check <doc-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			doc, err := a.Docs.FindDocumentByID(ctx, args[0])
			if err != nil {
				return fmt.Errorf("find document: %w", err)
			}

			if doc.DocType == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "document has no type set; nothing to check")
				return nil
			}

			templates, err := a.Templates.GetTemplates(ctx, doc.DocType)
			if err != nil {
				return fmt.Errorf("get templates: %w", err)
			}

			if len(templates) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no templates defined for type %q\n", doc.DocType)
				return nil
			}

			sections, err := a.Docs.ListSections(ctx, args[0])
			if err != nil {
				return fmt.Errorf("list sections: %w", err)
			}

			violations := core.ValidateAgainstTemplates(sections, templates)

			if isJSON(cmd) {
				type violationEntry struct {
					DocType string `json:"doc_type"`
					Kind    string `json:"kind"`
					Want    int    `json:"want"`
					Got     int    `json:"got"`
				}
				entries := make([]violationEntry, 0, len(violations))
				for _, v := range violations {
					entries = append(entries, violationEntry{
						DocType: v.DocType, Kind: v.Kind,
						Want: v.Want, Got: v.Got,
					})
				}
				return writeJSON(cmd.OutOrStdout(), entries)
			}

			if len(violations) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "all template requirements met")
				return nil
			}

			for _, v := range violations {
				fmt.Fprintf(cmd.OutOrStdout(), "VIOLATION: %s requires >=%d %q section(s), got %d\n", v.DocType, v.Want, v.Kind, v.Got)
			}
			return fmt.Errorf("%d template violation(s)", len(violations))
		},
	}
}
