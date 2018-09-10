package cli

import (
	"errors"
	"flag"
)

func testCasesEmpty() []testCase {
	return []testCase{
		{
			description: "nil",
			expectedErr: flag.ErrHelp,
		},
		{
			description: "empty",
			args:        []string{},
			expectedErr: flag.ErrHelp,
		},
	}
}

func testCasesUndefinedCommand() []testCase {
	return []testCase{
		{
			description: "args: foo",
			args:        []string{"foo"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo bar",
			args:        []string{"foo", "bar"},
			expectedErr: errors.New("bar: no such command"),
		},
	}
}

func testCasesWithCommands() []testCase {
	return []testCase{
		{
			description: "args: foo test",
			args:        []string{"foo", "test"},
		},
		{
			description: "args: foo test foo",
			args:        []string{"foo", "test", "foo"},
		},
		{
			description: "args: foo test foo bar",
			args:        []string{"foo", "test", "foo", "bar"},
		},
		{
			description: "args: foo error",
			args:        []string{"foo", "error"},
			expectedErr: errExpectedFromCommand,
		},
		{
			description: "args: foo error foo",
			args:        []string{"foo", "error", "foo"},
			expectedErr: errExpectedFromCommand,
		},
		{
			description: "args: foo error foo bar",
			args:        []string{"foo", "error", "foo", "bar"},
			expectedErr: errExpectedFromCommand,
		},
		{
			description:    "args: foo version",
			args:           []string{"foo", "version"},
			expectedStdout: versionCommandExpectedStdout,
		},
		{
			description:    "args: foo version foo",
			args:           []string{"foo", "version", "foo"},
			expectedStdout: versionCommandExpectedStdout,
		},
		{
			description:    "args: foo version foo bar",
			args:           []string{"foo", "version", "foo", "bar"},
			expectedStdout: versionCommandExpectedStdout,
		},
	}
}

func testCasesHelp() []testCase {
	return []testCase{
		{
			description: "args: foo --help",
			args:        []string{"foo", "--help"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo help",
			args:        []string{"foo", "help"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo -h",
			args:        []string{"foo", "-h"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo -h test foo",
			args:        []string{"foo", "-h", "test", "foo"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo help bar --thing",
			args:        []string{"foo", "help", "bar", "--thing"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo bar --help",
			args:        []string{"foo", "bar", "--help"},
			expectedErr: errors.New("bar: no such command"),
		},
		{
			description:    "args: foo test --help",
			args:           []string{"foo", "test", "--help"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: testCommandExpectedHelp,
		},
		{
			description:    "args: foo error -h",
			args:           []string{"foo", "error", "-h"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: errorCommandExpectedHelp,
		},
		{
			description: "args: foo error foo --help",
			args:        []string{"foo", "error", "foo", "--help"},
			expectedErr: flag.ErrHelp,
		},
		{
			description: "args: foo error foo bar --help",
			args:        []string{"foo", "error", "foo", "bar", "--help"},
			expectedErr: flag.ErrHelp,
		},
		{
			description:    "args: foo version --help",
			args:           []string{"foo", "version", "--help"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: versionCommandExpectedHelp,
		},
		{
			description:    "args: foo version -h",
			args:           []string{"foo", "version", "-h"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: versionCommandExpectedHelp,
		},
		{
			description:    "args: foo version --help another",
			args:           []string{"foo", "version", "--help", "another"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: versionCommandExpectedHelp,
		},
		{
			description:    "args: foo version -h another",
			args:           []string{"foo", "version", "-h", "another"},
			expectedErr:    flag.ErrHelp,
			expectedStderr: versionCommandExpectedHelp,
		},
	}
}

func testCasesWithAction() []testCase {
	return []testCase{
		{
			description: "args: foo",
			args:        []string{"foo"},
		},
		{
			description: "args: foo bar",
			args:        []string{"foo", "bar"},
		},
	}
}
