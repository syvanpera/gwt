package ops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/syvanpera/gwt/internal/gitutil"
)

type CloneOptions struct {
	Repository string
	Directory  string
	Location   string
	Branch     string
}

type CheckoutOptions struct {
	Branch     string
	Directory  string
	Base       string
	NoUpstream bool
	RootDir    string
}

type RemoveOptions struct {
	Worktree   string
	Force      bool
	KeepBranch bool
	RootDir    string
}

type Worktree struct {
	Path   string
	Branch string
	Head   string
	Bare   bool
}

func Clone(em *Emitter, opts CloneOptions) error {
	if opts.Repository == "" {
		return errors.New("repository is required")
	}
	target := opts.Directory
	if target == "" {
		target = defaultDirName(opts.Repository)
	}
	if opts.Location == "" {
		opts.Location = ".bare"
	}
	em.Started("clone", "Preparing clone")
	em.Progress("Validating target directory", 0.05)
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("target directory already exists: %s", target)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	bareLocation := filepath.Join(target, opts.Location)
	em.ProgressIndeterminate("Cloning bare repository")
	if err := gitutil.Run("", "clone", "--bare", opts.Repository, bareLocation); err != nil {
		return err
	}

	em.Progress("Configuring .git pointer", 0.55)
	gitPointer := fmt.Sprintf("gitdir: ./%s\n", strings.TrimPrefix(filepath.ToSlash(opts.Location), "./"))
	if err := os.WriteFile(filepath.Join(target, ".git"), []byte(gitPointer), 0o644); err != nil {
		return fmt.Errorf("write .git pointer: %w", err)
	}

	if opts.Branch != "" {
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("resolve target path: %w", err)
		}
		worktreeDir := filepath.Join(absTarget, opts.Branch)
		em.ProgressIndeterminate("Creating initial worktree")
		exists := gitutil.Run(target, "show-ref", "--verify", "--quiet", "refs/heads/"+opts.Branch) == nil
		if exists {
			if err := gitutil.Run(target, "worktree", "add", worktreeDir, opts.Branch); err != nil {
				return err
			}
		} else {
			if err := gitutil.Run(target, "worktree", "add", "-b", opts.Branch, worktreeDir, "origin/"+opts.Branch); err != nil {
				return err
			}
		}
	}

	em.Completed("Repository ready")
	return nil
}

func Checkout(em *Emitter, opts CheckoutOptions) error {
	if opts.Branch == "" {
		return errors.New("branch is required")
	}
	if opts.Base == "" {
		opts.Base = "origin/main"
	}
	root := opts.RootDir
	if root == "" {
		var err error
		root, err = gitutil.ResolveGWTRepoRoot(".")
		if err != nil {
			return fmt.Errorf("resolve repo root: %w", err)
		}
	}
	worktreeDir := opts.Directory
	if worktreeDir == "" {
		worktreeDir = opts.Branch
	}
	if !filepath.IsAbs(worktreeDir) {
		worktreeDir = filepath.Join(root, worktreeDir)
	}

	em.Started("checkout", "Preparing worktree checkout")
	em.ProgressIndeterminate("Fetching origin")
	_ = gitutil.Run(root, "fetch", "origin")
	baseRef := resolveBaseRef(root, opts.Base)

	em.Progress("Resolving branch", 0.2)
	exists := gitutil.Run(root, "show-ref", "--verify", "--quiet", "refs/heads/"+opts.Branch) == nil

	em.ProgressIndeterminate("Creating worktree")
	if exists {
		if err := gitutil.Run(root, "worktree", "add", worktreeDir, opts.Branch); err != nil {
			return err
		}
	} else {
		if err := gitutil.Run(root, "worktree", "add", "-b", opts.Branch, worktreeDir, baseRef); err != nil {
			return err
		}
	}

	if !opts.NoUpstream {
		em.Progress("Configuring upstream", 0.8)
		remoteRef := "refs/remotes/origin/" + opts.Branch
		hasRemote := gitutil.Run(root, "show-ref", "--verify", "--quiet", remoteRef) == nil
		if hasRemote {
			if err := gitutil.Run(worktreeDir, "branch", "--set-upstream-to", "origin/"+opts.Branch, opts.Branch); err != nil {
				return err
			}
		} else {
			if err := gitutil.Run(worktreeDir, "push", "-u", "origin", opts.Branch); err != nil {
				return err
			}
		}
	}

	em.Completed("Worktree created")
	return nil
}

func ListWorktrees(root string) ([]Worktree, error) {
	if root == "" {
		var err error
		root, err = gitutil.ResolveGWTRepoRoot(".")
		if err != nil {
			return nil, fmt.Errorf("resolve repo root: %w", err)
		}
	}
	out, err := gitutil.Output(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreePorcelain(out), nil
}

func Remove(em *Emitter, opts RemoveOptions) error {
	if opts.Worktree == "" {
		return errors.New("worktree is required")
	}
	root := opts.RootDir
	if root == "" {
		var err error
		root, err = gitutil.ResolveGWTRepoRoot(".")
		if err != nil {
			return fmt.Errorf("resolve repo root: %w", err)
		}
	}

	em.Started("remove", "Resolving worktree")
	list, err := ListWorktrees(root)
	if err != nil {
		return err
	}

	var selected *Worktree
	relativeTarget := opts.Worktree
	if !filepath.IsAbs(relativeTarget) {
		relativeTarget = filepath.Clean(filepath.Join(root, opts.Worktree))
	}
	for i := range list {
		wt := list[i]
		if wt.Path == opts.Worktree || wt.Path == relativeTarget || filepath.Base(wt.Path) == opts.Worktree {
			selected = &wt
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("worktree not found: %s", opts.Worktree)
	}

	args := []string{"worktree", "remove"}
	if opts.Force {
		args = append(args, "--force")
	}
	args = append(args, selected.Path)
	em.ProgressIndeterminate("Removing worktree")
	if err := gitutil.Run(root, args...); err != nil {
		return err
	}

	if !opts.KeepBranch && selected.Branch != "" {
		branch := strings.TrimPrefix(selected.Branch, "refs/heads/")
		em.Progress("Deleting local branch", 0.8)
		deleteArgs := []string{"branch", "-d", branch}
		if opts.Force {
			deleteArgs = []string{"branch", "-D", branch}
		}
		if err := gitutil.Run(root, deleteArgs...); err != nil {
			return err
		}
	}

	em.Completed("Worktree removed")
	return nil
}

func defaultDirName(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	repo = strings.TrimSuffix(repo, "/")
	parts := strings.Split(repo, "/")
	if len(parts) == 0 {
		return "repo"
	}
	name := parts[len(parts)-1]
	if name == "" {
		return "repo"
	}
	return name
}

func parseWorktreePorcelain(in string) []Worktree {
	if strings.TrimSpace(in) == "" {
		return nil
	}
	lines := strings.Split(in, "\n")
	res := make([]Worktree, 0)
	var current *Worktree

	flush := func() {
		if current != nil {
			res = append(res, *current)
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			current = nil
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				flush()
			}
			current = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
			continue
		}
		if current == nil {
			continue
		}
		switch {
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch ")
		case line == "bare":
			current.Bare = true
		}
	}
	flush()
	return res
}

func resolveBaseRef(root, base string) string {
	candidates := []string{base}
	if strings.HasPrefix(base, "origin/") {
		candidates = append(candidates, strings.TrimPrefix(base, "origin/"))
	}
	candidates = append(candidates, "main", "master")
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if gitutil.Run(root, "rev-parse", "--verify", "--quiet", c+"^{commit}") == nil {
			return c
		}
	}
	return base
}
