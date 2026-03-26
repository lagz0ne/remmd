package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
