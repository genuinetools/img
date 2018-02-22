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
