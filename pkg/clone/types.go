package clone

import "github.com/civicactions/ddev-clone/pkg/volume"

// CloneInfo represents metadata about a single DDEV-managed clone.
type CloneInfo struct {
	CloneName         string `json:"clone_name"`
	SourceProjectName string `json:"source_project_name"`
	ProjectName       string `json:"project_name"`
	ClonePath         string `json:"path"`
	Branch            string `json:"branch"`
	Status            string `json:"status"`
	CodeStrategyType  string `json:"strategy"`
	Current           bool   `json:"current"`
}

// CreateOptions holds parameters for clone creation.
type CreateOptions struct {
	SourceProjectName string
	SourcePath        string
	CloneName         string
	Branch            string
	CodeStrategyName  string // "worktree", "copy", or "" (auto-detect)
	NoStart           bool
	VolumeCloner      volume.VolumeCloner
}

// RemoveOptions holds parameters for clone removal.
type RemoveOptions struct {
	SourceProjectName string
	SourcePath        string
	CloneName         string
	Force             bool
	// ConfirmFunc is called when dirty changes are detected and Force is false.
	// It should return true to proceed with removal.
	ConfirmFunc func(cloneName string, changedFiles []string) bool
}

// PruneOptions holds parameters for stale clone cleanup.
type PruneOptions struct {
	SourceProjectName string
	SourcePath        string
	DryRun            bool
}

// VolumePair maps a source Docker volume name to its clone counterpart.
type VolumePair struct {
	Source string
	Target string
}
