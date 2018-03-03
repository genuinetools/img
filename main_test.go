package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	// exeSuffix is the suffix of executable files; ".exe" on Windows.
	exeSuffix string
	// stateDir is the temporary state directory used for the tests.
	testStateDir string

	mu sync.Mutex
)

func init() {
	switch runtime.GOOS {
	case "windows":
		exeSuffix = ".exe"
	}
}

// The TestMain function creates a img command for testing purposes and
// deletes it after the tests have been run.
func TestMain(m *testing.M) {
	os.Unsetenv("IMG_RUNNING_TESTS")
	args := []string{"build", "-o", "testimg" + exeSuffix}
	out, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "building testimg failed: %v\n%s\n", err, out)
		os.Exit(2)
	}

	// Create the temporary state directory.
	testStateDir, err = ioutil.TempDir("", "img-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temporary directory failed: %v\n", err)
		os.Exit(2)
	}
	defer os.RemoveAll(testStateDir)

	r := m.Run()

	os.Remove("testimg" + exeSuffix)

	os.Exit(r)
}

// doRun runs the test command, recording stdout and stderr and
// returning exit status.
func doRun(args []string, stdin io.Reader) (string, error) {
	prog := "./testimg" + exeSuffix

	newargs := []string{args[0], "--state", testStateDir}
	newargs = append(newargs, args[1:]...)

	// TODO(genuinetools): the sudo here is horrible, I know.
	cmd := exec.Command(prog, newargs...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("Error running %s: %s\n%v", strings.Join(newargs, " "), string(out), err)
	}

	return string(out), nil
}

// run runs the test command, and expects it to succeed.
func run(t *testing.T, args ...string) string {
	if runtime.GOOS == "windows" {
		mu.Lock()
		defer mu.Unlock()
	}

	out, err := doRun(args, nil)
	if err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}

	return out
}

func runBuild(t *testing.T, name string, stdin io.Reader) {
	if runtime.GOOS == "windows" {
		mu.Lock()
		defer mu.Unlock()
	}

	buildCtx := "."
	if stdin != nil {
		buildCtx = "-"
	}
	args := []string{"build", "-t", name, buildCtx}
	if _, err := doRun(args, stdin); err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}
}

func withDockerfile(dockerfile string) io.Reader {
	return strings.NewReader(dockerfile)
}
