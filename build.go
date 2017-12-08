package main

import (
	"flag"
	"fmt"
	"strings"
)

const buildShortHelp = `Build a Dockerfile or OCI campatible spec into an image.`
const buildLongHelp = `
`

func (cmd *buildCommand) Name() string      { return "build" }
func (cmd *buildCommand) Args() string      { return "[root]" }
func (cmd *buildCommand) ShortHelp() string { return buildShortHelp }
func (cmd *buildCommand) LongHelp() string  { return buildLongHelp }
func (cmd *buildCommand) Hidden() bool      { return false }

func (cmd *buildCommand) Register(fs *flag.FlagSet) {}

type buildCommand struct{}

func (cmd *buildCommand) Run(args []string) error {
	fmt.Printf("build command run with args: %s", strings.Join(args, ", "))

	return nil
}
