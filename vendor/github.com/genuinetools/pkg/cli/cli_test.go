package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

const (
	testCommandExpectedHelp = `Usage: yo test` + " " + `

Show the test information.

`
	errorCommandExpectedHelp = `Usage: yo error` + " " + `

Show the error information.

`
	versionCommandExpectedHelp = `Usage: yo version` + " " + `

Show the version information.

`
)

var (
	nilFunction = func(ctx context.Context) error {
		return nil
	}
	nilActionFunction = func(ctx context.Context, args []string) error {
		return nil
	}

	errExpected            = errors.New("expected error")
	errExpectedFromCommand = errors.New("expected error command error")
	errFunction            = func(ctx context.Context) error {
		return errExpected
	}

	versionCommandExpectedStdout = fmt.Sprintf(`yo:
 version     : 0.0.0
 git hash    :`+" "+`
 go version  : %s
 go compiler : %s
 platform    : %s/%s
`, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
)

type testCase struct {
	description      string
	args             []string
	shouldPrintUsage bool
	expectedErr      error
	expectedStderr   string
	expectedStdout   string
}

// Define the testCommand.
type testCommand struct{}

func (cmd *testCommand) Name() string                                 { return "test" }
func (cmd *testCommand) Args() string                                 { return "" }
func (cmd *testCommand) ShortHelp() string                            { return "Show the test information." }
func (cmd *testCommand) LongHelp() string                             { return "Show the test information." }
func (cmd *testCommand) Hidden() bool                                 { return false }
func (cmd *testCommand) Register(fs *flag.FlagSet)                    {}
func (cmd *testCommand) Run(ctx context.Context, args []string) error { return nil }

// Define the errorCommand.
type errorCommand struct{}

func (cmd *errorCommand) Name() string                                 { return "error" }
func (cmd *errorCommand) Args() string                                 { return "" }
func (cmd *errorCommand) ShortHelp() string                            { return "Show the error information." }
func (cmd *errorCommand) LongHelp() string                             { return "Show the error information." }
func (cmd *errorCommand) Hidden() bool                                 { return false }
func (cmd *errorCommand) Register(fs *flag.FlagSet)                    {}
func (cmd *errorCommand) Run(ctx context.Context, args []string) error { return errExpectedFromCommand }

func TestProgramUsage(t *testing.T) {
	var (
		debug  bool
		token  string
		output string

		expectedOutput = `sample -  My sample command line tool.

Usage: sample <command>

Flags:

  -d, --debug  enable debug logging (default: false)
  -o           where to save the output (default: defaultOutput)
  -t, --thing  a flag for thing (default: false)
  --token      API token (default: <none>)

Commands:

  error    Show the error information.
  test     Show the test information.
  version  Show the version information.

`

		expectedVersionOutput = `Usage: sample version` + " " + `

Show the version information.

Flags:

  -d, --debug  enable debug logging (default: false)
  -o           where to save the output (default: defaultOutput)
  -t, --thing  a flag for thing (default: false)
  --token      API token (default: <none>)

`
	)

	// Setup the program.
	p := NewProgram()
	p.Name = "sample"
	p.Description = "My sample command line tool"

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&token, "token", "", "API token")
	p.FlagSet.StringVar(&output, "o", "defaultOutput", "where to save the output")
	p.FlagSet.BoolVar(&debug, "thing", false, "a flag for thing")
	p.FlagSet.BoolVar(&debug, "t", false, "a flag for thing")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")

	p.Commands = []Command{
		&errorCommand{},
		&testCommand{},
	}
	p.Action = nilActionFunction

	p.Run()

	c := startCapture(t)
	if err := p.usage(p.defaultContext()); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := c.finish()
	if stderr != expectedOutput {
		t.Fatalf("expected: %s\ngot: %s", expectedOutput, stderr)
	}
	if len(stdout) > 0 {
		t.Fatalf("expected no stdout, got: %s", stdout)
	}

	// Test versionCommand.
	vcmd := &versionCommand{}
	c = startCapture(t)
	p.resetCommandUsage(vcmd)
	p.FlagSet.Usage()
	stdout, stderr = c.finish()
	if stderr != expectedVersionOutput {
		t.Fatalf("expected: %q\ngot: %q", expectedVersionOutput, stderr)
	}
	if len(stdout) > 0 {
		t.Fatalf("expected no stdout, got: %s", stdout)
	}
}

func TestProgramWithNoCommandsOrFlagsOrAction(t *testing.T) {
	p := NewProgram()
	p.Name = "yo"
	testCases := append(testCasesEmpty(), testCasesUndefinedCommand()...)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}
}

func TestProgramWithNoCommandsOrFlags(t *testing.T) {
	p := NewProgram()
	p.Name = "yo"
	p.Action = nilActionFunction
	testCases := append(testCasesEmpty(), testCasesWithAction()...)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}
}

func TestProgramHelpFlag(t *testing.T) {
	p := NewProgram()
	p.Name = "yo"
	p.FlagSet = flag.NewFlagSet("global", flag.ContinueOnError)
	p.Commands = []Command{
		&testCommand{},
		&errorCommand{},
	}
	testCases := testCasesHelp()

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := startCapture(t)
			err := p.run(p.defaultContext(), tc.args)
			stdout, stderr := c.finish()
			compareErrors(t, err, tc.expectedErr)
			if tc.expectedStderr != stderr {
				t.Fatalf("expected stderr: %q\ngot: %q", tc.expectedStderr, stderr)
			}
			if tc.expectedStdout != stdout {
				t.Fatalf("expected stdout: %q\ngot: %q", tc.expectedStdout, stdout)
			}
		})
	}
}

func TestProgramWithCommandsAndAction(t *testing.T) {
	p := NewProgram()
	p.Name = "yo"
	p.Commands = []Command{
		&errorCommand{},
		&testCommand{},
	}
	p.Action = nilActionFunction
	testCases := append(append(testCasesEmpty(),
		testCasesWithCommands()...),
		testCasesWithAction()...)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Add a Before.
	p.Before = nilFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with Successful Before -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Add an After.
	p.After = nilFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with successful After -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Test program with an error on After.
	p.After = errFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with error on After -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Test program with an error on Before.
	p.Before = errFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with error on Before -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}
}

func TestProgramWithCommands(t *testing.T) {
	p := NewProgram()
	p.Name = "yo"
	p.Commands = []Command{
		&errorCommand{},
		&testCommand{},
	}
	testCases := append(append(testCasesEmpty(),
		testCasesUndefinedCommand()...),
		testCasesWithCommands()...)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Add a Before.
	p.Before = nilFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with Successful Before -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Add an After.
	p.After = nilFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with successful After -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Test program with an error on After.
	p.After = errFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with error on After -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}

	// Test program with an error on Before.
	p.Before = errFunction
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with error on Before -> %s", tc.description), func(t *testing.T) {
			p.doTestRun(t, tc)
		})
	}
}

func compareErrors(t *testing.T, err, expectedErr error) {
	if expectedErr != nil {
		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatalf("expected error %#v got: %#v", expectedErr, err)
		}

		return
	}

	if err != expectedErr {
		t.Fatalf("expected error %#v got: %#v", expectedErr, err)
	}

	return
}

type capture struct {
	stdout, stderr *os.File
	ro, re         *os.File
	wo, we         *os.File
	co, ce         chan string
}

func startCapture(t *testing.T) capture {
	c := capture{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	// Pipe it to a reader and writer.
	var err error
	c.ro, c.wo, err = os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = c.wo
	c.re, c.we, err = os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = c.we

	return c
}

func (c *capture) finish() (string, string) {
	defer c.ro.Close()
	defer c.re.Close()

	// Copy the output in a separate goroutine so printing can't block indefinitely.
	c.co = make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, c.ro)
		c.co <- buf.String()
	}()
	c.ce = make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, c.re)
		c.ce <- buf.String()
	}()

	// Close everything.
	c.wo.Close()
	c.we.Close()

	// Reset.
	os.Stdout = c.stdout
	os.Stderr = c.stderr

	o := <-c.co
	e := <-c.ce
	return o, e
}

func (p *Program) isErrorOnBefore() bool {
	return p.Before != nil && p.Before(context.Background()) != nil
}

func (p *Program) isErrorOnAfter() bool {
	return p.After != nil && p.After(context.Background()) != nil
}

func (tc *testCase) expectUsageToBePrintedBeforeBefore(p *Program) bool {
	return tc.args == nil || len(tc.args) < 1 ||
		(p.Action == nil && len(p.Commands) > 1 && len(tc.args) < 2) ||
		(tc.expectedErr != nil && strings.Contains(tc.expectedErr.Error(), "no such command"))
}

func (p *Program) doTestRun(t *testing.T, tc testCase) {
	c := startCapture(t)
	err := p.run(p.defaultContext(), tc.args)
	stdout, stderr := c.finish()
	if len(stderr) > 0 {
		t.Fatalf("expected no stderr, got: %s", stderr)
	}

	// Check that the stdout is what we expected.
	if !strings.HasPrefix(stdout, tc.expectedStdout) {
		t.Fatalf("expected stdout: %q\ngot: %q", tc.expectedStdout, stdout)
	}

	// IF
	// we DON'T EXPECT an error on Before OR After
	// OR
	// we EXPECT the usage to be printed (<nil> or empty)
	// OR
	// we DON'T EXPECT an error on Before but we EXPECT an error on After AND the command was EXPECTED to error
	// THEN
	// check we got the expected error defined in the testcase.
	if (!p.isErrorOnAfter() && !p.isErrorOnBefore()) ||
		tc.expectUsageToBePrintedBeforeBefore(p) ||
		(!p.isErrorOnBefore() && p.isErrorOnAfter() && tc.expectedErr == errExpectedFromCommand) {
		compareErrors(t, err, tc.expectedErr)
	}

	// IF
	// we EXPECT an error on Before
	// OR
	// we EXPECT an error on After AND the command was NOT EXPECTED to error
	// AND
	// we DON'T EXPECT the usage to be printed (<nil> or empty)
	// THEN
	// check we got the expected error from Before/After.
	if (p.isErrorOnBefore() ||
		(p.isErrorOnAfter() && tc.expectedErr != errExpectedFromCommand)) &&
		!tc.expectUsageToBePrintedBeforeBefore(p) {
		compareErrors(t, err, errExpected)
	}
}
