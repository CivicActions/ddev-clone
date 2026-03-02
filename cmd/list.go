package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/civicactions/ddev-clone/pkg/clone"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clones of the current project",
	Long:  `List all clones associated with the current project. Can be run from the source project or any of its clones.`,
	Args:  cobra.NoArgs,
	RunE:  runList,
	Example: `  ddev clone list
  ddev clone list -j`,
}

func init() {
	listCmd.Flags().BoolP("json-output", "j", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
	projectRoot, err := getProjectRoot(cmd)
	if err != nil {
		return err
	}

	jsonOutput, _ := cmd.Flags().GetBool("json-output")

	_, sourceProjectName, err := resolveSourceProject(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to determine source project: %w", err)
	}

	clones, err := clone.List(sourceProjectName, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to list clones: %w", err)
	}

	if jsonOutput {
		return printJSON(clones)
	}

	return printTable(clones, sourceProjectName)
}

func printJSON(clones []clone.CloneInfo) error {
	data, err := json.MarshalIndent(clones, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printTable(clones []clone.CloneInfo, sourceProjectName string) error {
	if len(clones) == 0 {
		fmt.Printf("No clones found for project '%s'.\n", sourceProjectName)
		fmt.Println("Create one with: ddev clone create <clone-name>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, " CLONE NAME\tPATH\tBRANCH\tSTRATEGY\tSTATUS")

	for _, c := range clones {
		marker := " "
		if c.Current {
			marker = "*"
		}
		fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\t%s\n",
			marker, c.CloneName, c.ClonePath, c.Branch, c.CodeStrategyType, c.Status)
	}

	return w.Flush()
}
