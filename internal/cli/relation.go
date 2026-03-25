package cli

import (
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newRelationGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation",
		Short: "Manage structural relations between documents",
	}

	cmd.AddCommand(newRelationCreateCmd())
	cmd.AddCommand(newRelationListCmd())
	cmd.AddCommand(newRelationDeleteCmd())

	return cmd
}

func newRelationCreateCmd() *cobra.Command {
	var relType string
	cmd := &cobra.Command{
		Use:   "create [from-doc-id] [to-doc-id]",
		Short: "Create a relation between two documents",
		Example: `  remmd relation create <from-doc-id> <to-doc-id> --type "depends_on"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			rel := core.NewRelation(args[0], args[1], relType)
			if err := a.Relations.CreateRelation(ctx, rel); err != nil {
				return fmt.Errorf("create relation: %w", err)
			}

			if isJSON(cmd) {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"id":   rel.ID,
					"from": rel.FromDocID,
					"to":   rel.ToDocID,
					"type": rel.RelationType,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "relation %s created: %s --%s-> %s\n", rel.ID, args[0], relType, args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&relType, "type", "related", "Relation type (e.g. depends_on, related, parent)")
	return cmd
}

func newRelationListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [doc-id]",
		Short: "List relations for a document",
		Example: `  remmd relation list <doc-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			from, err := a.Relations.ListRelationsFrom(ctx, args[0])
			if err != nil {
				return fmt.Errorf("list relations from: %w", err)
			}
			to, err := a.Relations.ListRelationsTo(ctx, args[0])
			if err != nil {
				return fmt.Errorf("list relations to: %w", err)
			}

			if isJSON(cmd) {
				type relEntry struct {
					ID   string `json:"id"`
					From string `json:"from"`
					To   string `json:"to"`
					Type string `json:"type"`
				}
				all := make([]relEntry, 0, len(from)+len(to))
				for _, r := range from {
					all = append(all, relEntry{ID: r.ID, From: r.FromDocID, To: r.ToDocID, Type: r.RelationType})
				}
				for _, r := range to {
					all = append(all, relEntry{ID: r.ID, From: r.FromDocID, To: r.ToDocID, Type: r.RelationType})
				}
				return writeJSON(cmd.OutOrStdout(), all)
			}

			if len(from) == 0 && len(to) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no relations)")
				return nil
			}

			// Build title cache to avoid repeated lookups
			titleCache := make(map[string]string)
			resolveTitle := func(docID string) string {
				if title, ok := titleCache[docID]; ok {
					return title
				}
				doc, err := a.Docs.FindDocumentByID(ctx, docID)
				if err != nil || doc == nil {
					titleCache[docID] = docID
					return docID
				}
				titleCache[docID] = doc.Title
				return doc.Title
			}

			for _, r := range from {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  \"%s\" --%s-> \"%s\"\n", r.ID, resolveTitle(r.FromDocID), r.RelationType, resolveTitle(r.ToDocID))
			}
			for _, r := range to {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  \"%s\" --%s-> \"%s\"\n", r.ID, resolveTitle(r.FromDocID), r.RelationType, resolveTitle(r.ToDocID))
			}
			return nil
		},
	}
}

func newRelationDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [relation-id]",
		Short: "Delete a relation",
		Example: `  remmd relation delete <relation-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			if err := a.Relations.DeleteRelation(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("delete relation: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted relation %s\n", args[0])
			return nil
		},
	}
}
