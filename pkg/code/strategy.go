package code

// CodeCopyInfo holds information about a single code copy discovered by a strategy.
type CodeCopyInfo struct {
	// CloneName is the user-provided clone name (e.g., "feature-x").
	CloneName string
	// ClonePath is the absolute filesystem path to the clone directory.
	ClonePath string
	// Branch is the current git branch (empty if not applicable).
	Branch string
}

// CodeStrategy abstracts code isolation methods for creating clone environments.
// Each implementation handles creating, removing, listing, and inspecting code copies.
type CodeStrategy interface {
	// Create creates a code copy from sourcePath to clonePath.
	// branchName is used by git-based strategies; ignored by copy strategies.
	Create(sourcePath, clonePath, branchName string) error

	// Remove removes the code copy at clonePath.
	// sourcePath is needed by worktree strategy for git operations.
	// force skips safety checks (e.g., uncommitted changes).
	Remove(sourcePath, clonePath string, force bool) error

	// List returns all code copies matching the clone naming convention.
	List(sourcePath, sourceProjectName string) ([]CodeCopyInfo, error)

	// IsDirty checks for uncommitted changes in the code copy.
	// Returns the dirty flag and a list of changed file paths.
	IsDirty(clonePath string) (bool, []string, error)

	// Name returns the strategy identifier ("worktree" or "copy").
	Name() string
}
