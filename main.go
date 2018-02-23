package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jessfraz/img/types"
	"github.com/sirupsen/logrus"
)

const (
	defaultBackend        = types.NaiveBackend
	defaultDockerRegistry = "https://index.docker.io/v1/"
	defaultDockerfileName = "Dockerfile"

	latestTagSuffix = ":latest"
)

var (
	backend  string
	stateDir string
	debug    bool

	defaultStateDirectory = "/tmp/img"

	validBackends = []string{types.FUSEBackend, types.NaiveBackend, types.OverlayFSBackend}
)

type command interface {
	Name() string           // "foobar"
	Args() string           // "<baz> [quux...]"
	ShortHelp() string      // "Foo the first bar"
	LongHelp() string       // "Foo the first bar meeting the following conditions..."
	Register(*flag.FlagSet) // command-specific flags
	Hidden() bool           // indicates whether the command should be hidden from help output
	Run([]string) error
}

// stringSlice is a slice of strings
type stringSlice []string

// implement the flag interface for stringSlice
func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	// Build the list of available commands.
	commands := []command{
		&buildCommand{},
		&diskUsageCommand{},
		&listCommand{},
		&loginCommand{},
		&pullCommand{},
		&pushCommand{},
		&tagCommand{},
		&versionCommand{},
	}

	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", os.Args[0])
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr)
		w := tabwriter.NewWriter(os.Stderr, 0, 4, 2, ' ', 0)
		for _, command := range commands {
			if !command.Hidden() {
				fmt.Fprintf(w, "\t%s\t%s\n", command.Name(), command.ShortHelp())
			}
		}
		w.Flush()
		fmt.Fprintln(os.Stderr)
	}

	if len(os.Args) <= 1 || len(os.Args) == 2 && (strings.Contains(strings.ToLower(os.Args[1]), "help") || strings.ToLower(os.Args[1]) == "-h") {
		usage()
		os.Exit(1)
	}

	for _, command := range commands {
		if name := command.Name(); os.Args[1] == name {
			// Build flag set with global flags in there.
			fs := flag.NewFlagSet(name, flag.ExitOnError)
			fs.BoolVar(&debug, "d", false, "enable debug logging")
			fs.StringVar(&backend, "backend", defaultBackend, fmt.Sprintf("backend for snapshots (%v)", validBackends))
			fs.StringVar(&stateDir, "state", defaultStateDirectory, fmt.Sprintf("directory to hold the global state"))

			// Register the subcommand flags in there, too.
			command.Register(fs)

			// Override the usage text to something nicer.
			resetUsage(fs, command.Name(), command.Args(), command.LongHelp())

			// Parse the flags the user gave us.
			if err := fs.Parse(os.Args[2:]); err != nil {
				fs.Usage()
				os.Exit(1)
			}

			// set log level
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}

			// Make sure we have a valid backend.
			found := false
			for _, vb := range validBackends {
				if vb == backend {
					found = true
					break
				}
			}
			if !found {
				logrus.Fatalf("%s is not a valid snapshots backend", backend)
			}

			// Run the command with the post-flag-processing args.
			if err := command.Run(fs.Args()); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			// Easy peasy livin' breezy.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "%s: no such command\n", os.Args[1])
	usage()
	os.Exit(1)
}

func resetUsage(fs *flag.FlagSet, name, args, longHelp string) {
	var (
		hasFlags   bool
		flagBlock  bytes.Buffer
		flagWriter = tabwriter.NewWriter(&flagBlock, 0, 4, 2, ' ', 0)
	)
	fs.VisitAll(func(f *flag.Flag) {
		hasFlags = true
		// Default-empty string vars should read "(default: <none>)"
		// rather than the comparatively ugly "(default: )".
		defValue := f.DefValue
		if defValue == "" {
			defValue = "<none>"
		}
		fmt.Fprintf(flagWriter, "\t-%s\t%s (default: %s)\n", f.Name, f.Usage, defValue)
	})
	flagWriter.Flush()
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s\n", os.Args[0], name, args)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, strings.TrimSpace(longHelp))
		fmt.Fprintln(os.Stderr)
		if hasFlags {
			fmt.Fprintln(os.Stderr, "Flags:")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, flagBlock.String())
		}
	}
}
