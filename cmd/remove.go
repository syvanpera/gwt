package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/ops"
	"github.com/syvanpera/gwt/internal/ui"
)

func init() {
	var force bool
	var keepBranch bool

	cmd := &cobra.Command{
		Use:     "remove <worktree>",
		Aliases: []string{"rm"},
		Short:   "Remove a worktree and optionally its branch",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ops.RemoveOptions{Worktree: args[0], Force: force, KeepBranch: keepBranch}
			return ui.RunOperation("gwt remove", "Remove worktree and cleanup branch", func(_ context.Context, em *ops.Emitter) error {
				return ops.Remove(em, opts)
			})
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force removal even with local changes")
	cmd.Flags().BoolVar(&keepBranch, "keep-branch", false, "remove worktree but keep local branch")
	rootCmd.AddCommand(cmd)
}
