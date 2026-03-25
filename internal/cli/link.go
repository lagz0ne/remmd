package cli

import (
	"fmt"
	"strings"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newLinkGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Propose, approve, and manage links between sections",
		Long: `Manage bilateral agreements between document sections.

  propose     Open a link (like opening a PR)
  approve     Approve from your side → both sides = ALIGNED
  reaffirm    After your content changes, reaffirm or withdraw
  comment     Add to the link's review thread
  list        Show links (--stale, --mine)`,
	}

	cmd.AddCommand(newLinkProposeCmd())
	cmd.AddCommand(newLinkApproveCmd())
	cmd.AddCommand(newLinkCommentCmd())
	cmd.AddCommand(newLinkReaffirmCmd())
	cmd.AddCommand(newLinkWithdrawCmd())
	cmd.AddCommand(newLinkListCmd())

	return cmd
}

func newLinkProposeCmd() *cobra.Command {
	var (
		implements string
		agreesWith string
		tests      string
		evidences  string
		rationale  string
	)
	cmd := &cobra.Command{
		Use:   "propose [left-ref]",
		Short: "Propose a link between section groups",
		Example: `  remmd link propose @a1 --implements @b1 --rationale "impl matches spec"
  remmd link propose @a1 --agrees-with @b1 --rationale "both describe same API"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			leftRef := args[0]

			var relType core.RelationshipType
			var rightRef string
			switch {
			case implements != "":
				relType, rightRef = core.RelImplements, implements
			case agreesWith != "":
				relType, rightRef = core.RelAgreesWith, agreesWith
			case tests != "":
				relType, rightRef = core.RelTests, tests
			case evidences != "":
				relType, rightRef = core.RelEvidences, evidences
			default:
				return fmt.Errorf("one of --implements, --agrees-with, --tests, or --evidences is required")
			}

			if _, err := core.ParseRef(leftRef); err != nil {
				return fmt.Errorf("invalid left ref: %w", err)
			}
			if _, err := core.ParseRef(rightRef); err != nil {
				return fmt.Errorf("invalid right ref: %w", err)
			}

			ctx := cmd.Context()

			leftSec, _, err := findSectionByRef(ctx, a.Docs, leftRef)
			if err != nil {
				return fmt.Errorf("left ref %s: %w", leftRef, err)
			}
			rightSec, _, err := findSectionByRef(ctx, a.Docs, rightRef)
			if err != nil {
				return fmt.Errorf("right ref %s: %w", rightRef, err)
			}

			link := core.NewLink(
				[]string{leftSec.ID},
				[]string{rightSec.ID},
				relType,
				core.Rationale{Claim: rationale},
				cliPrincipal,
			)

			if err := a.Links.CreateLink(ctx, link); err != nil {
				return fmt.Errorf("create link: %w", err)
			}

			if isJSON(cmd) {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"id":           link.ID,
					"state":        string(link.State),
					"left":         leftRef,
					"right":        rightRef,
					"relationship": string(link.RelationshipType),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "link %s opened: %s --%s-> %s\n",
				link.ID, leftRef, relType, rightRef)
			if rationale != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  rationale: %s\n", rationale)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&implements, "implements", "", "Right @ref (implements relationship)")
	cmd.Flags().StringVar(&agreesWith, "agrees-with", "", "Right @ref (agrees_with relationship)")
	cmd.Flags().StringVar(&tests, "tests", "", "Right @ref (tests relationship)")
	cmd.Flags().StringVar(&evidences, "evidences", "", "Right @ref (evidences relationship)")
	cmd.Flags().StringVar(&rationale, "rationale", "", "Rationale for the link")
	return cmd
}

func newLinkApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve [link-id...]",
		Short: "Approve a link from your side",
		Example: `  remmd link approve <link-id>
  remmd link approve <id1> <id2> <id3>`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			principal := core.Principal{ID: cliPrincipal, Type: core.PrincipalHuman, Name: "CLI User"}
			for _, id := range args {
				snap, err := a.Snapshots.ComputeSnapshot(ctx, id)
				if err != nil {
					return fmt.Errorf("compute snapshot for %s: %w", id, err)
				}
				if err := a.Reviews.Approve(ctx, principal, id, snap.Hash()); err != nil {
					return fmt.Errorf("approve %s: %w", id, err)
				}
				if isJSON(cmd) {
					if err := writeJSON(cmd.OutOrStdout(), map[string]string{"id": id, "state": "aligned"}); err != nil {
						return err
					}
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "approved %s\n", id)
			}
			return nil
		},
	}
}

func newLinkCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "comment [link-id] [text]",
		Short:   "Comment on a link's thread",
		Example: `  remmd link comment <link-id> "Looks good, but check edge cases"`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			linkID := args[0]
			text := args[1]

			if err := a.Links.AddThreadEntry(cmd.Context(), linkID, core.EntryComment, cliPrincipal, text); err != nil {
				return fmt.Errorf("comment on %s: %w", linkID, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "commented on %s\n", linkID)
			return nil
		},
	}
}

func newLinkReaffirmCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "reaffirm [link-id]",
		Short: "Reaffirm a link after your content change",
		Example: `  remmd link reaffirm <link-id>
  remmd link reaffirm --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			principal := core.Principal{ID: cliPrincipal, Type: core.PrincipalHuman, Name: "CLI User"}

			if all {
				links, err := a.Links.ListLinks(ctx, string(core.LinkStale))
				if err != nil {
					return fmt.Errorf("list stale links: %w", err)
				}
				for _, link := range links {
					if err := a.Reviews.Reaffirm(ctx, principal, link.ID); err != nil {
						return fmt.Errorf("reaffirm %s: %w", link.ID, err)
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "reaffirmed %d stale links\n", len(links))
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("link-id required (or use --all)")
			}

			id := args[0]
			if err := a.Reviews.Reaffirm(ctx, principal, id); err != nil {
				return fmt.Errorf("reaffirm %s: %w", id, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "reaffirmed %s\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Reaffirm all stale links")
	return cmd
}

func newLinkWithdrawCmd() *cobra.Command {
	var reason string
	cmd := &cobra.Command{
		Use:   "withdraw [link-id]",
		Short: "Withdraw (archive) a link",
		Example: `  remmd link withdraw <link-id> --reason "No longer needed"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			principal := core.Principal{ID: cliPrincipal, Type: core.PrincipalHuman, Name: "CLI User"}

			id := args[0]
			if reason == "" {
				reason = "withdrawn"
			}
			if err := a.Reviews.Withdraw(ctx, principal, id, reason); err != nil {
				return fmt.Errorf("withdraw %s: %w", id, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "withdrawn %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for withdrawal")
	return cmd
}

func newLinkListCmd() *cobra.Command {
	var (
		stale bool
		mine  bool
	)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List verification agreements",
		Example: `  remmd link list
  remmd link list --stale
  remmd link list --mine`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			var stateFilter string
			if stale {
				stateFilter = string(core.LinkStale)
			}
			// TODO: filter by proposer_id == "cli-user" once we have auth
			if mine {
				stateFilter = string(core.LinkPending)
			}

			links, err := a.Links.ListLinks(cmd.Context(), stateFilter)
			if err != nil {
				return fmt.Errorf("list links: %w", err)
			}

			if isJSON(cmd) {
				type linkEntry struct {
					ID           string `json:"id"`
					State        string `json:"state"`
					Relationship string `json:"relationship"`
					Left         string `json:"left"`
					Right        string `json:"right"`
				}
				entries := make([]linkEntry, 0, len(links))
				for _, l := range links {
					entries = append(entries, linkEntry{
						ID:           l.ID,
						State:        string(l.State),
						Relationship: string(l.RelationshipType),
						Left:         strings.Join(l.LeftSectionIDs, ","),
						Right:        strings.Join(l.RightSectionIDs, ","),
					})
				}
				return writeJSON(cmd.OutOrStdout(), entries)
			}

			if len(links) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no links)")
				return nil
			}

			for _, l := range links {
				leftIDs := strings.Join(l.LeftSectionIDs, ",")
				rightIDs := strings.Join(l.RightSectionIDs, ",")
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s --%s-> %s  [%s]\n",
					l.ID, leftIDs, l.RelationshipType, rightIDs, l.State)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&stale, "stale", false, "Show only stale links")
	cmd.Flags().BoolVar(&mine, "mine", false, "Show only links waiting on me")
	return cmd
}
