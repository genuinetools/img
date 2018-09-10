package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/genuinetools/reg/registry"
)

const digestHelp = `Get the digest for a repository.`

func (cmd *digestCommand) Name() string      { return "digest" }
func (cmd *digestCommand) Args() string      { return "[OPTIONS] NAME[:TAG]" }
func (cmd *digestCommand) ShortHelp() string { return digestHelp }
func (cmd *digestCommand) LongHelp() string  { return digestHelp }
func (cmd *digestCommand) Hidden() bool      { return false }

func (cmd *digestCommand) Register(fs *flag.FlagSet) {}

type digestCommand struct{}

func (cmd *digestCommand) Run(ctx context.Context, args []string) error {
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

	fmt.Println(digest.String())

	return nil
}
