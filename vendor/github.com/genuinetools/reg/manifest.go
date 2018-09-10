package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/genuinetools/reg/registry"
)

const manifestHelp = `Get the json manifest for a repository.`

func (cmd *manifestCommand) Name() string      { return "manifest" }
func (cmd *manifestCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *manifestCommand) ShortHelp() string { return manifestHelp }
func (cmd *manifestCommand) LongHelp() string  { return manifestHelp }
func (cmd *manifestCommand) Hidden() bool      { return false }

func (cmd *manifestCommand) Register(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.v1, "v1", false, "force the version of the manifest retrieved to v1")
}

type manifestCommand struct {
	v1 bool
}

func (cmd *manifestCommand) Run(ctx context.Context, args []string) error {
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

	var manifest interface{}
	if cmd.v1 {
		// Get the v1 manifest if it was explicitly asked for.
		manifest, err = r.ManifestV1(image.Path, image.Reference())
		if err != nil {
			return err
		}
	} else {
		// Get the v2 manifest.
		manifest, err = r.Manifest(image.Path, image.Reference())
		if err != nil {
			return err
		}
	}

	b, err := json.MarshalIndent(manifest, " ", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
