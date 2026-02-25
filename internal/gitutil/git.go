package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(dir string, args ...string) error {
	_, err := Output(dir, args...)
	return err
}

func Output(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out.String()), nil
}

func RepoRoot(dir string) (string, error) {
	out, err := Output(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Clean(out), nil
}

func ResolveGWTRepoRoot(dir string) (string, error) {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	if root, err := RepoRoot(abs); err == nil {
		return root, nil
	}

	if isGWTLayout(abs) {
		return abs, nil
	}

	gitDir, err := Output(abs, "rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(abs, gitDir)
	}
	gitDir = filepath.Clean(gitDir)
	marker := string(filepath.Separator) + ".bare" + string(filepath.Separator) + "worktrees" + string(filepath.Separator)
	if idx := strings.Index(gitDir, marker); idx != -1 {
		return gitDir[:idx], nil
	}
	suffix := string(filepath.Separator) + ".bare"
	if strings.HasSuffix(gitDir, suffix) {
		return strings.TrimSuffix(gitDir, suffix), nil
	}
	return "", errors.New("could not resolve gwt repo root")
}

func isGWTLayout(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	barePath := filepath.Join(dir, ".bare")
	if fi, err := os.Stat(gitPath); err != nil || fi.IsDir() {
		return false
	}
	if fi, err := os.Stat(barePath); err != nil || !fi.IsDir() {
		return false
	}
	return true
}
