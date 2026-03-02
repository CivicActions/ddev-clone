package integration

// Integration tests for ddev-clone.
//
// These tests require a running Docker daemon and DDEV installed.
// They are intended to be run manually or in CI environments with
// Docker available.
//
// Run with: go test -v -tags=integration ./test/integration/
//
// These tests are skipped by default in short mode:
//   go test -short ./...

import (
	"os"
	"testing"
)

func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	// Check if Docker socket is available
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		t.Skip("skipping: Docker not available")
	}
}

func skipIfNoDDEV(t *testing.T) {
	t.Helper()
	// Check if ddev binary is in PATH
	path := os.Getenv("PATH")
	if path == "" {
		t.Skip("skipping: PATH not set")
	}
}

// TestCreateWorktree tests the full create flow with the git worktree strategy.
// Requires: Docker, DDEV, a configured DDEV project with a git repo
func TestCreateWorktree(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for worktree clone creation")
}

// TestCreateCopy tests the full create flow with the directory copy strategy.
// Requires: Docker, DDEV, a configured DDEV project
func TestCreateCopy(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for copy clone creation")
}

// TestList tests listing clones of a source project.
// Requires: Docker, DDEV, existing clones
func TestList(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for clone listing")
}

// TestRemove tests removing a clone with cleanup.
// Requires: Docker, DDEV, an existing clone
func TestRemove(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for clone removal")
}

// TestPrune tests pruning stale clones.
// Requires: Docker, DDEV, stale clone entries
func TestPrune(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for clone pruning")
}

// TestVolumeClone tests Docker volume cloning with tar pipe.
// Requires: Docker daemon
func TestVolumeClone(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	t.Skip("TODO: implement full integration test for volume cloning")
}

// TestAutoFallbackToCopy tests that a non-git project automatically falls back
// to the directory copy strategy.
// Requires: Docker, DDEV, a configured DDEV project without git
func TestAutoFallbackToCopy(t *testing.T) {
	skipIfShort(t)
	skipIfNoDocker(t)
	skipIfNoDDEV(t)
	t.Skip("TODO: implement full integration test for auto-fallback to copy strategy")
}
