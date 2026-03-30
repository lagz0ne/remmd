package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/playbook"
	"github.com/spf13/cobra"
)

func newPlaybookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playbook",
		Short: "Playbook operations",
	}
	cmd.AddCommand(newPlaybookCheckCmd())
	cmd.AddCommand(newPlaybookImportCmd())
	cmd.AddCommand(newPlaybookValidateCmd())
	return cmd
}

func newPlaybookCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <file>",
		Short: "Validate a playbook YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read playbook: %w", err)
			}
			pb, err := playbook.Parse(data)
			if err != nil {
				return fmt.Errorf("parse playbook: %w", err)
			}

			var errs []string
			for name, edge := range pb.Edges {
				for _, from := range edge.From {
					if pb.Type(from) == nil {
						errs = append(errs, fmt.Sprintf("edge %q: source type %q not defined", name, from))
					}
				}
				if pb.Type(edge.To) == nil {
					errs = append(errs, fmt.Sprintf("edge %q: target type %q not defined", name, edge.To))
				}
			}

			checker, err := playbook.NewValidationChecker()
			if err != nil {
				return fmt.Errorf("create checker: %w", err)
			}
			for _, r := range pb.Rules {
				if err := checker.Compile(r.Expr); err != nil {
					errs = append(errs, fmt.Sprintf("rule %q: %v", r.Name, err))
				}
			}
			for typeName, td := range pb.Types {
				for _, r := range td.Rules {
					if err := checker.Compile(r.Expr); err != nil {
						errs = append(errs, fmt.Sprintf("type %q rule %q: %v", typeName, r.Name, err))
					}
				}
			}

			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(cmd.ErrOrStderr(), "ERROR:", e)
				}
				return fmt.Errorf("%d validation error(s)", len(errs))
			}

			ruleCount := len(pb.Rules)
			for _, td := range pb.Types {
				ruleCount += len(td.Rules)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "OK: %d types, %d edges, %d rules\n",
				len(pb.Types), len(pb.Edges), ruleCount)
			return nil
		},
	}
}

func newPlaybookValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Run playbook rules against all documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()

			// Try common playbook names
			var pb *playbook.Playbook
			for _, name := range []string{"default", "c3", "sft"} {
				pb, _, err = a.Playbooks.Latest(ctx, name)
				if err != nil {
					return fmt.Errorf("load playbook: %w", err)
				}
				if pb != nil {
					break
				}
			}
			if err != nil {
				return fmt.Errorf("load playbook: %w", err)
			}
			if pb == nil {
				return fmt.Errorf("no playbook imported — run 'remmd playbook import' first")
			}

			docs, err := a.Docs.ListDocuments(ctx)
			if err != nil {
				return fmt.Errorf("list documents: %w", err)
			}

			var nodes []playbook.Node
			for _, d := range docs {
				data := map[string]any{
					"_node_id":         d.ID,
					"title":            d.Title,
					"status":           string(d.Status),
					"source":           d.Source,
					"owner":            "",
					"origin":           "",
					"golden_example":   "",
					"responsibilities": "",
					"affects":          []any{},
					"sources":          []any{},
				}
				// Enrich with section content as fields
				if sections, err := a.Docs.ListSections(ctx, d.ID); err == nil {
					for _, s := range sections {
						key := strings.ToLower(strings.ReplaceAll(s.Title, " ", "_"))
						if key != "" && s.Content != "" {
							data[key] = s.Content
						}
					}
				}
				nodes = append(nodes, playbook.Node{
					Type: d.DocType,
					ID:   d.ID,
					Data: data,
				})
			}

			relations, err := a.Relations.ListAllRelations(ctx)
			if err != nil {
				return fmt.Errorf("list relations: %w", err)
			}

			gc := newRelationGraphForValidation(relations, nodes)
			diags := playbook.RunWithGraph(pb, nodes, gc)

			if len(diags) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "PASS: %d documents, 0 issues\n", len(docs))
				return nil
			}

			errors, warnings := 0, 0
			for _, d := range diags {
				severity := strings.ToUpper(d.Severity)
				if d.Severity == "error" {
					errors++
				} else {
					warnings++
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s: %s\n", severity, d.NodeType, d.NodeID[:8], d.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d documents, %d errors, %d warnings\n", len(docs), errors, warnings)
			if errors > 0 {
				return fmt.Errorf("%d validation error(s)", errors)
			}
			return nil
		},
	}
}

// relationGraphForValidation adapts relations for CLI validation.
type relationGraphForValidation struct {
	outEdges map[string][]struct{ id, to, relType string }
	inEdges  map[string][]struct{ id, from, relType string }
	nodeMap  map[string]bool
}

func newRelationGraphForValidation(relations []core.Relation, nodes []playbook.Node) *relationGraphForValidation {
	g := &relationGraphForValidation{
		outEdges: make(map[string][]struct{ id, to, relType string }),
		inEdges:  make(map[string][]struct{ id, from, relType string }),
		nodeMap:  make(map[string]bool),
	}
	for _, r := range relations {
		g.outEdges[r.FromDocID] = append(g.outEdges[r.FromDocID], struct{ id, to, relType string }{r.ID, r.ToDocID, r.RelationType})
		g.inEdges[r.ToDocID] = append(g.inEdges[r.ToDocID], struct{ id, from, relType string }{r.ID, r.FromDocID, r.RelationType})
	}
	for _, n := range nodes {
		g.nodeMap[n.Type+":"+n.ID] = true
	}
	return g
}

func (g *relationGraphForValidation) EdgesOut(nodeID string, edgeType string) []map[string]any {
	var result []map[string]any
	for _, e := range g.outEdges[nodeID] {
		if e.relType == edgeType {
			result = append(result, map[string]any{"id": e.id, "source_id": nodeID, "target_id": e.to, "type": e.relType})
		}
	}
	return result
}

func (g *relationGraphForValidation) EdgesIn(nodeID string, edgeType string) []map[string]any {
	var result []map[string]any
	for _, e := range g.inEdges[nodeID] {
		if e.relType == edgeType {
			result = append(result, map[string]any{"id": e.id, "source_id": e.from, "target_id": nodeID, "type": e.relType})
		}
	}
	return result
}

func (g *relationGraphForValidation) NodeExists(nodeType string, nodeID string) bool {
	return g.nodeMap[nodeType+":"+nodeID]
}

func newPlaybookImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import a playbook YAML into the database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read: %w", err)
			}
			name := strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
			name = strings.TrimSuffix(name, ".playbook")

			ver, isNew, err := a.Playbooks.Import(cmd.Context(), name, data)
			if err != nil {
				return err
			}
			if isNew {
				fmt.Fprintf(cmd.OutOrStdout(), "Imported %q version %d\n", name, ver)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "No changes. %q already at version %d\n", name, ver)
			}
			return nil
		},
	}
}
