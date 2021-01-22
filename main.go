package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/genuinetools/img/internal/binutils"
	_ "github.com/genuinetools/img/internal/unshare"
	"github.com/genuinetools/img/types"
	"github.com/genuinetools/img/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	validBackends = []string{types.AutoBackend, types.NativeBackend, types.OverlayFSBackend, types.FUSEOverlayFSBackend}
)

const rootHelpTemplate = `{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

const rootUsageTemplate = `{{.Name}} -  {{.Short}}

Usage: {{if .Runnable}}{{.UseLine}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func main() {
	var printVersionAndExit bool

	cmd := &cobra.Command{
		Use:              "img [OPTIONS] COMMAND [ARG...]",
		Short:            "Standalone, daemon-less, unprivileged Dockerfile and OCI compatible container image builder",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("img: '%s' is not an img command.\nSee 'img --help'", args[0])

		},
		Version:               fmt.Sprintf("%s, build %s", version.VERSION, version.GITCOMMIT),
		DisableFlagsInUseLine: true,
	}

	cmd.SetHelpTemplate(rootHelpTemplate)
	cmd.SetUsageTemplate(rootUsageTemplate)

	cmd.AddCommand(
		newBuildCommand(),
		newDiskUsageCommand(),
		newInspectCommand(),
		newListCommand(),
		newLoadCommand(),
		newLoginCommand(),
		newLogoutCommand(),
		newPruneCommand(),
		newPullCommand(),
		newPushCommand(),
		newRemoveCommand(),
		newSaveCommand(),
		newTagCommand(),
		newUnpackCommand(),
		newVersionCommand(),
	)

	defaultStateDir := defaultStateDirectory()

	// Version flag
	cmd.Flags().BoolVarP(&printVersionAndExit, "version", "v", false, "Print version information and quit")

	// Setup the global flags.
	flags := cmd.PersistentFlags()
	flags.BoolVarP(&debug, "debug", "d", false, "enable debug logging")
	flags.StringVarP(&backend, "backend", "b", defaultBackend, fmt.Sprintf("backend for snapshots (%v)", validBackends))
	flags.StringVarP(&stateDir, "state", "s", defaultStateDir, "directory to hold the global state")

	// Set the before function.
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if printVersionAndExit {
			fmt.Printf("img %s, build %s", version.VERSION, version.GITCOMMIT)
			os.Exit(0)
		}

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

		// check that runc is available
		b := binutils.BinaryAvailabilityCheck{
			StateDir:            stateDir,
			DisableEmbeddedRunc: len(os.Getenv("IMG_DISABLE_EMBEDDED_RUNC")) > 0,
		}
		err := b.EnsureRuncIsAvailable()
		if err != nil {
			return err
		}

		return nil
	}

	// Run our program.
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
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
