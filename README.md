# gwt

A git worktree manager. `gwt` makes it ergonomic to work exclusively from
git worktrees — one directory per branch — instead of switching branches
in a single checkout.

## How it works

`gwt clone` sets up a repository using the **bare repo + `.git` pointer** pattern:

```
my-repo/
├── .bare/        ← bare clone (all git data lives here)
├── .git          ← pointer file: "gitdir: ./.bare"
├── main/         ← worktree for the main branch
└── my-feature/   ← worktree for a feature branch
```

Each subdirectory is an independent working tree. You can have multiple branches
open simultaneously without stashing or switching.

## Installation

```sh
go install github.com/syvanpera/gwt@latest
```

Or build from source:

```sh
git clone https://github.com/syvanpera/gwt
cd gwt
go build -o gwt .
```

## Commands

### `gwt clone <repository> [directory]`

Clone a repository as a bare repo and configure it for worktree-based
development.

```sh
gwt clone git@github.com:example/my-repo.git
gwt clone git@github.com:example/my-repo.git my-dir
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `-l`, `--location` | `.bare` | Subdirectory for bare repo contents |
| `-b`, `--branch` | — | Branch to check out as a worktree immediately after cloning |

**Example**

```sh
# Clone and immediately check out the main branch as a worktree
gwt clone git@github.com:example/my-repo.git --branch main
cd my-repo/main
```

---

### `gwt checkout <branch> [directory]` / `gwt co <branch> [directory]`

Create a new worktree for a branch. If the branch does not exist locally it is
created from `--base`. The worktree directory defaults to the branch name.

After creating the worktree, `gwt` sets up remote tracking: if the branch
already exists on `origin` it fetches and sets the upstream; if not, it pushes
the new branch and sets the upstream.

```sh
gwt checkout my-feature
gwt checkout my-feature my-feature-dir
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `-B`, `--base` | `origin/main` | Base branch to create the new branch from |
| `-N`, `--no-upstream` | `false` | Skip creating or setting the upstream remote branch |

**Example**

```sh
# Create a new branch from origin/main and open it as a worktree
gwt checkout feature/auth
cd feature/auth

# Check out an existing branch as a worktree, based off a different branch
gwt checkout hotfix/login --base origin/release-2.0
```

---

### `gwt list` / `gwt ls`

List worktrees for the current repository (bare repo entry is hidden by default).

```sh
gwt list
gwt ls
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `--all` | `false` | Include the bare repository entry in output |

---

### `gwt remove <worktree>` / `gwt rm <worktree>`

Remove a worktree and delete its local branch.

```sh
gwt remove my-feature
gwt rm my-feature
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `-f`, `--force` | `false` | Force removal even if the worktree has uncommitted changes or the branch is unmerged |
| `--keep-branch` | `false` | Remove the worktree but leave the local branch intact |

**Example**

```sh
# Remove a merged feature branch and its worktree
gwt remove feature/auth

# Remove a dirty worktree without caring about uncommitted changes
gwt remove feature/auth --force

# Remove the worktree but keep the branch around
gwt remove feature/auth --keep-branch
```

---

## UI behavior

Mutating commands (`clone`, `checkout`, `remove`) run in a Bubble Tea UI with
consistent operation states:

- `RUNNING` (blue): operation is in progress
- `OK` (green): operation finished successfully
- `FAILED` (red): operation failed

For long-running tasks, `gwt` shows:

- a spinner + animated `Working...` text while running
- a short operation history and elapsed time

On failure, error details are shown once in the status panel.

`list`/`ls` uses a styled panel output with colored rows.

Use `--ui-step-delay` to slow down operation UI transitions for review:

```sh
gwt --ui-step-delay 400ms clone <repo>
gwt --ui-step-delay 1s checkout feature/demo
```

---

## Shell completion

`gwt` provides completion scripts for bash, zsh, fish, and PowerShell via Cobra:

```sh
# zsh
gwt completion zsh > "${fpath[1]}/_gwt"

# bash
gwt completion bash > /etc/bash_completion.d/gwt

# fish
gwt completion fish > ~/.config/fish/completions/gwt.fish
```

## Built with

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — UI components (spinner)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling
