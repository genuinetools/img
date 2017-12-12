package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/builder/dockerfile/parser"
)

const buildShortHelp = `Build an image from a Dockerfile.`
const buildLongHelp = `
`

func (cmd *buildCommand) Name() string      { return "build" }
func (cmd *buildCommand) Args() string      { return "[OPTIONS] PATH" }
func (cmd *buildCommand) ShortHelp() string { return buildShortHelp }
func (cmd *buildCommand) LongHelp() string  { return buildLongHelp }
func (cmd *buildCommand) Hidden() bool      { return false }

func (cmd *buildCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.dockerfilePath, "f", "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
}

type buildCommand struct {
	buildCtx       io.ReadCloser
	contextDir     string
	dockerfilePath string
	relDockerfile  string
	dockerfileCtx  io.ReadCloser
}

func (cmd *buildCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass a path to build")
	}

	specifiedContext := args[0]

	// Parse what is set to come from stdin.
	if cmd.dockerfilePath == "-" {
		if specifiedContext == "-" {
			return errors.New("invalid argument: can't use stdin for both build context and dockerfile")
		}

		cmd.dockerfileCtx = os.Stdin
	}

	switch {
	case specifiedContext == "-":
		// buildCtx is tar archive. if stdin was dockerfile then it is wrapped
		cmd.buildCtx, cmd.relDockerfile, err = build.GetContextFromReader(os.Stdin, cmd.dockerfilePath)
	case isLocalDir(specifiedContext):
		cmd.contextDir, cmd.relDockerfile, err = build.GetContextFromLocalDir(specifiedContext, cmd.dockerfilePath)
	default:
		return fmt.Errorf("unable to prepare context: path %q not found", specifiedContext)
	}
	if err != nil {
		return fmt.Errorf("unable to prepare context: %s", err)
	}

	// if dockerfile was not from stdin then read from file
	// to the same reader that is usually stdin
	if cmd.dockerfileCtx == nil {
		cmd.dockerfileCtx, err = os.Open(cmd.relDockerfile)
		if err != nil {
			return fmt.Errorf("failed to open %q: %v", cmd.relDockerfile, err)
		}
		defer cmd.dockerfileCtx.Close()
	}

	// Parse the Dockerfile.
	result, err := parser.Parse(cmd.dockerfileCtx)
	if err != nil {
		return err
	}
	ast := result.AST
	nodes := []*parser.Node{ast}
	if ast.Children != nil {
		nodes = append(nodes, ast.Children...)
	}

	// get the image name
	var image string
	for _, n := range nodes {
		if n.Value == "from" {
			image = n.Next.Value
			// TODO: actually parse the other shit
			break
		}
	}

	// Create the rootfs directory.
	rootFSPath, err := ioutil.TempDir("", "img-rootfs-")
	if err != nil {
		return fmt.Errorf("creating rootfs temporary directory failed: %v", err)
	}

	// Create the rootfs from the FROM image in the given Dockerfile.
	return createRootFS(image, rootFSPath)
}

func isLocalDir(c string) bool {
	_, err := os.Stat(c)
	return err == nil
}
