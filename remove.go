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

const removeHelp = `Remove image.`

func (cmd *removeCommand) Name() string      { return "rm" }
func (cmd *removeCommand) Args() string      { return "NAME[:TAG]" }
func (cmd *removeCommand) ShortHelp() string { return removeHelp }
func (cmd *removeCommand) LongHelp() string  { return removeHelp }
func (cmd *removeCommand) Hidden() bool      { return false }

func (cmd *removeCommand) Register(fs *flag.FlagSet) {}

type removeCommand struct {
	image string
}

func (cmd *removeCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to remove")
	}

	// Get the specified image.
	cmd.image = args[0]
	// Add the latest lag if they did not provide one.
	cmd.image = addLatestTagSuffix(cmd.image)

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

	fmt.Printf("Removing %s...\n", cmd.image)

	err = c.RemoveImage(ctx, cmd.image)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully removed %s\n", cmd.image)

	return nil
}
