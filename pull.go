package main

import (
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	units "github.com/docker/go-units"
	"github.com/jessfraz/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const pullHelp = `Pull an image or a repository from a registry.`

func (cmd *pullCommand) Name() string      { return "pull" }
func (cmd *pullCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *pullCommand) ShortHelp() string { return pullHelp }
func (cmd *pullCommand) LongHelp() string  { return pullHelp }
func (cmd *pullCommand) Hidden() bool      { return false }

func (cmd *pullCommand) Register(fs *flag.FlagSet) {}

type pullCommand struct {
	image string
}

func (cmd *pullCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to pull")
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

	fmt.Printf("Pulling %s...\n", cmd.image)

	ref, err := c.Pull(ctx, cmd.image)
	if err != nil {
		return err
	}

	fmt.Printf("Snapshot ref: %s\n", ref.ID())

	// Get the size.
	size, err := ref.Size(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Size: %s\n", units.BytesSize(float64(size)))

	return nil
}
