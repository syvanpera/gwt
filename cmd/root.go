package cmd

import (
	"fmt"
	"os"
	"time"

	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/ui"
)

var uiStepDelay time.Duration

var rootCmd = &cobra.Command{
	Use:           "gwt",
	Short:         "Git worktree management tool",
	Long:          "gwt - Git worktree manager using a simple, status-driven terminal UI.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ui.SetEventDelay(uiStepDelay)
	},
}

func Execute() {
	cc.Init(&cc.Config{
		RootCmd:         rootCmd,
		Headings:        cc.HiGreen + cc.Bold + cc.Underline,
		Commands:        cc.HiYellow + cc.Bold,
		Example:         cc.Italic,
		ExecName:        cc.Bold,
		Flags:           cc.Bold,
		NoExtraNewlines: true,
	})

	if err := rootCmd.Execute(); err != nil {
		if !ui.IsDisplayedError(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().DurationVar(&uiStepDelay, "ui-step-delay", 0, "artificial delay between UI status updates (e.g. 300ms, 1s)")
}
