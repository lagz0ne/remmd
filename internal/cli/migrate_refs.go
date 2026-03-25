package cli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newMigrateRefsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-refs",
		Short: "Re-number all section refs to be globally unique",
		Long:  "Assigns globally unique refs across all documents. Fixes ref collisions where multiple documents share @a1, @b2, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			db := a.DB

			docs, err := a.Docs.ListDocuments(ctx)
			if err != nil {
				return fmt.Errorf("list documents: %w", err)
			}

			// Collect all sections across all documents
			type sectionEntry struct {
				id    string
				docID string
			}
			var all []sectionEntry
			for _, doc := range docs {
				sections, err := a.Docs.ListSections(ctx, doc.ID)
				if err != nil {
					return fmt.Errorf("list sections for %s: %w", doc.Title, err)
				}
				for _, s := range sections {
					all = append(all, sectionEntry{id: s.ID, docID: s.DocID})
				}
			}

			if len(all) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no sections to migrate")
				return nil
			}

			// Pass 1: rename all refs to temporary UUIDs to avoid UNIQUE collisions
			fmt.Fprintf(cmd.OutOrStdout(), "pass 1: clearing %d refs...\n", len(all))
			for i, s := range all {
				tmpRef := fmt.Sprintf("@tmp_%d", i)
				if err := execUpdate(ctx, db, s.id, tmpRef); err != nil {
					return fmt.Errorf("pass 1 section %s: %w", s.id, err)
				}
			}

			// Pass 2: assign globally unique refs in order
			fmt.Fprintf(cmd.OutOrStdout(), "pass 2: assigning unique refs...\n")
			seq := 1
			docIdx := 0
			sectionIdx := 0
			for _, doc := range docs {
				sections, _ := a.Docs.ListSections(ctx, doc.ID)
				if len(sections) == 0 {
					docIdx++
					continue
				}
				startSeq := seq
				for range sections {
					newRef := core.NewRef(doc.ID, seq)
					if err := execUpdate(ctx, db, all[sectionIdx].id, newRef.String()); err != nil {
						return fmt.Errorf("pass 2 section %s: %w", all[sectionIdx].id, err)
					}
					seq++
					sectionIdx++
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d sections (%s..%s)\n",
					doc.Title, len(sections),
					core.NewRef(doc.ID, startSeq).String(),
					core.NewRef(doc.ID, seq-1).String(),
				)
				docIdx++
			}

			// Update ref_counter
			if _, err := db.ExecContext(ctx, "UPDATE ref_counter SET next_seq = ? WHERE id = 1", seq); err != nil {
				return fmt.Errorf("update ref_counter: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nmigrated %d sections across %d documents\n", len(all), len(docs))
			return nil
		},
	}
}

func execUpdate(ctx context.Context, db *sql.DB, sectionID, newRef string) error {
	_, err := db.ExecContext(ctx, "UPDATE sections SET ref = ? WHERE id = ?", newRef, sectionID)
	return err
}
