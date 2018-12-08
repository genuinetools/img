package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/genuinetools/img/internal/binutils"
	_ "github.com/genuinetools/img/internal/unshare"
	"github.com/genuinetools/img/types"
	"github.com/genuinetools/img/version"
	"github.com/genuinetools/pkg/cli"
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
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "img"
	p.Description = "Standalone, daemon-less, unprivileged Dockerfile and OCI compatible container image builder"
	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Build the list of available commands.
	p.Commands = []cli.Command{
		&buildCommand{},
		&diskUsageCommand{},
		&listCommand{},
		&loginCommand{},
		&logoutCommand{},
		&pruneCommand{},
		&pullCommand{},
		&pushCommand{},
		&removeCommand{},
		&saveCommand{},
		&tagCommand{},
		&unpackCommand{},
	}

	defaultStateDir := defaultStateDirectory()

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("img", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.StringVar(&backend, "backend", defaultBackend, fmt.Sprintf("backend for snapshots (%v)", validBackends))
	p.FlagSet.StringVar(&backend, "b", defaultBackend, fmt.Sprintf("backend for snapshots (%v)", validBackends))
	p.FlagSet.StringVar(&stateDir, "state", defaultStateDir, fmt.Sprintf("directory to hold the global state"))
	p.FlagSet.StringVar(&stateDir, "s", defaultStateDir, fmt.Sprintf("directory to hold the global state"))

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
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
			return fmt.Errorf("%s is not a valid snapshots backend", backend)
		}

		return nil
	}

	// Run our program.
	p.Run()
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

// If the command requires runc and we do not have it installed,
// install it from the embedded asset.
func installRuncIfDNE() error {
	if binutils.RuncBinaryExists() {
		// return early.
		return nil
	}

	if len(os.Getenv("IMG_DISABLE_EMBEDDED_RUNC")) > 0 {
		// Fail early with the error to install runc.
		return fmt.Errorf("please install `runc`")
	}

	if _, err := binutils.InstallRuncBinary(); err != nil {
		return fmt.Errorf("Installing embedded runc binary failed: %v", err)
	}

	return nil
}
