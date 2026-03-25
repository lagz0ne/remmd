package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

type importEntry struct {
	Title       string   `json:"title"`
	DocType     string   `json:"doc_type,omitempty"`
	ParentDocID string   `json:"parent_doc_id,omitempty"`
	Content     string   `json:"content,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type importResult struct {
	ID       string               `json:"id"`
	Title    string               `json:"title"`
	Sections []importSectionResult `json:"sections,omitempty"`
	Error    string               `json:"error,omitempty"`
}

type importSectionResult struct {
	Ref  string `json:"ref"`
	Kind string `json:"kind,omitempty"`
}

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Bulk import documents from JSONL on stdin",
		Long: `Read JSON lines from stdin, create one document per line.
Each line: {"title": "...", "doc_type": "...", "content": "...", "tags": [...]}
Outputs one JSON result per line.`,
		Example: `  echo '{"title":"Auth","doc_type":"c3:component","content":"# Goal\nHandle auth"}' | remmd import
  cat entities.jsonl | remmd import`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB lines
			enc := json.NewEncoder(cmd.OutOrStdout())

			count := 0
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}

				var entry importEntry
				if err := json.Unmarshal(line, &entry); err != nil {
					enc.Encode(importResult{Error: fmt.Sprintf("parse: %v", err)})
					continue
				}

				doc := core.NewDocument(entry.Title, cliPrincipal)
				doc.DocType = entry.DocType
				doc.ParentDocID = entry.ParentDocID

				if err := a.Docs.CreateDocument(ctx, doc); err != nil {
					enc.Encode(importResult{Title: entry.Title, Error: fmt.Sprintf("create: %v", err)})
					continue
				}

				result := importResult{ID: doc.ID, Title: entry.Title}

				if entry.Content != "" {
					sections := core.Parse(doc.ID, entry.Content, 0)
					if len(sections) > 0 {
						startSeq, err := a.Docs.NextRefSeq(ctx, len(sections))
						if err != nil {
							enc.Encode(importResult{ID: doc.ID, Title: entry.Title, Error: fmt.Sprintf("refs: %v", err)})
							continue
						}
						core.RebaseRefs(sections, startSeq)

						if err := a.Docs.CreateSections(ctx, sections); err != nil {
							enc.Encode(importResult{ID: doc.ID, Title: entry.Title, Error: fmt.Sprintf("sections: %v", err)})
							continue
						}

						for _, s := range sections {
							result.Sections = append(result.Sections, importSectionResult{
								Ref: s.Ref.String(), Kind: s.Kind,
							})
						}

						// Apply tags to first section
						if len(entry.Tags) > 0 {
							for _, tag := range entry.Tags {
								a.Docs.AddTag(ctx, sections[0].ID, tag)
							}
						}
					}
				}

				enc.Encode(result)
				count++
			}

			if !isJSON(cmd) {
				fmt.Fprintf(cmd.ErrOrStderr(), "imported %d documents\n", count)
			}
			return scanner.Err()
		},
	}
}
