package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Full-text search across sections",
		Example: `  remmd search "authentication"
  remmd search "JWT bearer"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			sections, err := a.Docs.SearchSections(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			if isJSON(cmd) {
				type searchResult struct {
					Ref   string `json:"ref"`
					Kind  string `json:"kind,omitempty"`
					Title string `json:"title"`
					DocID string `json:"doc_id"`
				}
				results := make([]searchResult, 0, len(sections))
				for _, s := range sections {
					results = append(results, searchResult{
						Ref:   s.Ref.String(),
						Kind:  s.Kind,
						Title: s.Title,
						DocID: s.DocID,
					})
				}
				return writeJSON(cmd.OutOrStdout(), results)
			}

			if len(sections) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no results)")
				return nil
			}

			titleCache := make(map[string]string)
			for _, s := range sections {
				docTitle, ok := titleCache[s.DocID]
				if !ok {
					if doc, err := a.Docs.FindDocumentByID(cmd.Context(), s.DocID); err == nil {
						docTitle = doc.Title
					} else {
						docTitle = s.DocID
					}
					titleCache[s.DocID] = docTitle
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s %q\n", docTitle, s.Ref.String(), s.Title)
				if s.Content != "" && s.Content != s.Title {
					snippet := s.Content
					if len(snippet) > 100 {
						snippet = snippet[:100] + "..."
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", snippet)
				}
			}
			return nil
		},
	}
}
