package cli

import (
	"context"
	"flag"
	"fmt"
	"runtime"
)

const versionHelp = `Show the version information.`

func (cmd *versionCommand) Name() string      { return "version" }
func (cmd *versionCommand) Args() string      { return "" }
func (cmd *versionCommand) ShortHelp() string { return versionHelp }
func (cmd *versionCommand) LongHelp() string  { return versionHelp }
func (cmd *versionCommand) Hidden() bool      { return false }

func (cmd *versionCommand) Register(fs *flag.FlagSet) {}

type versionCommand struct{}

func (cmd *versionCommand) Run(ctx context.Context, args []string) error {
	fmt.Printf(`%s:
 version     : %s
 git hash    : %s
 go version  : %s
 go compiler : %s
 platform    : %s/%s
`, ctx.Value(NameKey).(string), ctx.Value(VersionKey).(string), ctx.Value(GitCommitKey).(string),
		runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
	return nil
}
