package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/civicactions/ddev-clone/pkg/clone"
	"github.com/civicactions/ddev-clone/pkg/ddev"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ddev-clone",
	Short: "Clone management for DDEV projects",
	Long:  "Create, list, remove, and prune cloned DDEV project environments.",
}

func init() {
	rootCmd.PersistentFlags().String("project-root", "", "DDEV project root (default: $DDEV_APPROOT or cwd)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// getProjectRoot resolves the DDEV project root from flag, env var, or cwd.
func getProjectRoot(cmd *cobra.Command) (string, error) {
	// 1. Check --project-root flag
	root, _ := cmd.Flags().GetString("project-root")
	if root != "" {
		return validateProjectRoot(root)
	}

	// 2. Check $DDEV_APPROOT env var
	root = os.Getenv("DDEV_APPROOT")
	if root != "" {
		return validateProjectRoot(root)
	}

	// 3. Fall back to cwd
	root, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return validateProjectRoot(root)
}

// validateProjectRoot checks that the path contains a .ddev/config.yaml file.
func validateProjectRoot(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %q: %w", root, err)
	}

	configPath := filepath.Join(absRoot, ".ddev", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("not a DDEV project: %s (missing .ddev/config.yaml)", absRoot)
	}

	return absRoot, nil
}

// resolveSourceProject describes the DDEV project at projectRoot and determines
// the source project name. If the project name contains "-clone-" but the
// suspected source project doesn't exist, the current project is treated as the source.
func resolveSourceProject(projectRoot string) (*ddev.DescribeResult, string, error) {
	desc, err := ddev.DescribeByPath(projectRoot)
	if err != nil {
		// Try to get project name from DDEV_PROJECT env var
		projectName := os.Getenv("DDEV_PROJECT")
		if projectName == "" {
			return nil, "", fmt.Errorf("failed to determine DDEV project: %w", err)
		}
		desc, err = ddev.Describe(projectName)
		if err != nil {
			return nil, "", fmt.Errorf("failed to describe project %q: %w", projectName, err)
		}
	}

	sourceProjectName := clone.GetSourceProjectName(desc.Name)
	if sourceProjectName != desc.Name {
		// Verify the suspected source actually exists; if not, the current
		// project IS the source (its name just happens to contain "-clone-").
		if _, err := ddev.Describe(sourceProjectName); err != nil {
			sourceProjectName = desc.Name
		}
	}

	return desc, sourceProjectName, nil
}
