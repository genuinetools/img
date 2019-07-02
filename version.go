package main

import (
	"fmt"
	"github.com/genuinetools/img/internal/binutils"
	"github.com/genuinetools/img/version"
	"github.com/spf13/cobra"
	"runtime"
)

const versionHelp = `Show the version information.`

// newVersionCommand creates a command that returns information about the current build
func newVersionCommand() *cobra.Command {

	version := &versionCommand{}

	cmd := &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 versionHelp,
		Long:                  versionHelp,
		Args:                  validateHasNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.Run(args)
		},
	}

	return cmd
}

type versionCommand struct{}

func (cmd *versionCommand) Run(args []string) error {
	printImgVersion()
	printRuncVersion()

	return nil
}

func printImgVersion() {
	fmt.Printf(`%s:
 version     : %s
 git hash    : %s
 go version  : %s
 go compiler : %s
 platform    : %s/%s
`, "img", version.VERSION, version.GITCOMMIT,
		runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
}

func printRuncVersion() {
	v, err := binutils.GetRuncVersion()
	if err != nil {
		fmt.Printf(`runc: 
 error: %s
`, err)
		return
	}

	fmt.Printf(`%s:
 version     : %s
 commit      : %s
 spec        : %s
`, "runc", v.Runc, v.Commit,
		v.Spec)
}
