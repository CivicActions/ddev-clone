package clone

import (
	"testing"
)

func TestGetSourceProjectName_Simple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple project name",
			input:    "myproject",
			expected: "myproject",
		},
		{
			name:     "clone name returns source",
			input:    "myproject-clone-feature",
			expected: "myproject",
		},
		{
			name:     "nested clone returns outermost source",
			input:    "myproject-clone-feat1-clone-feat2",
			expected: "myproject-clone-feat1",
		},
		{
			name:     "no clone suffix",
			input:    "some-other-project",
			expected: "some-other-project",
		},
		{
			name:     "clone at beginning",
			input:    "clone-test",
			expected: "clone-test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "hyphenated project with clone",
			input:    "my-cool-project-clone-bugfix",
			expected: "my-cool-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSourceProjectName(tt.input)
			if result != tt.expected {
				t.Errorf("GetSourceProjectName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCloneProjectName(t *testing.T) {
	tests := []struct {
		name       string
		sourceName string
		branchName string
		expected   string
	}{
		{
			name:       "simple branch",
			sourceName: "myproject",
			branchName: "feature-x",
			expected:   "myproject-clone-feature-x",
		},
		{
			name:       "branch with slashes preserved",
			sourceName: "myproject",
			branchName: "feature/new-thing",
			expected:   "myproject-clone-feature/new-thing",
		},
		{
			name:       "empty branch",
			sourceName: "myproject",
			branchName: "",
			expected:   "myproject-clone-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCloneProjectName(tt.sourceName, tt.branchName)
			if result != tt.expected {
				t.Errorf("getCloneProjectName(%q, %q) = %q, want %q",
					tt.sourceName, tt.branchName, result, tt.expected)
			}
		})
	}
}

func TestGetBranchFromPath_NonGitDirReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	result := getBranchFromPath(dir)
	if result != "" {
		t.Errorf("expected empty string for non-git dir, got %q", result)
	}
}

func TestGetBranchFromPath_NonExistentDirReturnsEmpty(t *testing.T) {
	result := getBranchFromPath("/nonexistent/path/xyz")
	if result != "" {
		t.Errorf("expected empty string for nonexistent dir, got %q", result)
	}
}
