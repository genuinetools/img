package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildShCmdJSONEntrypoint(t *testing.T) {
	name := "testbuildshcmdjsonentrypoint"

	runBuild(t, name, withDockerfile(`
    FROM busybox
    ENTRYPOINT ["echo"]
    CMD echo test
    `))
}

func TestBuildEnvironmentReplacementUser(t *testing.T) {
	name := "testbuildenvironmentreplacement"

	runBuild(t, name, withDockerfile(`
  FROM scratch
  ENV user foo
  USER ${user}
  `))
}

func TestBuildEnvironmentReplacementVolume(t *testing.T) {
	name := "testbuildenvironmentreplacement"

	volumePath := "/quux"
	if runtime.GOOS == "windows" {
		volumePath = "c:/quux"
	}

	runBuild(t, name, withDockerfile(`
  FROM busybox
  ENV volume `+volumePath+`
  VOLUME ${volume}
  `))
}

func TestBuildEnvironmentReplacementExpose(t *testing.T) {
	name := "testbuildenvironmentreplacement"

	runBuild(t, name, withDockerfile(`
  FROM scratch
  ENV port 80
  EXPOSE ${port}
  ENV ports "  99   100 "
  EXPOSE ${ports}
  `))
}

func TestBuildEnvironmentReplacementWorkdir(t *testing.T) {
	name := "testbuildenvironmentreplacement"

	runBuild(t, name, withDockerfile(`
  FROM busybox
  ENV MYWORKDIR /work
  RUN mkdir ${MYWORKDIR}
  WORKDIR ${MYWORKDIR}
  `))
}

func TestBuildFromScratch(t *testing.T) {
	name := "testbuildfromscratch"

	runBuild(t, name, withDockerfile(`
  FROM scratch
  COPY . .
  `))
}

func TestBuildDockerfileNotInContext(t *testing.T) {
	name := "testbuilddockerfilenotincontext"

	run(t, "build", "-t", name, "-f", "testdata/Dockerfile.test-build-dockerfile-not-in-context", "types")
}

func TestBuildDockerfileNotInContextRoot(t *testing.T) {
	name := "testbuilddockerfilenotincontextroot"

	run(t, "build", "-t", name, "-f", "testdata/Dockerfile.test-build-dockerfile-not-in-context", ".")
}

// Make sure the client exits with the correct exit code.
// https://github.com/genuinetools/img/issues/101
func TestBuildDockerfileFailing(t *testing.T) {
	name := "testbuilddockerfilefailing"

	args := []string{"build", "-t", name, "-f", "testdata/Dockerfile.test-build-failing", "."}
	out, err := doRun(args, nil)
	if err == nil {
		t.Logf("img %v should have failed but did not: %s", args, out)
		t.FailNow()
	}
}

// Using apt requires subuid, subgid, setgroups, and networking to be enabled.
// https://github.com/genuinetools/img/issues/96
func TestBuildAPT(t *testing.T) {
	name := "testbuildapt"

	runBuild(t, name, withDockerfile(`
  FROM debian:9-slim
  RUN apt update
  `))
}

func TestBuildLabels(t *testing.T) {
	name := "testbuildlabels"

	args := []string{"build", "-t", name, "--label", "cli-label-1=cli1", "--label", "cli-label-2=cli2", "-"}
	_, err := doRun(args, withDockerfile(`
  FROM scratch as builder
  LABEL stage "builder"
  FROM scratch
  LABEL stage "final"
  `))

	if err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}
}

func TestBuildMultipleTags(t *testing.T) {
	names := []string{"testbuildmultipletags", "testbuildmultipletags:v1", "testbuildmultipletagsv1"}
	args := []string{"build"}

	for _, name := range names {
		args = append(args, "-t", name)
	}
	args = append(args, "-")

	_, err := doRun(args, withDockerfile(`
  FROM scratch
  `))

	if err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}
}

func TestBuildMultiplePlatforms(t *testing.T) {
	args := []string{"build", "--platform", "amd64", "--platform", "linux/arm64,linux/arm/v7", "-t", "testbuildplatforms", "-"}

	_, err := doRun(args, withDockerfile(`
  FROM alpine
  `))

	if err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}
}

func TestBuildContextFirstInCommand(t *testing.T) {
	args := []string{"build", "-", "-t", "testbuildargsfirst"}

	_, err := doRun(args, withDockerfile(`
  FROM busybox
  `))

	if err != nil {
		t.Logf("img %v failed unexpectedly: %v", args, err)
		t.FailNow()
	}
}

func TestBuildOutputLocal(t *testing.T) {

	tmpd, err := ioutil.TempDir("", "img-buildoutputlocal")
	if err != nil {
		t.Fatalf("creating temporary directory for build output failed: %v", err)
	}
	defer os.RemoveAll(tmpd)
	rootfs := filepath.Join(tmpd, "rootfs")

	args := []string{"build", "-", "-o", fmt.Sprintf("type=local,dest=%s", rootfs)}
	_, err = doRun(args, withDockerfile(`
	FROM busybox
	RUN touch /imgout
	`))
	if err != nil {
		t.Fatalf("img %v failed unexpectedly: %v", args, err)
	}

	// Make sure the image actually is unpacked in the directory.
	file := filepath.Join(rootfs, "imgout")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Fatalf("expected file at %q to exist but it did not", file)
	}
}

func testBuildOutputArchive(otype string, t *testing.T) {

	tmpd, err := ioutil.TempDir("", "img-buildoutput"+otype)
	if err != nil {
		t.Fatalf("creating temporary directory for build output failed: %v", err)
	}
	defer os.RemoveAll(tmpd)
	archive := filepath.Join(tmpd, "output.tar")

	args := []string{"build", "-", "-o", fmt.Sprintf("type=%s,dest=%s", otype, archive)}
	_, err = doRun(args, withDockerfile(`
	FROM busybox
	`))
	if err != nil {
		t.Fatalf("img %v failed unexpectedly: %v", args, err)
	}

	// Make sure the output is a valid tar archive.
	f, err := os.Open(archive)
	if err != nil {
		t.Fatalf("could not open output archive at %q: %s", archive, err)
	}
	defer f.Close()
	tr := tar.NewReader(f)
	if _, err = tr.Next(); err != nil {
		t.Fatalf("could not read first item in %s archive: %s", otype, err)
	}
}

func TestBuildOutputTar(t *testing.T) {
	testBuildOutputArchive("tar", t)
}

func TestBuildOutputDocker(t *testing.T) {
	testBuildOutputArchive("docker", t)
}

func TestBuildOutputOCI(t *testing.T) {
	testBuildOutputArchive("oci", t)
}

func TestBuildOutputTarStdout(t *testing.T) {

	args := []string{"build", "-", "-o", "type=tar"}

	// modified doRun() function to capture stdout seperately
	doRunStdout := func(args []string, stdin io.Reader) ([]byte, error) {
		prog := "./testimg" + exeSuffix

		newargs := []string{args[0], "--state", testStateDir}
		newargs = append(newargs, args[1:]...)

		cmd := exec.Command(prog, newargs...)
		if stdin != nil {
			cmd.Stdin = stdin
		}
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("Error running %s: %v", strings.Join(newargs, " "), err)
		}
		return out, nil
	}

	out, err := doRunStdout(args, withDockerfile(`
	FROM busybox
	`))
	if err != nil {
		t.Fatalf("img %v failed unexpectedly: %v", args, err)
	}

	// try to read tar entry from stdout
	tr := tar.NewReader(bytes.NewReader(out))
	if _, err = tr.Next(); err != nil {
		t.Logf("could not read tar archive from stdout: %s", err)
		t.Logf("first 256 bytes: %s", out[:256])
		t.FailNow()
	}
}

func TestBuildOutputImage(t *testing.T) {
	name := "testbuildoutputimage"

	args := []string{"build", "-", "-o", fmt.Sprintf("type=image,name=%s", name)}
	_, err := doRun(args, withDockerfile(`
	FROM busybox
	`))
	if err != nil {
		t.Fatalf("img %v failed unexpectedly: %v", args, err)
	}
}

func TestBuildOutputImageFailing(t *testing.T) {
	name := "testbuildoutputimagefailing"

	args := []string{"build", "-", "-o", fmt.Sprintf("type=image,dest=%s", name)}
	out, err := doRun(args, withDockerfile(`
	FROM busybox
	`))
	if err == nil {
		t.Fatalf("img %v should have failed but did not: %s", args, out)
	}
}
