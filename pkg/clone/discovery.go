package clone

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/civicactions/ddev-clone/pkg/code"
	"github.com/civicactions/ddev-clone/pkg/ddev"
)

const cloneInfix = "-clone-"

// DiscoverClones finds all clones of the given source project by scanning
// the DDEV project list for names matching <sourceProjectName>-clone-*.
func DiscoverClones(sourceProjectName, sourcePath string) ([]CloneInfo, error) {
	projects, err := ddev.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list DDEV projects: %w", err)
	}

	prefix := sourceProjectName + cloneInfix

	// Build a set of worktree paths for strategy detection
	worktreeStrategy := code.NewGitWorktreeStrategy()
	worktreeInfos, _ := worktreeStrategy.List(sourcePath, sourceProjectName)
	worktreePathSet := make(map[string]code.CodeCopyInfo)
	for _, info := range worktreeInfos {
		worktreePathSet[info.ClonePath] = info
	}

	cwd, _ := os.Getwd()

	var clones []CloneInfo
	for _, proj := range projects {
		if !strings.HasPrefix(proj.Name, prefix) {
			continue
		}

		cloneName := strings.TrimPrefix(proj.Name, prefix)
		if cloneName == "" {
			continue
		}

		info := CloneInfo{
			CloneName:         cloneName,
			SourceProjectName: sourceProjectName,
			ProjectName:       proj.Name,
			ClonePath:         proj.AppRoot,
			Status:            proj.Status,
		}

		// Determine strategy type and branch
		if wtInfo, ok := worktreePathSet[proj.AppRoot]; ok {
			info.CodeStrategyType = "worktree"
			info.Branch = wtInfo.Branch
		} else {
			info.CodeStrategyType = "copy"
			// Try to get branch info from go-git
			info.Branch = getBranchFromPath(proj.AppRoot)
		}

		// Get detailed status if available
		if desc, err := ddev.Describe(proj.Name); err == nil {
			info.Status = desc.Status
		}

		// Check if current directory is within this clone
		if cwd != "" && proj.AppRoot != "" {
			absClone, err1 := filepath.Abs(proj.AppRoot)
			absCwd, err2 := filepath.Abs(cwd)
			if err1 == nil && err2 == nil {
				if absCwd == absClone || strings.HasPrefix(absCwd, absClone+string(filepath.Separator)) {
					info.Current = true
				}
			}
		}

		clones = append(clones, info)
	}

	// Sort by clone name
	sort.Slice(clones, func(i, j int) bool {
		return clones[i].CloneName < clones[j].CloneName
	})

	return clones, nil
}

// GetSourceProjectName determines the source project name from a potentially
// clone project name. If projectName contains "-clone-", everything before
// the last occurrence is the source name. Otherwise, projectName IS the source.
func GetSourceProjectName(projectName string) string {
	idx := strings.LastIndex(projectName, cloneInfix)
	if idx < 0 {
		return projectName
	}
	return projectName[:idx]
}

// getCloneProjectName returns the DDEV project name for a clone.
func getCloneProjectName(sourceProjectName, cloneName string) string {
	return sourceProjectName + cloneInfix + cloneName
}

// getBranchFromPath tries to detect the current branch of a git repo at the given path.
// Returns empty string if not a git repo or on error.
func getBranchFromPath(repoPath string) string {
	repo, err := openRepo(repoPath)
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	if err != nil {
		return ""
	}
	return head.Name().Short()
}
