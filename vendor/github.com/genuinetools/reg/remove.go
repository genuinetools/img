package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/genuinetools/reg/registry"
)

const removeHelp = `Delete a specific reference of a repository.`

func (cmd *removeCommand) Name() string      { return "rm" }
func (cmd *removeCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *removeCommand) ShortHelp() string { return removeHelp }
func (cmd *removeCommand) LongHelp() string  { return removeHelp }
func (cmd *removeCommand) Hidden() bool      { return false }

func (cmd *removeCommand) Register(fs *flag.FlagSet) {}

type removeCommand struct{}

func (cmd *removeCommand) Run(ctx context.Context, args []string) error {
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

	if err := image.WithDigest(digest); err != nil {
		return err
	}

	// Delete the reference.
	if err := r.Delete(image.Path, digest); err != nil {
		return err
	}
	fmt.Printf("Deleted %s\n", image.String())

	return nil
}
