package main

import (
	"runtime"
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

// apt requires subuid, subgid, setgroups, and networking to be enabled.
// https://github.com/genuinetools/img/issues/96
func TestBuildAPT(t *testing.T) {
	name := "testbuildapt"

	runBuild(t, name, withDockerfile(`
  FROM debian:9-slim
  RUN apt update
  `))
}
