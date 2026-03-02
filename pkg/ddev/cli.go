package ddev

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Describe runs "ddev describe <projectName> -j" and returns parsed project info.
func Describe(projectName string) (*DescribeResult, error) {
	cmd := exec.Command("ddev", "describe", projectName, "-j")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ddev describe %q failed: %s: %w", projectName, string(output), err)
	}

	raw, err := extractRaw(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ddev describe output: %w", err)
	}

	var result DescribeResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal describe result: %w", err)
	}
	return &result, nil
}

// ListProjects runs "ddev list -j" and returns all registered projects.
func ListProjects() ([]ProjectInfo, error) {
	cmd := exec.Command("ddev", "list", "-j")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ddev list failed: %s: %w", string(output), err)
	}

	raw, err := extractRaw(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ddev list output: %w", err)
	}

	var projects []ProjectInfo
	if err := json.Unmarshal(raw, &projects); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project list: %w", err)
	}
	return projects, nil
}

// StartProject runs "ddev start <projectName>".
func StartProject(projectName string) error {
	cmd := exec.Command("ddev", "start", projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ddev start %q failed: %s: %w", projectName, string(output), err)
	}
	return nil
}

// StartProjectByPath runs "ddev start -y" from within the given directory,
// which allows DDEV to discover and register a newly-configured project.
func StartProjectByPath(projectPath string) error {
	cmd := exec.Command("ddev", "start", "-y")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ddev start at %q failed: %s: %w", projectPath, string(output), err)
	}
	return nil
}

// StopProject runs "ddev stop <projectName>".
func StopProject(projectName string) error {
	cmd := exec.Command("ddev", "stop", projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ddev stop %q failed: %s: %w", projectName, string(output), err)
	}
	return nil
}

// DeleteProject runs "ddev delete <projectName> -O -y" to remove all project
// resources including volumes, containers, and registrations.
func DeleteProject(projectName string) error {
	cmd := exec.Command("ddev", "delete", projectName, "-O", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ddev delete %q failed: %s: %w", projectName, string(output), err)
	}
	return nil
}

// IsRunning returns true if the given project is in "running" status.
func IsRunning(projectName string) (bool, error) {
	result, err := Describe(projectName)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(result.Status, "running"), nil
}

// DescribeByPath runs "ddev describe -j" from the given directory path,
// which causes DDEV to describe the project at that location.
func DescribeByPath(projectPath string) (*DescribeResult, error) {
	cmd := exec.Command("ddev", "describe", "-j")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ddev describe at %q failed: %s: %w", projectPath, string(output), err)
	}

	raw, err := extractRaw(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ddev describe output: %w", err)
	}

	var result DescribeResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal describe result: %w", err)
	}
	return &result, nil
}

// extractRaw finds the JSON line with level="info" and a non-null raw field,
// then returns the raw field bytes. DDEV may emit multiple JSON lines (logrus).
func extractRaw(output []byte) (json.RawMessage, error) {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var envelope DdevJSONOutput
		if err := json.Unmarshal([]byte(line), &envelope); err != nil {
			continue // skip non-JSON lines
		}
		if len(envelope.Raw) > 0 && string(envelope.Raw) != "null" {
			return envelope.Raw, nil
		}
	}
	return nil, fmt.Errorf("no valid JSON raw data found in output:\n%s", string(output))
}
