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

const tagHelp = `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`

func (cmd *tagCommand) Name() string      { return "tag" }
func (cmd *tagCommand) Args() string      { return "SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]" }
func (cmd *tagCommand) ShortHelp() string { return tagHelp }
func (cmd *tagCommand) LongHelp() string  { return tagHelp }
func (cmd *tagCommand) Hidden() bool      { return false }

func (cmd *tagCommand) Register(fs *flag.FlagSet) {}

type tagCommand struct {
	image  string
	target string
}

func (cmd *tagCommand) Run(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("must pass an image or repository and target to tag")
	}

	// Get the specified image and target.
	cmd.image = args[0]
	cmd.target = args[1]

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

	if err := c.TagImage(ctx, cmd.image, cmd.target); err != nil {
		return err
	}

	fmt.Printf("Successfully tagged %s as %s", cmd.image, cmd.target)

	return nil
}
