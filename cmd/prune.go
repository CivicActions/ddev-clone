package cmd

import (
	"fmt"

	"github.com/civicactions/ddev-clone/pkg/clone"
	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up stale clone references",
	Long:  `Clean up clones whose directories no longer exist on disk. Removes DDEV project registrations and prunes git worktree references.`,
	Args:  cobra.NoArgs,
	RunE:  runPrune,
	Example: `  ddev clone prune
  ddev clone prune --dry-run`,
}

func init() {
	pruneCmd.Flags().Bool("dry-run", false, "Show what would be pruned without taking action")
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, _ []string) error {
	projectRoot, err := getProjectRoot(cmd)
	if err != nil {
		return err
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	_, sourceProjectName, err := resolveSourceProject(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to determine source project: %w", err)
	}

	opts := clone.PruneOptions{
		SourceProjectName: sourceProjectName,
		SourcePath:        projectRoot,
		DryRun:            dryRun,
	}

	pruned, err := clone.Prune(opts)
	if err != nil {
		return err
	}

	if len(pruned) == 0 {
		fmt.Printf("No stale clones found for project '%s'.\n", sourceProjectName)
		return nil
	}

	if dryRun {
		fmt.Printf("%d stale clones would be pruned.\n", len(pruned))
	} else {
		fmt.Printf("Pruned %d stale clone(s).\n", len(pruned))
	}

	return nil
}
