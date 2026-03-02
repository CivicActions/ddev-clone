package volume

import (
	"context"
	"fmt"

	"github.com/civicactions/ddev-clone/pkg/docker"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// VolumeCloner abstracts Docker volume duplication strategies.
type VolumeCloner interface {
	// CloneVolume copies all data from sourceVol to targetVol.
	// The target volume is created if it does not exist.
	CloneVolume(ctx context.Context, sourceVol, targetVol string) error

	// Name returns the strategy identifier (e.g., "tar-copy").
	Name() string
}

const alpineImage = "alpine:latest"

// TarCopyCloner copies Docker volumes using an ephemeral Alpine container
// that streams data via tar between source and target mounts.
type TarCopyCloner struct{}

// NewTarCopyCloner returns a new TarCopyCloner instance.
func NewTarCopyCloner() *TarCopyCloner {
	return &TarCopyCloner{}
}

// Name returns the strategy identifier.
func (t *TarCopyCloner) Name() string {
	return "tar-copy"
}

// CloneVolume copies all data from sourceVol to targetVol using an ephemeral
// Alpine container that runs: sh -c "cd /source && tar cf - . | (cd /target && tar xf -)"
func (t *TarCopyCloner) CloneVolume(ctx context.Context, sourceVol, targetVol string) error {
	cli, err := docker.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Create target volume if it does not exist
	if !docker.VolumeExists(ctx, targetVol) {
		if err := docker.CreateVolume(ctx, targetVol, nil); err != nil {
			return fmt.Errorf("failed to create target volume %q: %w", targetVol, err)
		}
	}

	// Ensure alpine image is available
	if err := ensureImage(ctx, cli, alpineImage); err != nil {
		return fmt.Errorf("failed to pull %s: %w", alpineImage, err)
	}

	// Create ephemeral container with both volumes mounted
	containerName := fmt.Sprintf("ddev-clone-copy-%s", targetVol)
	createResp, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: alpineImage,
			Cmd:   []string{"sh", "-c", "cd /source && tar cf - . | (cd /target && tar xf -)"},
		},
		HostConfig: &container.HostConfig{
			Binds: []string{
				sourceVol + ":/source:ro",
				targetVol + ":/target",
			},
		},
		Name: containerName,
	})
	if err != nil {
		return fmt.Errorf("failed to create copy container: %w", err)
	}

	// Ensure container is removed on exit
	defer func() {
		_, _ = cli.ContainerRemove(ctx, createResp.ID, client.ContainerRemoveOptions{Force: true})
	}()

	// Start the container
	if _, err := cli.ContainerStart(ctx, createResp.ID, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start copy container: %w", err)
	}

	// Wait for the container to complete
	waitResult := cli.ContainerWait(ctx, createResp.ID, client.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning,
	})

	select {
	case err := <-waitResult.Error:
		if err != nil {
			return fmt.Errorf("error waiting for copy container: %w", err)
		}
	case result := <-waitResult.Result:
		if result.StatusCode != 0 {
			return fmt.Errorf("volume copy failed with exit code %d", result.StatusCode)
		}
	case <-ctx.Done():
		return fmt.Errorf("volume copy cancelled: %w", ctx.Err())
	}

	return nil
}

// ensureImage pulls the specified image if it's not already available locally.
func ensureImage(ctx context.Context, cli *client.Client, image string) error {
	// Check if image exists locally
	_, err := cli.ImageInspect(ctx, image)
	if err == nil {
		return nil // Image already exists
	}

	// Pull the image
	resp, err := cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	return resp.Wait(ctx)
}
