package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/civicactions/ddev-clone/pkg/clone"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <clone-name>",
	Short: "Remove a clone and clean up all its resources",
	Long:  `Remove a clone and clean up all its Docker resources, code copy, and project registration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
	Example: `  ddev clone remove feature-x
  ddev clone remove feature-x --force`,
}

func init() {
	removeCmd.Flags().Bool("force", false, "Skip confirmation for dirty worktrees/directories")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	projectRoot, err := getProjectRoot(cmd)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	_, sourceProjectName, err := resolveSourceProject(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to determine source project: %w", err)
	}

	opts := clone.RemoveOptions{
		SourceProjectName: sourceProjectName,
		SourcePath:        projectRoot,
		CloneName:         args[0],
		Force:             force,
		ConfirmFunc:       confirmRemoval,
	}

	return clone.Remove(opts)
}

// confirmRemoval prompts the user to confirm removal of a dirty clone.
func confirmRemoval(cloneName string, changedFiles []string) bool {
	fmt.Printf("Clone '%s' has uncommitted changes:\n", cloneName)
	for _, f := range changedFiles {
		fmt.Printf("  %s\n", f)
	}
	fmt.Print("Remove anyway? [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
