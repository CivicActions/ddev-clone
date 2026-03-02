package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// StopDBService stops only the database service of a DDEV project
// using docker compose.
func StopDBService(projectName string) error {
	composeName := GetComposeProjectName(projectName)
	cmd := exec.Command("docker", "compose", "-p", composeName, "stop", "db")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop db service for %q: %s: %w", projectName, string(output), err)
	}
	return nil
}

// StartDBService starts only the database service of a DDEV project
// using docker compose.
func StartDBService(projectName string) error {
	composeName := GetComposeProjectName(projectName)
	cmd := exec.Command("docker", "compose", "-p", composeName, "start", "db")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start db service for %q: %s: %w", projectName, string(output), err)
	}
	return nil
}

// GetComposeProjectName derives the Docker Compose project name from a DDEV
// project name. DDEV uses: "ddev-" + lowercase name with dots removed.
func GetComposeProjectName(ddevProjectName string) string {
	name := strings.ToLower(ddevProjectName)
	name = strings.ReplaceAll(name, ".", "")
	return "ddev-" + name
}
