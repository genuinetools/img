package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const removeHelp = `Remove one or more images.`

func (cmd *removeCommand) Name() string      { return "rm" }
func (cmd *removeCommand) Args() string      { return "[OPTIONS] IMAGE [IMAGE...]" }
func (cmd *removeCommand) ShortHelp() string { return removeHelp }
func (cmd *removeCommand) LongHelp() string  { return removeHelp }
func (cmd *removeCommand) Hidden() bool      { return false }

func (cmd *removeCommand) Register(fs *flag.FlagSet) {}

type removeCommand struct{}

func (cmd *removeCommand) Run(ctx context.Context, args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to remove")
	}

	reexec()

	// Create the context.
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, "buildkit")

	// Create the client.
	c, err := client.New(stateDir, backend, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Loop over the arguments as images and run remove.
	for _, image := range args {
		fmt.Printf("Removing %s...\n", image)

		err = c.RemoveImage(ctx, image)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully removed %s\n", image)
	}

	return nil
}
