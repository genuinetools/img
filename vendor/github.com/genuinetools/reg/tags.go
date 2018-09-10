package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/genuinetools/reg/registry"
)

const tagsHelp = `Get the tags for a repository.`

func (cmd *tagsCommand) Name() string      { return "tags" }
func (cmd *tagsCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *tagsCommand) ShortHelp() string { return tagsHelp }
func (cmd *tagsCommand) LongHelp() string  { return tagsHelp }
func (cmd *tagsCommand) Hidden() bool      { return false }

func (cmd *tagsCommand) Register(fs *flag.FlagSet) {}

type tagsCommand struct{}

func (cmd *tagsCommand) Run(ctx context.Context, args []string) error {
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

	tags, err := r.Tags(image.Path)
	if err != nil {
		return err
	}
	sort.Strings(tags)

	// Print the tags.
	fmt.Println(strings.Join(tags, "\n"))

	return nil
}
