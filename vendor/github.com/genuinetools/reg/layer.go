package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/genuinetools/reg/registry"
)

const layerHelp = `Download a layer for a repository.`

func (cmd *layerCommand) Name() string      { return "layer" }
func (cmd *layerCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *layerCommand) ShortHelp() string { return layerHelp }
func (cmd *layerCommand) LongHelp() string  { return layerHelp }
func (cmd *layerCommand) Hidden() bool      { return false }

func (cmd *layerCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.output, "output", "", "output file, defaults to stdout")
	fs.StringVar(&cmd.output, "o", "", "output file, defaults to stdout")
}

type layerCommand struct {
	output string
}

func (cmd *layerCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("pass the name of the repository")
	}

	image, err := registry.ParseImage(args[0])
	if err != nil {
		return err
	}

	// Create the registry client.
	r, err := createRegistryClient(image.Domain)
	if err != nil {
		return err
	}

	// Get the digest.
	digest, err := r.Digest(image)
	if err != nil {
		return err
	}

	// Download the layer.
	layer, err := r.DownloadLayer(image.Path, digest)
	if err != nil {
		return err
	}
	defer layer.Close()

	b, err := ioutil.ReadAll(layer)
	if err != nil {
		return err
	}

	if len(cmd.output) > 0 {
		return ioutil.WriteFile(cmd.output, b, 0644)
	}

	fmt.Fprint(os.Stdout, string(b))

	return nil
}
