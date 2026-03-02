package docker

import (
	"sync"

	dockerclient "github.com/moby/moby/client"
)

var (
	dockerClient *dockerclient.Client
	clientOnce   sync.Once
	clientErr    error
)

// GetClient returns a Docker SDK client as a lazy singleton.
func GetClient() (*dockerclient.Client, error) {
	clientOnce.Do(func() {
		dockerClient, clientErr = dockerclient.NewClientWithOpts(
			dockerclient.FromEnv,
			dockerclient.WithAPIVersionNegotiation(),
		)
	})
	return dockerClient, clientErr
}
