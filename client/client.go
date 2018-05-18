package client

import (
	"os"
	"path/filepath"

	"github.com/hanwen/go-fuse/fuse"
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

	fuseserver *fuse.Server

	sessionManager *session.Manager
	controller     *control.Controller
}

// New returns a new client for communicating with the buildkit controller.
func New(root, backend string, localDirs map[string]string) (*Client, error) {
	// Set the name for the directory executor.
	name := "runc"

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

// Close terminates the client and unmount the fuseserver if it is mounted.
func (c *Client) Close() {
	if c.fuseserver == nil {
		return
	}

	if err := c.fuseserver.Unmount(); err != nil {
		logrus.Errorf("Unmounting FUSE server failed: %v", err)
	}
}
