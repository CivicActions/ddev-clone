package code

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryCopyStrategy_Name(t *testing.T) {
	s := NewDirectoryCopyStrategy()
	if s.Name() != "copy" {
		t.Errorf("expected 'copy', got %q", s.Name())
	}
}

func TestDirectoryCopyStrategy_CreateAndRemove(t *testing.T) {
	// Create a temp source directory with some files
	srcDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(srcDir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	// Create a symlink
	if err := os.Symlink("hello.txt", filepath.Join(srcDir, "link.txt")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Copy to destination
	dstDir := filepath.Join(t.TempDir(), "clone-copy")
	s := NewDirectoryCopyStrategy()

	if err := s.Create(srcDir, dstDir, ""); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify files exist
	content, err := os.ReadFile(filepath.Join(dstDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}

	// Verify nested file
	content, err = os.ReadFile(filepath.Join(dstDir, "subdir", "nested.txt"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(content) != "nested content" {
		t.Errorf("expected 'nested content', got %q", string(content))
	}

	// Verify symlink was preserved
	linkTarget, err := os.Readlink(filepath.Join(dstDir, "link.txt"))
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkTarget != "hello.txt" {
		t.Errorf("expected symlink target 'hello.txt', got %q", linkTarget)
	}

	// Test Remove
	if err := s.Remove("", dstDir, false); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := os.Stat(dstDir); !os.IsNotExist(err) {
		t.Error("expected directory to be removed")
	}
}

func TestDirectoryCopyStrategy_IsDirty_NonGitDir(t *testing.T) {
	dir := t.TempDir()

	s := NewDirectoryCopyStrategy()
	dirty, files, err := s.IsDirty(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Error("expected non-git dir to not be dirty")
	}
	if len(files) != 0 {
		t.Errorf("expected no changed files, got %d", len(files))
	}
}
