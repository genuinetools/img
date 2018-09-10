package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/docker/docker/client"
	"github.com/genuinetools/reg/testutils"
)

const (
	domain = "localhost:5000"
)

var (
	exeSuffix string // ".exe" on Windows

	registryConfigs = []struct {
		config   string
		username string
		password string
	}{
		{
			config:   "noauth.yml",
			username: "blah",
			password: "blah",
		},
		{
			config:   "basicauth.yml",
			username: "admin",
			password: "testing",
		},
	}
	registryHelper *testutils.RegistryHelper
)

func init() {
	switch runtime.GOOS {
	case "windows":
		exeSuffix = ".exe"
	}
}

// The TestMain function creates a reg command for testing purposes and
// deletes it after the tests have been run.
// It also spins up a local registry prefilled with an alpine image and
// removes that after the tests have been run.
func TestMain(m *testing.M) {
	// build the test binary
	args := []string{"build", "-o", "testreg" + exeSuffix}
	out, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "building testreg failed: %v\n%s", err, out)
		os.Exit(2)
	}
	// remove test binary
	defer os.Remove("testreg" + exeSuffix)

	// create the docker client
	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(fmt.Errorf("could not connect to docker: %v", err))
	}

	// start the clair containers.
	dbID, clairID, err := testutils.StartClair(dcli)
	if err != nil {
		testutils.RemoveContainer(dcli, dbID, clairID)
		panic(fmt.Errorf("starting clair containers failed: %v", err))
	}

	for _, regConfig := range registryConfigs {
		// start each registry
		regID, _, err := testutils.StartRegistry(dcli, regConfig.config, regConfig.username, regConfig.password)
		if err != nil {
			testutils.RemoveContainer(dcli, dbID, clairID, regID)
			panic(fmt.Errorf("starting registry container %s failed: %v", regConfig.config, err))
		}

		registryHelper, err = testutils.NewRegistryHelper(dcli, regConfig.username, regConfig.password, domain)
		if err != nil {
			panic(fmt.Errorf("creating registry helper %s failed: %v", regConfig.config, err))
		}

		flag.Parse()
		merr := m.Run()

		// remove registry
		if err := testutils.RemoveContainer(dcli, regID); err != nil {
			log.Printf("couldn't remove registry container %s: %v", regConfig.config, err)
		}

		if merr != 0 {
			testutils.RemoveContainer(dcli, dbID, clairID)
			fmt.Printf("testing config %s failed\n", regConfig.config)
			os.Exit(merr)
		}
	}

	// remove clair containers.
	if err := testutils.RemoveContainer(dcli, dbID, clairID); err != nil {
		log.Printf("couldn't remove clair containers: %v", err)
	}

	os.Exit(0)
}

func run(args ...string) (string, error) {
	prog := "./testreg" + exeSuffix
	// always add trust insecure, and the registry
	newargs := []string{args[0], "-d", "-k"}
	if len(args) > 1 {
		newargs = append(newargs, args[1:]...)
	}
	cmd := exec.Command(prog, newargs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
