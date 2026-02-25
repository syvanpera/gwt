package ops

import "testing"

func TestDefaultDirName(t *testing.T) {
	tests := map[string]string{
		"git@github.com:example/repo.git":     "repo",
		"https://github.com/example/repo.git": "repo",
		"https://github.com/example/repo":     "repo",
	}
	for in, want := range tests {
		got := defaultDirName(in)
		if got != want {
			t.Fatalf("defaultDirName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseWorktreePorcelain(t *testing.T) {
	in := `worktree /tmp/repo
HEAD 1234567890abcdef
branch refs/heads/main

worktree /tmp/repo/feature/a
HEAD fedcba0987654321
branch refs/heads/feature/a
`
	got := parseWorktreePorcelain(in)
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[1].Path != "/tmp/repo/feature/a" {
		t.Fatalf("unexpected path: %s", got[1].Path)
	}
	if got[1].Branch != "refs/heads/feature/a" {
		t.Fatalf("unexpected branch: %s", got[1].Branch)
	}
}
