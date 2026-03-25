package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lagz0ne/remmd/internal/app"
	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

// findSectionByRef searches for a section matching the given ref string across all documents.
// Returns the section, the owning document ID, and any error.
func findSectionByRef(ctx context.Context, docs core.DocumentRepository, refStr string) (*core.Section, string, error) {
	return docs.FindSectionByRefGlobal(ctx, refStr)
}

func newDocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Create and list documents",
		Long:  "Create documents from content (markdown parsed into sections with auto @refs), list all documents.",
	}
	cmd.AddCommand(newDocCreateCmd())
	cmd.AddCommand(newDocListCmd())
	cmd.AddCommand(newDocDeleteCmd())
	cmd.AddCommand(newDocArchiveCmd())
	cmd.AddCommand(newSectionAddCmd())
	return cmd
}

func newDocCreateCmd() *cobra.Command {
	var (
		content    string
		external   bool
		system     string
		externalID string
		hash       string
		metadata   string
		docType    string
		parentID   string
		tags       []string
	)
	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a new document",
		Long:  "Create a new document from a title. Content can be provided via --content flag or stdin.",
		Example: `  remmd doc create "API Specification"
  remmd doc create "Design Doc" --content "# Overview\nFirst draft"
  remmd doc create "Notion Page" --external --system notion --external-id page-abc --hash sha256abc`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			title := args[0]

			if external {
				if system == "" || externalID == "" || hash == "" {
					return fmt.Errorf("--external requires --system, --external-id, and --hash")
				}
				if metadata != "" && !json.Valid([]byte(metadata)) {
					return fmt.Errorf("--metadata must be valid JSON")
				}

				doc := core.NewDocument(title, cliPrincipal)
				doc.Source = system
				doc.DocType = docType
				doc.ParentDocID = parentID
				if err := a.Docs.CreateDocument(ctx, doc); err != nil {
					return fmt.Errorf("create document: %w", err)
				}

				sec := core.Section{
					ID:          core.NewID().String(),
					Ref:         core.NewExternalRef(system, externalID),
					DocID:       doc.ID,
					Type:        core.SectionHeading,
					Title:       title,
					Content:     "",
					ContentHash: hash,
					ContentType: core.ContentExternal,
					Metadata:    metadata,
					Order:       0,
				}
				if err := a.Docs.CreateSection(ctx, &sec); err != nil {
					return fmt.Errorf("create section: %w", err)
				}

				if isJSON(cmd) {
					return writeJSON(cmd.OutOrStdout(), map[string]any{
						"id":       doc.ID,
						"title":    doc.Title,
						"doc_type": doc.DocType,
						"sections": []map[string]any{
							{"ref": sec.Ref.String(), "kind": "", "title": sec.Title},
						},
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "doc %s %q created (1 external section %s)\n", doc.ID, doc.Title, sec.Ref.String())
				return nil
			}

			if content == "" {
				stat, _ := cmd.InOrStdin().(interface{ Stat() (interface{ Size() int64 }, error) })
				if stat == nil {
					// Try reading from stdin if it's not a terminal
					data, err := io.ReadAll(cmd.InOrStdin())
					if err == nil && len(data) > 0 {
						content = string(data)
					}
				}
			}

			doc := core.NewDocument(title, cliPrincipal)
			doc.DocType = docType
			doc.ParentDocID = parentID
			if err := a.Docs.CreateDocument(ctx, doc); err != nil {
				return fmt.Errorf("create document: %w", err)
			}

			var sections []core.Section
			if content != "" {
				sections = core.Parse(doc.ID, content, 0)
				if len(sections) > 0 {
					startSeq, err := a.Docs.NextRefSeq(ctx, len(sections))
					if err != nil {
						return fmt.Errorf("reserve refs: %w", err)
					}
					core.RebaseRefs(sections, startSeq)
				}
				for i := range sections {
					if err := a.Docs.CreateSection(ctx, &sections[i]); err != nil {
						return fmt.Errorf("create section: %w", err)
					}
				}
			}

			// Apply tags to the first section
			if len(tags) > 0 && len(sections) > 0 {
				for _, tag := range tags {
					if err := a.Docs.AddTag(ctx, sections[0].ID, tag); err != nil {
						return fmt.Errorf("add tag %q: %w", tag, err)
					}
				}
			}

			if isJSON(cmd) {
				var sectionRefs []map[string]any
				for _, s := range sections {
					sectionRefs = append(sectionRefs, map[string]any{
						"ref":   s.Ref.String(),
						"kind":  s.Kind,
						"title": s.Title,
					})
				}
				if sectionRefs == nil {
					sectionRefs = []map[string]any{}
				}
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"id":       doc.ID,
					"title":    doc.Title,
					"doc_type": doc.DocType,
					"sections": sectionRefs,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "doc %s %q created (%d sections)\n", doc.ID, doc.Title, len(sections))
			printSectionTree(cmd.OutOrStdout(), sections)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "Inline document content (markdown)")
	cmd.Flags().BoolVar(&external, "external", false, "Create an external content document")
	cmd.Flags().StringVar(&system, "system", "", "External system name (e.g. notion, figma)")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External resource ID")
	cmd.Flags().StringVar(&hash, "hash", "", "Content hash for external section")
	cmd.Flags().StringVar(&metadata, "metadata", "", "JSON metadata for external section")
	cmd.Flags().StringVar(&docType, "type", "", "Document type (e.g. spec, design, runbook)")
	cmd.Flags().StringVar(&parentID, "parent", "", "Parent document ID")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply to the first section (repeatable)")
	return cmd
}

func newDocListCmd() *cobra.Command {
	var showAll bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all documents",
		Long:    "List documents. Archived documents are hidden by default; use --all to include them.",
		Example: `  remmd doc list
  remmd doc ls
  remmd doc list --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			summaries, err := a.Docs.ListDocumentsWithSectionCounts(ctx)
			if err != nil {
				return fmt.Errorf("list documents: %w", err)
			}

			var filtered []core.DocumentSummary
			for _, s := range summaries {
				if !showAll && s.Document.Status == core.DocumentArchived {
					continue
				}
				filtered = append(filtered, s)
			}

			if isJSON(cmd) {
				type docEntry struct {
					ID           string `json:"id"`
					Title        string `json:"title"`
					DocType      string `json:"doc_type,omitempty"`
					ParentDocID  string `json:"parent_doc_id,omitempty"`
					SectionCount int    `json:"section_count"`
				}
				entries := make([]docEntry, 0, len(filtered))
				for _, s := range filtered {
					entries = append(entries, docEntry{
						ID:           s.Document.ID,
						Title:        s.Document.Title,
						DocType:      s.Document.DocType,
						ParentDocID:  s.Document.ParentDocID,
						SectionCount: s.SectionCount,
					})
				}
				return writeJSON(cmd.OutOrStdout(), entries)
			}

			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no documents)")
				return nil
			}

			for _, s := range filtered {
				status := ""
				if s.Document.Status == core.DocumentArchived {
					status = " [archived]"
				}
				if s.Document.DocType != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s]%s  (%d sections)\n", s.Document.ID, s.Document.Title, s.Document.DocType, status, s.SectionCount)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %s%s  (%d sections)\n", s.Document.ID, s.Document.Title, status, s.SectionCount)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all documents including archived")
	return cmd
}

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [ref]",
		Short: "Show a document or section by @ref",
		Long:  "Display the contents of a document or a specific section identified by its @ref.",
		Example: `  remmd show @a1
  remmd show @b2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			ref := args[0]

			doc, err := a.Docs.FindDocumentByID(ctx, ref)
			if err == nil {
				sections, _ := a.Docs.ListSections(ctx, doc.ID)
				if isJSON(cmd) {
					type sectionEntry struct {
						Ref     string `json:"ref"`
						Kind    string `json:"kind,omitempty"`
						Title   string `json:"title"`
						Content string `json:"content,omitempty"`
					}
					secs := make([]sectionEntry, 0, len(sections))
					for _, s := range sections {
						secs = append(secs, sectionEntry{
							Ref:     s.Ref.String(),
							Kind:    s.Kind,
							Title:   s.Title,
							Content: s.Content,
						})
					}
					return writeJSON(cmd.OutOrStdout(), map[string]any{
						"doc_id":   doc.ID,
						"title":    doc.Title,
						"sections": secs,
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %q (%d sections)\n", doc.ID, doc.Title, len(sections))
				secsCopy := make([]core.Section, len(sections))
				for i, s := range sections {
					secsCopy[i] = *s
				}
				printSectionTree(cmd.OutOrStdout(), secsCopy)
				return nil
			}

			sec, _, lookupErr := findSectionByRef(ctx, a.Docs, ref)
			if lookupErr != nil {
				return fmt.Errorf("not found: %s", ref)
			}

			if isJSON(cmd) {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"ref":     sec.Ref.String(),
					"kind":    sec.Kind,
					"title":   sec.Title,
					"content": sec.Content,
					"doc_id":  sec.DocID,
				})
			}

			if sec.ContentType == core.ContentExternal {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (external) system=%s hash=%s\n", sec.Ref.String(), sec.Ref.System(), sec.ContentHash)
				if sec.Metadata != "" && sec.Metadata != "{}" {
					fmt.Fprintf(cmd.OutOrStdout(), "  metadata: %s\n", sec.Metadata)
				}
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s\n", sec.Ref.String(), sec.Type, sec.Title)
			if sec.Content != sec.Title {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", sec.Content)
			}

			versions, verErr := a.Docs.GetSectionVersions(ctx, sec.ID)
			if verErr == nil && len(versions) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%d version(s)\n", len(versions))
			}
			return nil
		},
	}
}

func newEditCmd() *cobra.Command {
	var (
		content  string
		tag      string
		hashFlag string
	)
	cmd := &cobra.Command{
		Use:   "edit [ref]",
		Short: "Edit a section by @ref",
		Long:  "Update the content or tags of a section identified by its @ref.",
		Example: `  remmd edit @a1 --content "Updated content"
  remmd edit @a1 --tag "api"
  remmd edit @ext:notion/page-abc --hash newhash`,
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

			if sec.ContentType == core.ContentExternal {
				if content != "" {
					return fmt.Errorf("cannot set body content on external section")
				}
				if hashFlag != "" {
					if err := a.Docs.UpdateSectionContent(ctx, sec.ID, "", hashFlag); err != nil {
						return fmt.Errorf("update hash: %w", err)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s hash updated to %q\n", ref, hashFlag)
					walkStaleLinks(ctx, cmd, a, sec.ID, ref)
					return nil
				}
			}

			if content != "" {
				hash := core.ContentHash(content)
				if err := a.Docs.UpdateSectionContent(ctx, sec.ID, content, hash); err != nil {
					return fmt.Errorf("update content: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s content updated (%d bytes)\n", ref, len(content))
				walkStaleLinks(ctx, cmd, a, sec.ID, ref)
			}

			if tag != "" {
				if err := a.Docs.AddTag(ctx, sec.ID, tag); err != nil {
					return fmt.Errorf("add tag: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s tagged %q\n", ref, tag)
			}

			if content == "" && tag == "" && hashFlag == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (no changes)\n", ref)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "New section content")
	cmd.Flags().StringVar(&tag, "tag", "", "Tag to apply to section")
	cmd.Flags().StringVar(&hashFlag, "hash", "", "New content hash (for external sections)")
	return cmd
}

// walkStaleLinks transitions aligned links to stale after a section change.
func walkStaleLinks(ctx context.Context, cmd *cobra.Command, a *app.App, sectionID, ref string) {
	walker := newWalker(a.Links)
	impacted, walkErr := walker.WalkFromSection(ctx, sectionID)
	if walkErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: graph walk failed: %v\n", walkErr)
		return
	}
	staleCount := 0
	for _, il := range impacted {
		if il.State == string(core.LinkAligned) {
			if err := a.Links.UpdateLinkState(ctx, il.LinkID, core.LinkStale); err == nil {
				staleCount++
				_ = a.Links.AddThreadEntry(ctx, il.LinkID, core.EntrySystem, cliPrincipal,
					fmt.Sprintf("content changed on %s — link marked stale", ref))
			}
		}
	}
	if staleCount > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "%d link(s) marked stale\n", staleCount)
	}
}

func newDeleteCmd() *cobra.Command {
	var (
		reason      string
		replacement string
	)
	cmd := &cobra.Command{
		Use:   "delete [ref]",
		Short: "Delete a section by @ref",
		Long:  "Remove a section from its document. Optionally provide a reason and replacement ref.",
		Example: `  remmd delete @a1
  remmd delete @a1 --reason "Obsolete"
  remmd delete @a1 --reason "Merged" --replacement @b2`,
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

			if err := a.Docs.DeleteSection(ctx, sec.ID); err != nil {
				return fmt.Errorf("delete section: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", ref)
			if reason != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  reason: %s\n", reason)
			}
			if replacement != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  replacement: %s\n", replacement)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for deletion")
	cmd.Flags().StringVar(&replacement, "replacement", "", "Replacement @ref")
	return cmd
}

func newDocDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete [doc-id...]",
		Short: "Delete one or more documents",
		Long:  "Permanently delete documents and all their sections, versions, and tags.",
		Example: `  remmd doc delete <doc-id>
  remmd doc delete <doc-id-1> <doc-id-2>
  remmd doc delete <doc-id> --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			for _, docID := range args {
				doc, err := a.Docs.FindDocumentByID(ctx, docID)
				if err != nil {
					return fmt.Errorf("document not found: %s", docID)
				}

				sections, _ := a.Docs.ListSections(ctx, docID)

				if !force {
					from, _ := a.Relations.ListRelationsFrom(ctx, docID)
					to, _ := a.Relations.ListRelationsTo(ctx, docID)
					if len(from)+len(to) > 0 {
						fmt.Fprintf(cmd.ErrOrStderr(), "warning: doc %s has %d relation(s), deleting anyway\n", docID, len(from)+len(to))
					}
				}

				if err := a.Docs.DeleteDocument(ctx, docID); err != nil {
					return fmt.Errorf("delete document %s: %w", docID, err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "deleted %s %q (%d sections removed)\n", docID, doc.Title, len(sections))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Skip relation warnings")
	return cmd
}

func newDocArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive [doc-id...]",
		Short: "Archive one or more documents",
		Long:  "Set documents to archived status. Archived documents are hidden from doc list by default.",
		Example: `  remmd doc archive <doc-id>
  remmd doc archive <doc-id-1> <doc-id-2>`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			for _, docID := range args {
				doc, err := a.Docs.FindDocumentByID(ctx, docID)
				if err != nil {
					return fmt.Errorf("document not found: %s", docID)
				}

				if err := a.Docs.ArchiveDocument(ctx, docID); err != nil {
					return fmt.Errorf("archive document %s: %w", docID, err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "archived %s %q\n", docID, doc.Title)
			}
			return nil
		},
	}
}

func newSectionAddCmd() *cobra.Command {
	var (
		kind    string
		content string
		secTags []string
	)
	cmd := &cobra.Command{
		Use:   "section [doc-id]",
		Short: "Add a section to an existing document",
		Example: `  remmd doc section <doc-id> --kind region --content "Credentials form"
  remmd doc section <doc-id> --kind goal --content "Handle authentication"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			docID := args[0]

			// Verify doc exists
			if _, err := a.Docs.FindDocumentByID(ctx, docID); err != nil {
				return fmt.Errorf("document not found: %w", err)
			}

			// Get current section count for ordering
			existing, err := a.Docs.ListSections(ctx, docID)
			if err != nil {
				return err
			}

			// Reserve a ref
			startSeq, err := a.Docs.NextRefSeq(ctx, 1)
			if err != nil {
				return fmt.Errorf("reserve ref: %w", err)
			}

			ref := core.NewRef(docID, startSeq)
			sec := core.Section{
				ID:          core.NewID().String(),
				Ref:         ref,
				DocID:       docID,
				Type:        core.SectionHeading,
				Kind:        kind,
				Title:       content,
				Content:     content,
				ContentHash: core.ContentHash(content),
				Order:       len(existing),
			}

			if err := a.Docs.CreateSection(ctx, &sec); err != nil {
				return fmt.Errorf("create section: %w", err)
			}

			// Apply tags to the new section
			for _, tag := range secTags {
				if err := a.Docs.AddTag(ctx, sec.ID, tag); err != nil {
					return fmt.Errorf("add tag %q: %w", tag, err)
				}
			}

			if isJSON(cmd) {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"ref":    ref.String(),
					"kind":   kind,
					"doc_id": docID,
					"id":     sec.ID,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "section %s added to %s (kind: %s)\n", ref.String(), docID, kind)
			return nil
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Section kind (e.g. region, goal, requirement)")
	cmd.Flags().StringVar(&content, "content", "", "Section content")
	cmd.Flags().StringArrayVar(&secTags, "tag", nil, "Tag to apply (repeatable)")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

// printSectionTree prints sections as an indented tree based on parent refs.
func printSectionTree(w io.Writer, sections []core.Section) {
	if len(sections) == 0 {
		return
	}

	refDepth := make(map[string]int)
	for _, s := range sections {
		depth := 0
		if s.ParentRef != nil {
			if d, ok := refDepth[s.ParentRef.String()]; ok {
				depth = d + 1
			} else {
				depth = 1
			}
		}
		refDepth[s.Ref.String()] = depth
	}

	for _, s := range sections {
		depth := refDepth[s.Ref.String()]
		indent := strings.Repeat("  ", depth+1)
		fmt.Fprintf(w, "%s%s %s\n", indent, s.Ref.String(), s.Title)
	}
}
