package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/ops"
	"github.com/syvanpera/gwt/internal/ui"
)

func init() {
	var location string
	var branch string

	cmd := &cobra.Command{
		Use:   "clone <repository> [directory]",
		Short: "Clone repository using bare + worktree layout",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ops.CloneOptions{
				Repository: args[0],
				Location:   location,
				Branch:     branch,
			}
			if len(args) > 1 {
				opts.Directory = args[1]
			}
			return ui.RunOperation("gwt clone", "Clone repository and configure worktree layout", func(_ context.Context, em *ops.Emitter) error {
				return ops.Clone(em, opts)
			})
		},
	}

	cmd.Flags().StringVarP(&location, "location", "l", ".bare", "subdirectory for bare repository")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "branch to check out as initial worktree")
	rootCmd.AddCommand(cmd)
}
