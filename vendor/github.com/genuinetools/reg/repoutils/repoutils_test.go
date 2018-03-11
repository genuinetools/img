package repoutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker-ce/components/cli/cli/config"
	"github.com/docker/docker/api/types"
	"github.com/google/go-cmp/cmp"
)

func TestGetAuthConfig(t *testing.T) {
	configTestcases := []struct {
		name                         string
		username, password, registry string
		configdir                    string
		err                          error
		config                       types.AuthConfig
	}{
		{
			name:     "pass in all details",
			username: "jess",
			password: "password",
			registry: "r.j3ss.co",
			config: types.AuthConfig{
				Username:      "jess",
				Password:      "password",
				ServerAddress: "r.j3ss.co",
			},
		},
		{
			name:      "invalid config dir",
			configdir: "testdata/invalid",
			err:       errors.New("Loading config file failed: "),
			config:    types.AuthConfig{},
		},
		{
			name:      "empty config",
			configdir: "testdata/empty",
			config:    types.AuthConfig{},
		},
		{
			name:      "empty config with docker.io",
			registry:  "docker.io",
			configdir: "testdata/empty",
			config: types.AuthConfig{
				ServerAddress: DefaultDockerRegistry,
			},
		},
		{
			name:      "empty config with registry",
			registry:  "r.j3ss.co",
			configdir: "testdata/empty",
			config: types.AuthConfig{
				ServerAddress: "r.j3ss.co",
			},
		},
		{
			name:      "valid with multiple",
			registry:  "r.j3ss.co",
			configdir: "testdata/valid",
			config: types.AuthConfig{
				ServerAddress: "r.j3ss.co",
				Username:      "user",
				Password:      "blah\n",
			},
		},
		{
			name:      "valid with multiple and https:// prefix",
			registry:  "https://r.j3ss.co",
			configdir: "testdata/valid",
			config: types.AuthConfig{
				ServerAddress: "r.j3ss.co",
				Username:      "user",
				Password:      "blah\n",
			},
		},
		{
			name:      "valid with multiple and http:// prefix",
			registry:  "http://r.j3ss.co",
			configdir: "testdata/valid",
			config: types.AuthConfig{
				ServerAddress: "r.j3ss.co",
				Username:      "user",
				Password:      "blah\n",
			},
		},
		{
			name:      "valid with multiple and no https:// prefix",
			registry:  "reg.j3ss.co",
			configdir: "testdata/valid",
			config: types.AuthConfig{
				ServerAddress: "https://reg.j3ss.co",
				Username:      "joe",
				Password:      "otherthing\n",
			},
		},
		{
			name:      "valid with multiple and but registry not found",
			registry:  "otherreg.j3ss.co",
			configdir: "testdata/valid",
			config: types.AuthConfig{
				ServerAddress: "otherreg.j3ss.co",
			},
		},
		{
			name:      "valid and no registry passed",
			configdir: "testdata/singlevalid",
			config: types.AuthConfig{
				ServerAddress: "https://index.docker.io/v1/",
				Username:      "user",
				Password:      "thing\n",
			},
		},
	}

	for _, testcase := range configTestcases {
		if testcase.configdir != "" {
			// Set the config directory.
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("get working directory failed: %v", err)
			}
			config.SetDir(filepath.Join(wd, testcase.configdir))
		}

		cfg, err := GetAuthConfig(testcase.username, testcase.password, testcase.registry)
		if err != nil || testcase.err != nil {
			if err == nil || testcase.err == nil {
				t.Fatalf("%q: expected err (%v), got err (%v)", testcase.name, testcase.err, err)
			}
			if !strings.Contains(err.Error(), testcase.err.Error()) {
				t.Fatalf("%q: expected err (%v), got err (%v)", testcase.name, testcase.err, err)
			}
			continue
		}

		if diff := cmp.Diff(testcase.config, cfg); diff != "" {
			t.Errorf("%s: authconfig differs: (-got +want)\n%s", testcase.name, diff)
		}
	}
}

func TestGetRepoAndRef(t *testing.T) {
	imageTestcases := []struct {
		// input is the repository name or name component testcase
		input string
		// err is the error expected from Parse, or nil
		err error
		// repository is the string representation for the reference
		repository string
		// ref the reference
		ref string
	}{
		{
			input:      "alpine",
			repository: "alpine",
			ref:        "latest",
		},
		{
			input:      "docker:dind",
			repository: "docker",
			ref:        "dind",
		},
		{
			input: "",
			err:   reference.ErrNameEmpty,
		},
		{
			input:      "chrome@sha256:2a6c8ad38c41ae5122d76be59b34893d7fa1bdfaddd85bf0e57d0d16c0f7f91e",
			repository: "chrome",
			ref:        "sha256:2a6c8ad38c41ae5122d76be59b34893d7fa1bdfaddd85bf0e57d0d16c0f7f91e",
		},
	}

	for _, testcase := range imageTestcases {
		repo, ref, err := GetRepoAndRef(testcase.input)
		if err != nil || testcase.err != nil {
			if err == nil || testcase.err == nil {
				t.Fatalf("%q: expected err (%v), got err (%v)", testcase.input, testcase.err, err)
			}
			if err.Error() != testcase.err.Error() {
				t.Fatalf("%q: expected err (%v), got err (%v)", testcase.input, testcase.err, err)
			}
			continue
		}

		if testcase.repository != repo {
			t.Fatalf("%q: expected repo (%s), got repo (%s)", testcase.input, testcase.repository, repo)
		}

		if testcase.ref != ref {
			t.Fatalf("%q: expected ref (%s), got ref (%s)", testcase.input, testcase.ref, ref)
		}
	}
}
