package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/genuinetools/img/internal/binutils"
	_ "github.com/genuinetools/img/internal/unshare"
	"github.com/genuinetools/img/types"
	"github.com/sirupsen/logrus"
)

const (
	defaultBackend        = types.AutoBackend
	defaultDockerRegistry = "https://index.docker.io/v1/"
	defaultDockerfileName = "Dockerfile"
)

var (
	backend  string
	stateDir string
	debug    bool

	validBackends = []string{types.AutoBackend, types.NativeBackend, types.OverlayFSBackend}
)

type command interface {
	Name() string           // "foobar"
	Args() string           // "<baz> [quux...]"
	ShortHelp() string      // "Foo the first bar"
	LongHelp() string       // "Foo the first bar meeting the following conditions..."
	Register(*flag.FlagSet) // command-specific flags
	Hidden() bool           // indicates whether the command should be hidden from help output
	DoReexec() bool         // indicates whether the command should preform a re-exec or not
	RequiresRunc() bool     // indicates whether the command requires the runc binary
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

func defaultStateDirectory() string {
	//  pam_systemd sets XDG_RUNTIME_DIR but not other dirs.
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		dirs := strings.Split(xdgDataHome, ":")
		return filepath.Join(dirs[0], "img")
	}
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, ".local", "share", "img")
	}
	return "/tmp/img"
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
		&removeCommand{},
		&saveCommand{},
		&tagCommand{},
		&unpackCommand{},
		&versionCommand{},
	}

	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", "img")
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

	defaultStateDir := defaultStateDirectory()
	for _, command := range commands {
		if name := command.Name(); os.Args[1] == name {
			// Build flag set with global flags in there.
			fs := flag.NewFlagSet(name, flag.ExitOnError)
			fs.BoolVar(&debug, "d", false, "enable debug logging")
			fs.StringVar(&backend, "backend", defaultBackend, fmt.Sprintf("backend for snapshots (%v)", validBackends))
			fs.StringVar(&stateDir, "state", defaultStateDir, fmt.Sprintf("directory to hold the global state"))

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

			// Perform the re-exec if necessary.
			if command.DoReexec() {
				reexec()
			}

			// If the command requires runc and we do not have it installed,
			// install it from the embedded asset.
			if command.RequiresRunc() && !binutils.RuncBinaryExists() {
				if len(os.Getenv("IMG_DISABLE_EMBEDDED_RUNC")) > 0 {
					// Fail early with the error to install runc.
					logrus.Fatal("please install `runc`")
				}
				runcDir, err := binutils.InstallRuncBinary()
				if err != nil {
					os.RemoveAll(runcDir)
					logrus.Fatalf("Installing embedded runc binary failed: %v", err)
				}
				defer os.RemoveAll(runcDir)
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
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s\n", "img", name, args)
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
