package code

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitWorktreeStrategy implements CodeStrategy using git worktrees for code isolation.
// Uses go-git for repo detection and status checks, CLI git for worktree operations.
type GitWorktreeStrategy struct{}

// NewGitWorktreeStrategy returns a new GitWorktreeStrategy instance.
func NewGitWorktreeStrategy() *GitWorktreeStrategy {
	return &GitWorktreeStrategy{}
}

// Name returns the strategy name.
func (s *GitWorktreeStrategy) Name() string {
	return "worktree"
}

// Create creates a git worktree at clonePath from the repo at sourcePath.
// If branchName is empty, it is derived from the clone path basename.
// If the branch already exists, the worktree checks it out; otherwise a new branch is created.
func (s *GitWorktreeStrategy) Create(sourcePath, clonePath, branchName string) error {
	// Validate sourcePath is a git repo using go-git
	repo, err := git.PlainOpen(sourcePath)
	if err != nil {
		return fmt.Errorf("source path %q is not a git repository: %w", sourcePath, err)
	}

	// Derive branch name from clone path basename if empty
	if branchName == "" {
		branchName = filepath.Base(clonePath)
	}

	// Check if branch already exists
	branchExists := false
	refName := plumbing.NewBranchReferenceName(branchName)
	_, err = repo.Reference(refName, true)
	if err == nil {
		branchExists = true
	}

	// Create worktree via CLI git (go-git v5 does not support linked worktrees)
	var cmd *exec.Cmd
	if branchExists {
		// Check out existing branch
		cmd = exec.Command("git", "-C", sourcePath, "worktree", "add", clonePath, branchName)
	} else {
		// Create new branch
		cmd = exec.Command("git", "-C", sourcePath, "worktree", "add", clonePath, "-b", branchName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// Remove removes the git worktree at clonePath and prunes stale references.
func (s *GitWorktreeStrategy) Remove(sourcePath, clonePath string, force bool) error {
	args := []string{"-C", sourcePath, "worktree", "remove", clonePath}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// Prune stale worktree references
	pruneCmd := exec.Command("git", "-C", sourcePath, "worktree", "prune")
	if pruneOutput, err := pruneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %s: %w", strings.TrimSpace(string(pruneOutput)), err)
	}

	return nil
}

// List returns all worktree-based code copies matching the clone naming convention.
// It parses git worktree list --porcelain output and filters by <sourceProjectName>-clone-*.
func (s *GitWorktreeStrategy) List(sourcePath, sourceProjectName string) ([]CodeCopyInfo, error) {
	cmd := exec.Command("git", "-C", sourcePath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreeList(string(output), sourceProjectName), nil
}

// IsDirty checks for uncommitted changes in the worktree at clonePath.
// Uses go-git to check working tree status.
func (s *GitWorktreeStrategy) IsDirty(clonePath string) (bool, []string, error) {
	repo, err := git.PlainOpen(clonePath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open repo at %q: %w", clonePath, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return false, nil, nil
	}

	var changedFiles []string
	for file := range status {
		changedFiles = append(changedFiles, file)
	}

	return true, changedFiles, nil
}

// parseWorktreeList parses git worktree list --porcelain output.
// Each block has the format:
//
//	worktree <path>
//	HEAD <sha>
//	branch refs/heads/<name>
//
// Blocks are separated by blank lines.
// Returns only entries whose directory basename matches <sourceProjectName>-clone-*.
func parseWorktreeList(output, sourceProjectName string) []CodeCopyInfo {
	var results []CodeCopyInfo

	prefix := sourceProjectName + "-clone-"
	blocks := strings.Split(strings.TrimSpace(output), "\n\n")

	for _, block := range blocks {
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		var wtPath, branch string

		for _, line := range lines {
			if strings.HasPrefix(line, "worktree ") {
				wtPath = strings.TrimPrefix(line, "worktree ")
			}
			if strings.HasPrefix(line, "branch ") {
				ref := strings.TrimPrefix(line, "branch ")
				// Extract short name from refs/heads/...
				branch = strings.TrimPrefix(ref, "refs/heads/")
			}
		}

		if wtPath == "" {
			continue
		}

		basename := filepath.Base(wtPath)
		if !strings.HasPrefix(basename, prefix) {
			continue
		}

		cloneName := strings.TrimPrefix(basename, prefix)
		if cloneName == "" {
			continue
		}

		results = append(results, CodeCopyInfo{
			CloneName: cloneName,
			ClonePath: wtPath,
			Branch:    branch,
		})
	}

	return results
}
