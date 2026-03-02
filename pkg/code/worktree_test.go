package code

import (
	"testing"
)

func TestParseWorktreeList_Basic(t *testing.T) {
	output := "worktree /home/user/mysite\nHEAD abc123\nbranch refs/heads/main\n\nworktree /home/user/mysite-clone-feature-x\nHEAD def456\nbranch refs/heads/feature-x\n\nworktree /home/user/mysite-clone-bugfix\nHEAD ghi789\nbranch refs/heads/bugfix\n"

	results := parseWorktreeList(output, "mysite")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].CloneName != "feature-x" {
		t.Errorf("expected clone name 'feature-x', got %q", results[0].CloneName)
	}
	if results[0].Branch != "feature-x" {
		t.Errorf("expected branch 'feature-x', got %q", results[0].Branch)
	}
	if results[0].ClonePath != "/home/user/mysite-clone-feature-x" {
		t.Errorf("unexpected clone path: %s", results[0].ClonePath)
	}
	if results[1].CloneName != "bugfix" {
		t.Errorf("expected clone name 'bugfix', got %q", results[1].CloneName)
	}
}

func TestParseWorktreeList_NoClones(t *testing.T) {
	output := "worktree /home/user/mysite\nHEAD abc123\nbranch refs/heads/main\n"

	results := parseWorktreeList(output, "mysite")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseWorktreeList_EmptyOutput(t *testing.T) {
	results := parseWorktreeList("", "mysite")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseWorktreeList_DetachedHead(t *testing.T) {
	// Detached HEAD entries don't have a branch line
	output := "worktree /home/user/mysite\nHEAD abc123\nbranch refs/heads/main\n\nworktree /home/user/mysite-clone-detached\nHEAD def456\ndetached\n"

	results := parseWorktreeList(output, "mysite")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].CloneName != "detached" {
		t.Errorf("expected clone name 'detached', got %q", results[0].CloneName)
	}
	if results[0].Branch != "" {
		t.Errorf("expected empty branch for detached HEAD, got %q", results[0].Branch)
	}
}

func TestParseWorktreeList_SkipsNonMatchingPaths(t *testing.T) {
	output := "worktree /home/user/mysite\nHEAD abc123\nbranch refs/heads/main\n\nworktree /home/user/otherproject-clone-something\nHEAD def456\nbranch refs/heads/something\n"

	results := parseWorktreeList(output, "mysite")
	if len(results) != 0 {
		t.Fatalf("expected 0 results (other project's clone), got %d", len(results))
	}
}

func TestGitWorktreeStrategy_Name(t *testing.T) {
	s := NewGitWorktreeStrategy()
	if s.Name() != "worktree" {
		t.Errorf("expected 'worktree', got %q", s.Name())
	}
}
