package cli_test

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/genuinetools/pkg/cli"
)

const yoHelp = `Send "yo" to the program.`

func (cmd *yoCommand) Name() string      { return "yo" }
func (cmd *yoCommand) Args() string      { return "" }
func (cmd *yoCommand) ShortHelp() string { return yoHelp }
func (cmd *yoCommand) LongHelp() string  { return yoHelp }
func (cmd *yoCommand) Hidden() bool      { return false }

func (cmd *yoCommand) Register(fs *flag.FlagSet) {}

type yoCommand struct{}

func (cmd *yoCommand) Run(ctx context.Context, args []string) error {
	fmt.Fprintln(os.Stdout, "yo")
	return nil
}

func ExampleNewProgram_withCommand() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "yo"
	p.Description = `A tool that prints "yo" when you run the command "yo"`

	// Set the GitCommit and Version.
	p.GitCommit = "ef2f64f"
	p.Version = "v0.1.0"

	// Setup the global flags.
	var (
		debug bool
	)
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			// Setup your logger here...
		}

		return nil
	}

	// Add our commands.
	p.Commands = []cli.Command{
		&yoCommand{},
	}

	// Run our program.
	p.Run()
	// Output: yo
}
