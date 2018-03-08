package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/pkg/archive"
	"github.com/jessfraz/img/client"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const buildHelp = `Build an image from a Dockerfile.`

func (cmd *buildCommand) Name() string      { return "build" }
func (cmd *buildCommand) Args() string      { return "[OPTIONS] PATH" }
func (cmd *buildCommand) ShortHelp() string { return buildHelp }
func (cmd *buildCommand) LongHelp() string  { return buildHelp }
func (cmd *buildCommand) Hidden() bool      { return false }

func (cmd *buildCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.dockerfilePath, "f", "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
	fs.StringVar(&cmd.tag, "t", "", "Name and optionally a tag in the 'name:tag' format")
	fs.StringVar(&cmd.target, "target", "", "Set the target build stage to build")
	fs.Var(&cmd.buildArgs, "build-arg", "Set build-time variables")
}

type buildCommand struct {
	buildArgs      stringSlice
	dockerfilePath string
	target         string
	tag            string

	contextDir string
}

func (cmd *buildCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass a path to build")
	}

	if cmd.tag == "" {
		return errors.New("please specify an image tag with `-t`")
	}

	// Get the specified context.
	cmd.contextDir = args[0]

	// Parse what is set to come from stdin.
	if cmd.dockerfilePath == "-" {
		cmd.dockerfilePath, err = dockerfileFromStdin()
		if err != nil {
			return fmt.Errorf("reading dockerfile from stdin failed: %v", err)
		}
		// On exit cleanup the temporary file we used hold the dockerfile from stdin.
		defer os.RemoveAll(cmd.dockerfilePath)
	}

	if cmd.contextDir == "" {
		return errors.New("please specify build context (e.g. \".\" for the current directory)")
	}

	if cmd.contextDir == "-" {
		cmd.contextDir, err = contextFromStdin(cmd.dockerfilePath)
		if err != nil {
			return fmt.Errorf("reading context from stdin failed: %v", err)
		}
		// On exit cleanup the temporary directory we used hold the files from stdin.
		defer os.RemoveAll(cmd.contextDir)
	}

	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(cmd.tag)
	if err != nil {
		return fmt.Errorf("parsing image name %q failed: %v", cmd.tag, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	cmd.tag = named.String()

	// Set the dockerfile path as the default if one was not given.
	if cmd.dockerfilePath == "" {
		cmd.dockerfilePath = filepath.Join(cmd.contextDir, defaultDockerfileName)
	}

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the client.
	c, err := client.New(stateDir, backend, cmd.getLocalDirs())
	if err != nil {
		return err
	}
	defer c.Close()

	// Create the frontend attrs.
	frontendAttrs := map[string]string{
		// We use the base for filename here becasue we already set up the local dirs which sets the path in createController.
		"filename": filepath.Base(cmd.dockerfilePath),
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
	if err := c.Solve(ctx, &controlapi.SolveRequest{
		Ref:      id,
		Session:  id,
		Exporter: "image",
		ExporterAttrs: map[string]string{
			"name": cmd.tag,
		},
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
	}); err != nil {
		return err
	}

	fmt.Printf("Successfully built %s\n", cmd.tag)

	return nil
}

// dockerfileFromStdin copies a dockerfile from stdin to a temporary file.
func dockerfileFromStdin() (string, error) {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading from stdin failed: %v", err)
	}

	// Create a temporary file for the Dockerfile
	f, err := ioutil.TempFile("", "img-build-dockerfile-")
	if err != nil {
		return f.Name(), fmt.Errorf("unable to create temporary file for dockerfile: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(stdin); err != nil {
		return f.Name(), fmt.Errorf("writing to temporary file for dockerfile failed: %v", err)
	}

	return f.Name(), nil
}

// contextFromStdin will read the contents of stdin as either a
// Dockerfile or tar archive. Returns the path to a temporary directory
// for the build context..
func contextFromStdin(dockerfileName string) (string, error) {
	// Set the dockerfile name if it is empty.
	if dockerfileName == "" {
		dockerfileName = defaultDockerfileName
	}

	// Create a temporary directory for the build context.
	tmpDir, err := ioutil.TempDir("", "img-build-context-")
	if err != nil {
		return "", fmt.Errorf("unable to create temporary context directory: %v", err)
	}

	// Create a new reader from stdin.
	buf := bufio.NewReader(os.Stdin)

	// Grab the magic number range from the reader.
	archiveHeaderSize := 512 // number of bytes in an archive header
	magic, err := buf.Peek(archiveHeaderSize)
	if err != nil && err != io.EOF {
		return tmpDir, fmt.Errorf("failed to peek context header from STDIN: %v", err)
	}

	// Validate if it is a tar archive.
	if isArchive(magic) {
		return tmpDir, untar(tmpDir, buf)
	}

	if dockerfileName == "-" {
		return tmpDir, errors.New("build context is not an archive")
	}

	// Create the dockerfile in the temporary directory.
	f, err := os.Create(filepath.Join(tmpDir, dockerfileName))
	if err != nil {
		return tmpDir, err
	}
	defer f.Close()

	// Copy the contents of the reader to the file.
	_, err = io.Copy(f, buf)
	return tmpDir, err
}

// isArchive checks for the magic bytes of a tar or any supported compression algorithm.
func isArchive(header []byte) bool {
	compression := archive.DetectCompression(header)
	if compression != archive.Uncompressed {
		return true
	}
	r := tar.NewReader(bytes.NewBuffer(header))
	_, err := r.Next()
	return err == nil
}

// untar unpacks a tarball to a given directory.
func untar(dest string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dest, header.Name)

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}

func (cmd *buildCommand) getLocalDirs() map[string]string {
	return map[string]string{
		"context":    cmd.contextDir,
		"dockerfile": filepath.Dir(cmd.dockerfilePath),
	}
}
