package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/containerd/containerd/namespaces"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const buildShortHelp = `Build an image from a Dockerfile.`

// TODO: make the long help actually useful
const buildLongHelp = `Build an image from a Dockerfile.`

func (cmd *buildCommand) Name() string      { return "build" }
func (cmd *buildCommand) Args() string      { return "[OPTIONS] PATH" }
func (cmd *buildCommand) ShortHelp() string { return buildShortHelp }
func (cmd *buildCommand) LongHelp() string  { return buildLongHelp }
func (cmd *buildCommand) Hidden() bool      { return false }

func (cmd *buildCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.dockerfilePath, "f", "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
	fs.StringVar(&cmd.tag, "t", "", "Name and optionally a tag in the 'name:tag' format")
	fs.StringVar(&cmd.target, "target", "", "Set the target build stage to build")
	fs.Var(&cmd.buildArgs, "build-arg", "Set build-time variables")
}

type buildCommand struct {
	buildArgs      stringSlice
	contextDir     string
	dockerfilePath string
	target         string
	tag            string
}

func (cmd *buildCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass a path to build")
	}

	// Get the specified context.
	cmd.contextDir = args[0]

	// Parse what is set to come from stdin.
	if cmd.dockerfilePath == "-" {
		return errors.New("stdin not supported for Dockerfile yet")
	}

	if cmd.contextDir == "" {
		return errors.New("please specify build context (e.g. \".\" for the current directory)")
	}

	if cmd.contextDir == "-" {
		return errors.New("stdin not supported for build context yet")
	}

	// Add the latest lag if they did not provide one.
	if !strings.Contains(cmd.tag, ":") {
		cmd.tag += ":latest"
	}

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the controller.
	c, fuseserver, err := createController(cmd)
	defer unmount(fuseserver)
	if err != nil {
		return err
	}
	handleSignals(fuseserver)

	// Create the frontend attrs.
	frontendAttrs := map[string]string{
		"filename": cmd.dockerfilePath,
		"target":   cmd.target,
	}

	// Get the build args and add them to frontend attrs.
	for _, buildArg := range cmd.buildArgs {
		kv := strings.SplitN(buildArg, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid build-arg value %s", buildArg)
		}
		frontendAttrs["build-arg:"+kv[0]] = kv[1]
	}

	fmt.Printf("Building %s\n", cmd.tag)
	fmt.Println("Setting up the rootfs... this may take a bit.")

	// Solve the dockerfile.
	_, err = c.Solve(ctx, &controlapi.SolveRequest{
		Ref:      id,
		Session:  id,
		Exporter: "image",
		ExporterAttrs: map[string]string{
			"name": cmd.tag,
		},
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
	})
	if err != nil {
		return fmt.Errorf("solving failed: %v", err)
	}

	fmt.Printf("Built image %s\n", cmd.tag)

	return nil
}
