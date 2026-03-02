package code

import (
	"fmt"
	"os"

	cp "github.com/otiai10/copy"

	"github.com/civicactions/ddev-clone/pkg/ddev"
	"github.com/go-git/go-git/v5"
)

// DirectoryCopyStrategy implements CodeStrategy using full directory copies.
type DirectoryCopyStrategy struct{}

// NewDirectoryCopyStrategy returns a new DirectoryCopyStrategy instance.
func NewDirectoryCopyStrategy() *DirectoryCopyStrategy {
	return &DirectoryCopyStrategy{}
}

// Name returns the strategy name.
func (s *DirectoryCopyStrategy) Name() string {
	return "copy"
}

// Create performs a recursive directory copy from sourcePath to clonePath
// preserving timestamps, ownership (best effort), and symlinks.
// branchName is ignored for copy strategy.
func (s *DirectoryCopyStrategy) Create(sourcePath, clonePath, _ string) error {
	opts := cp.Options{
		PreserveTimes: true,
		PreserveOwner: true, // gracefully degrades without root
		OnSymlink: func(_ string) cp.SymlinkAction {
			return cp.Shallow // preserve symlinks as-is
		},
	}

	if err := cp.Copy(sourcePath, clonePath, opts); err != nil {
		return fmt.Errorf("failed to copy directory: %w", err)
	}

	return nil
}

// Remove removes the clone directory.
func (s *DirectoryCopyStrategy) Remove(_, clonePath string, _ bool) error {
	return os.RemoveAll(clonePath)
}

// List returns all copy-based clones by scanning the DDEV project list.
func (s *DirectoryCopyStrategy) List(_, sourceProjectName string) ([]CodeCopyInfo, error) {
	projects, err := ddev.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list DDEV projects: %w", err)
	}

	var results []CodeCopyInfo
	prefix := sourceProjectName + "-clone-"
	for _, proj := range projects {
		if len(proj.Name) <= len(prefix) {
			continue
		}
		if proj.Name[:len(prefix)] != prefix {
			continue
		}
		cloneName := proj.Name[len(prefix):]

		branch := ""
		if repo, err := git.PlainOpen(proj.AppRoot); err == nil {
			if head, err := repo.Head(); err == nil {
				branch = head.Name().Short()
			}
		}

		results = append(results, CodeCopyInfo{
			CloneName: cloneName,
			ClonePath: proj.AppRoot,
			Branch:    branch,
		})
	}

	return results, nil
}

// IsDirty checks for uncommitted changes in the clone directory.
// Returns false if the directory is not a git repo.
func (s *DirectoryCopyStrategy) IsDirty(clonePath string) (bool, []string, error) {
	repo, err := git.PlainOpen(clonePath)
	if err != nil {
		// Not a git repo → not dirty
		return false, nil, nil
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
