package cmd

import (
	"fmt"

	"github.com/civicactions/ddev-clone/pkg/clone"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <clone-name>",
	Short: "Create a clone of the current DDEV project",
	Long: `Create a clone of the current DDEV project using code isolation and Docker volume duplication.

Creates a code copy (git worktree or directory copy) at ../<project>-clone-<name>,
copies all Docker volumes, configures a new DDEV project via config.local.yaml, and starts it.`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
	Example: `  ddev clone create feature-x
  ddev clone create feature-x --branch existing-branch
  ddev clone create feature-x --no-start
  ddev clone create feature-x --code-strategy=copy`,
}

func init() {
	createCmd.Flags().StringP("branch", "b", "", "Check out an existing branch (worktree strategy only)")
	createCmd.Flags().String("code-strategy", "", "Code isolation strategy: worktree or copy (default: auto-detect)")
	createCmd.Flags().Bool("no-start", false, "Configure clone but do not start it")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	projectRoot, err := getProjectRoot(cmd)
	if err != nil {
		return err
	}

	branch, _ := cmd.Flags().GetString("branch")
	codeStrategy, _ := cmd.Flags().GetString("code-strategy")
	noStart, _ := cmd.Flags().GetBool("no-start")

	_, sourceProjectName, err := resolveSourceProject(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to determine source project: %w", err)
	}

	opts := clone.CreateOptions{
		SourceProjectName: sourceProjectName,
		SourcePath:        projectRoot,
		CloneName:         args[0],
		Branch:            branch,
		CodeStrategyName:  codeStrategy,
		NoStart:           noStart,
	}

	return clone.Create(opts)
}
