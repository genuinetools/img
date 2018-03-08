package main

import (
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/jessfraz/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const pushHelp = `Push an image or a repository to a registry.`

func (cmd *pushCommand) Name() string      { return "push" }
func (cmd *pushCommand) Args() string      { return "[OPTIONS] NAME[:TAG]" }
func (cmd *pushCommand) ShortHelp() string { return pushHelp }
func (cmd *pushCommand) LongHelp() string  { return pushHelp }
func (cmd *pushCommand) Hidden() bool      { return false }

func (cmd *pushCommand) Register(fs *flag.FlagSet) {}

type pushCommand struct {
	image string
}

func (cmd *pushCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to push")
	}

	// Get the specified image.
	cmd.image = args[0]

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the client.
	c, err := client.New(stateDir, backend, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	fmt.Printf("Pushing %s...\n", cmd.image)

	// Snapshot the image.
	if err := c.Push(ctx, cmd.image); err != nil {
		return err
	}

	fmt.Printf("Successfully pushed %s", cmd.image)

	return nil
}
