package client

import (
	"os"
	"path/filepath"

	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/genuinetools/img/types"
	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/session"
	"github.com/sirupsen/logrus"
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

// Close safely closes the client.
// This used to shut down the FUSE server but since that was removed
// it is basically a no-op now.
func (c *Client) Close() {}
