package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	var shellFlag string

	cmd := &cobra.Command{
		Use:   "shell-init [shell]",
		Short: "Print shell integration for auto-cd switching",
		Long:  "Prints shell functions that let you switch worktrees and change directory in the current shell.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := strings.TrimSpace(shellFlag)
			if shell == "" && len(args) > 0 {
				shell = args[0]
			}
			if shell == "" {
				shell = detectShell()
			}

			snippet, err := shellInitSnippet(shell)
			if err != nil {
				return err
			}
			cmd.Print(snippet)
			return nil
		},
	}

	cmd.Flags().StringVar(&shellFlag, "shell", "", "target shell (bash, zsh, fish)")
	rootCmd.AddCommand(cmd)
}

func detectShell() string {
	s := strings.TrimSpace(os.Getenv("SHELL"))
	if s == "" {
		return ""
	}
	return filepath.Base(s)
}

func shellInitSnippet(shell string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(shell))
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSpace(s)

	switch s {
	case "bash", "zsh":
		return `# gwt shell integration
# Usage: gwt switch <worktree>
#        gwt switch            # opens interactive selector
#
# This function wraps the gwt binary:
# - intercepts "gwt switch" and cd's in the current shell
# - forwards all other commands to the real binary

gwt() {
  if [ "$1" = "switch" ]; then
    shift
    local target
    target="$(command gwt switch --print-path "$@")" || return
    cd "$target" || return
    return
  fi

  command gwt "$@"
}
`, nil
	case "fish":
		return `# gwt shell integration
# Usage: gwt switch <worktree>
#        gwt switch            # opens interactive selector
#
# This function wraps the gwt binary:
# - intercepts "gwt switch" and cd's in the current shell
# - forwards all other commands to the real binary

function gwt
    if test (count $argv) -gt 0; and test "$argv[1]" = "switch"
        set -e argv[1]
        set target (command gwt switch --print-path $argv); or return
        cd $target
        return
    end

    command gwt $argv
end
`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}
