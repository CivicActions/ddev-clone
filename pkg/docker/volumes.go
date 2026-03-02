package docker

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"
)

// VolumeExists checks if a Docker volume with the given name exists.
func VolumeExists(ctx context.Context, name string) bool {
	cli, err := GetClient()
	if err != nil {
		return false
	}
	_, err = cli.VolumeInspect(ctx, name, client.VolumeInspectOptions{})
	return err == nil
}

// CreateVolume creates a Docker volume with the given name and labels.
func CreateVolume(ctx context.Context, name string, labels map[string]string) error {
	cli, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	_, err = cli.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name:   name,
		Labels: labels,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume %q: %w", name, err)
	}
	return nil
}

// RemoveVolume removes a Docker volume by name (force=true).
func RemoveVolume(ctx context.Context, name string) error {
	cli, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	_, err = cli.VolumeRemove(ctx, name, client.VolumeRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove volume %q: %w", name, err)
	}
	return nil
}

// ListProjectVolumes returns all volume names with the given Docker Compose project label.
func ListProjectVolumes(ctx context.Context, composeProject string) ([]string, error) {
	cli, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	resp, err := cli.VolumeList(ctx, client.VolumeListOptions{
		Filters: client.Filters{}.Add("label", "com.docker.compose.project="+composeProject),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes for project %q: %w", composeProject, err)
	}

	var names []string
	for _, v := range resp.Items {
		names = append(names, v.Name)
	}
	return names, nil
}
