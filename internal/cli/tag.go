package cli

import (
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tag subscriptions",
		Long:  "Subscribe sections to tags, list subscriptions, and dismiss notifications.",
	}

	cmd.AddCommand(newTagSubscribeCmd())
	cmd.AddCommand(newTagUnsubscribeCmd())
	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagDismissCmd())

	return cmd
}

func newTagSubscribeCmd() *cobra.Command {
	var tag string
	cmd := &cobra.Command{
		Use:     "subscribe [ref]",
		Short:   "Subscribe a section to a tag",
		Example: `  remmd tag subscribe @a1 --tag payment`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			ref := args[0]
			if tag == "" {
				return fmt.Errorf("--tag is required")
			}

			sub := core.NewSubscription(ref, tag, cliPrincipal)
			if err := a.Subscriptions.CreateSubscription(cmd.Context(), sub); err != nil {
				return fmt.Errorf("subscribe: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "subscribed %s to tag %q (id: %s)\n", ref, tag, sub.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "Tag to subscribe to")
	return cmd
}

func newTagUnsubscribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "unsubscribe [subscription-id]",
		Short:   "Archive a subscription",
		Example: `  remmd tag unsubscribe <subscription-id>`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			id := args[0]
			if err := a.Subscriptions.ArchiveSubscription(cmd.Context(), id); err != nil {
				return fmt.Errorf("unsubscribe: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "unsubscribed %s\n", id)
			return nil
		},
	}
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active subscriptions",
		Example: `  remmd tag list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			subs, err := a.Subscriptions.ListActiveSubscriptions(cmd.Context())
			if err != nil {
				return fmt.Errorf("list subscriptions: %w", err)
			}

			if len(subs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no active subscriptions)")
				return nil
			}

			for _, s := range subs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s -> tag %q (by %s)\n",
					s.ID, s.SubscriberRef, s.TagExpression, s.CreatedBy)
			}
			return nil
		},
	}
}

func newTagDismissCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "dismiss [fire-id]",
		Short:   "Dismiss a subscription notification",
		Example: `  remmd tag dismiss <fire-id>`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := RequireApp(cmd)
			if err != nil {
				return err
			}

			id := args[0]
			if err := a.Subscriptions.DismissFire(cmd.Context(), id); err != nil {
				return fmt.Errorf("dismiss: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "dismissed %s\n", id)
			return nil
		},
	}
}
