package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
			if err := a.Docs.CreateDocument(ctx, doc); err != nil {
				return fmt.Errorf("create document: %w", err)
			}

			var sections []core.Section
			if content != "" {
				sections = core.Parse(doc.ID, content)
				for i := range sections {
					if err := a.Docs.CreateSection(ctx, &sections[i]); err != nil {
						return fmt.Errorf("create section: %w", err)
					}
				}
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
	return cmd
}

func newDocListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all documents",
		Example: `  remmd doc list
  remmd doc ls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			docs, err := a.Docs.ListDocuments(ctx)
			if err != nil {
				return fmt.Errorf("list documents: %w", err)
			}

			if len(docs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no documents)")
				return nil
			}

			for _, doc := range docs {
				sections, _ := a.Docs.ListSections(ctx, doc.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  (%d sections)\n", doc.ID, doc.Title, len(sections))
			}
			return nil
		},
	}
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
				fmt.Fprintf(cmd.OutOrStdout(), "%s %q (%d sections)\n", doc.ID, doc.Title, len(sections))
				secs := make([]core.Section, len(sections))
				for i, s := range sections {
					secs[i] = *s
				}
				printSectionTree(cmd.OutOrStdout(), secs)
				return nil
			}

			sec, _, lookupErr := findSectionByRef(ctx, a.Docs, ref)
			if lookupErr != nil {
				return fmt.Errorf("not found: %s", ref)
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
					return nil
				}
			}

			if content != "" {
				hash := core.ContentHash(content)
				if err := a.Docs.UpdateSectionContent(ctx, sec.ID, content, hash); err != nil {
					return fmt.Errorf("update content: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s content updated (%d bytes)\n", ref, len(content))
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
