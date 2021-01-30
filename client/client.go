package client

import (
	"os"
	"path/filepath"
	"net/http"

	fuseoverlayfs "github.com/AkihiroSuda/containerd-fuse-overlayfs"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/genuinetools/img/types"
	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/session"
	"github.com/sirupsen/logrus"
	"github.com/containerd/containerd/remotes/docker"
)

// Client holds the information for the client we will use for communicating
// with the buildkit controller.
type Client struct {
	backend   string
	localDirs map[string]string
	root      string

	sessionManager *session.Manager
	controller     *control.Controller
}

// New returns a new client for communicating with the buildkit controller.
func New(root, backend string, localDirs map[string]string) (*Client, error) {
	// Set the name for the directory executor.
	name := "runc"

	switch backend {
	case types.AutoBackend:
		if overlay.Supported(root) == nil {
			backend = types.OverlayFSBackend
		} else if fuseoverlayfs.Supported(root) == nil {
			backend = types.FUSEOverlayFSBackend
		} else {
			backend = types.NativeBackend
		}
		logrus.Debugf("using backend: %s", backend)
	}

	// Create the root/
	root = filepath.Join(root, name, backend)
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}

	// Create the start of the client.
	return &Client{
		backend:   backend,
		root:      root,
		localDirs: localDirs,
	}, nil
}

func configureRegistries(scheme string) docker.RegistryHosts {
	return func(host string) ([]docker.RegistryHost, error) {
		config := docker.RegistryHost{
			Client:       http.DefaultClient,
			Authorizer:   nil,
			Host:         host,
			Scheme:       scheme,
			Path:         "/v2",
			Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
		}

		if config.Client == nil {
			config.Client = http.DefaultClient
		}

		if host == "docker.io" {
			config.Host = "registry-1.docker.io"
		}

		return []docker.RegistryHost{config}, nil
	}
}

// Close safely closes the client.
// This used to shut down the FUSE server but since that was removed
// it is basically a no-op now.
func (c *Client) Close() {}
