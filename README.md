# img

[![Travis CI](https://img.shields.io/travis/genuinetools/img.svg?style=for-the-badge)](https://travis-ci.org/genuinetools/img)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/img)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/img/total.svg?style=for-the-badge)](https://github.com/genuinetools/img/releases)

Simple, standalone, daemon-less, unprivileged Dockerfile and OCI-compatible container image builder.

`img` is a simple CLI tool built on top of [buildkit](https://github.com/moby/buildkit).

The commands/UX are the same as `docker {build,tag,push,pull,login,logout,save}` so all you 
have to do is replace `docker` with `img` in your scripts, command line, CI/CD, and/or life.

**Table of Contents**

<!-- toc -->

- [Goals](#goals)
  * [Additional Reading](#additional-reading)
- [Getting Started](#getting-started)
  * [Run In Docker](#run-in-docker)
  * [Running with Kubernetes](#running-with-kubernetes)
  * [Mac or Windows Installation](#mac-or-windows-installation)
  * [Linux Installation](#linux-installation)
- [CLI Reference](#cli-reference)
  * [`build`](#build)
  * [`du`](#du)
  * [`login`](#login)
  * [`logout`](#logout)
  * [`ls`](#ls)
  * [`prune`](#prune)
  * [`pull`](#pull)
  * [`push`](#push)
  * [`rm`](#rm)
  * [`save`](#save)
  * [`tag`](#tag)
  * [`unpack`](#unpack)
- [Common Issues](#common-issues)
- [How It Works](#how-it-works)
  * [Unprivileged Mounting](#unprivileged-mounting)
  * [High Level](#high-level)
  * [Low Level](#low-level)
  * [Snapshotter Backends](#snapshotter-backends)
- [Contributing](#contributing)
  * [Quick-Start With Docker](#quick-start-with-docker)
- [Acknowledgements](#acknowledgements)

<!-- tocstop -->

## Goals

The key goals of Img are:

* **Least-Privileged.** Build containers without requiring root and utilizing as few permissions as possible.
* **Docker CLI Compatibility.** Don't change your development workflow, provide a drop-in replacement for docker for key functionality.
* **Dockerfile Compatibility.** Use the same Dockerfile you know and love to develop locally, and in your CI/CD builds.
* **Single Process.** Provide the simplest possible way to build without daemons or complex architectures. 

### Additional Reading

You might also be interested in reading:
 
* [the original design doc](https://docs.google.com/document/d/1rT2GUSqDGcI2e6fD5nef7amkW0VFggwhlljrKQPTn0s/edit?usp=sharing)
* [A blog post on building images securely in Kubernetes](https://blog.jessfraz.com/post/building-container-images-securely-on-kubernetes/) 
* [Benchmarks comparing various container builders](https://github.com/AkihiroSuda/buildbench/issues/1) by @AkihiroSuda.


## Getting Started

Img can be installed on any Linux distribution, or run via Docker on Windows or Mac. Img requires [runc](https://github.com/opencontainers/runc) and thus only 
supports Linux natively.

### Run In Docker

A prebuilt docker image is provided to run img via Docker: `r.j3ss.co/img`. This image is configured to be executed as 
an unprivileged user with UID 1000 and it does not need `--privileged` since `img` v0.5.7, but please note the security
options provided below.

#### Example Build

The following runs builds an image in an unprivileged container. This demonstrates that we are able to build images
within a container. The example below mounts the current directory as a volume, and also mounts docker credentials.

```console
$ docker run --rm -it \
    --name img \
    --volume $(pwd):/home/user/src:ro \
    --workdir /home/user/src \
    --volume "${HOME}/.docker:/root/.docker:ro" \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    --security-opt systempaths=unconfined \
    r.j3ss.co/img build -t user/myimage .
```

#### Run Interactively

Instead of directly calling img, you can enter a shell prompt to test out some of the capabilities of `img`. 

```console
$ docker run --rm -it \
    --name img \
    --volume $(pwd):/home/user/src:ro \
    --workdir /home/user/src \
    --volume "${HOME}/.docker:/root/.docker:ro" \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    --security-opt systempaths=unconfined \
    --entrypoint sh r.j3ss.co/img 
```

This will open a shell prompt where you can run `img` commands. Your current directory is also mounted as a volume, so 
you can also run a build your own project.

```console
$ img build -t user/myimage .
```

#### PID Namespace Isolation

To enable PID namespace isolation (which disallows build containers to `kill(2)` the `img` process), you need to specify
`--privileged` so that build containers can mount `/proc` with unshared PID namespaces.
Note that even with `--privileged`, `img` works as an unprivileged user with UID 1000.

### Running with Kubernetes

Since `img` v0.5.7, you don't need to specify any `securityContext` for running `img` as a Kubernetes container.

However the following security annotations are needed:
```
container.apparmor.security.beta.kubernetes.io/img: unconfined
container.seccomp.security.alpha.kubernetes.io/img: unconfined
```

To enable PID namespace isolation, you need to set `securityContext.procMount` to `Unmasked` (or simply set
`securityContext.privileged` to `true`).
`securityContext.procMount` is available since Kubernetes 1.12 with Docker 18.06/containerd 1.2/CRI-O 1.12.


### Mac or Windows Installation

To utilize img on Mac or Windows, install Docker for Desktop, and then utilize the Run In Docker instructions above.

### Linux Installation

#### Prerequisites

The following requirements must be met:

1. `newuidmap`. On Ubuntu, `newuidmap` is provided by the `uidmap` package.
2. `seccomp`. On Ubuntu, `seccomp` is provided by the `libseccomp-dev` package.
3. `runc` (optional). An embedded runc binary is provided within img if one is not available locally.
4. User namespace support enabled. On some distros (Debian and Arch Linux) this requires running `echo 1 > /proc/sys/kernel/unprivileged_userns_clone`.

##### Disable Embedded Runc

`runc` will be installed on start from an embedded binary if it is not already
available locally. If you would like to disable the embedded runc you can use `BUILDTAGS="seccomp
noembed"` while building from source with `make`. Or the environment variable
`IMG_DISABLE_EMBEDDED_RUNC=1` on execution of the `img` binary.

#### Binary Installation

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/img/releases).
This installation will ensure you have the correct version of `img` and also `runc`.

#### Install From Source

A build environment suitable for installing from source is provided in the [Dockerfile.dev](Dockerfile.dev) file. Ensure
system [prerequisites](#prerequisites) are met.

```bash
$ mkdir -p $GOPATH/src/github.com/genuinetools
$ git clone https://github.com/genuinetools/img $GOPATH/src/github.com/genuinetools/img
$ cd !$
$ make
$ sudo make install

# For packagers if you would like to disable the embedded `runc`, please use:
$ make BUILDTAGS="seccomp noembed"
```

#### Linux Distribution Packages

##### Alpine Linux

There is an [APKBUILD](https://pkgs.alpinelinux.org/package/edge/testing/x86_64/img).

```console
$ apk add img --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing
```

##### Arch Linux

There is an [AUR build](https://aur.archlinux.org/packages/img/).

```console
# Use whichever AUR helper you prefer
$ yay -S img

# Or build from the source PKGBUILD
$ git clone https://aur.archlinux.org/packages/img.git
$ cd img
$ makepkg -si
```

##### Gentoo

There is an [ebuild](https://github.com/gentoo/gentoo/tree/master/app-emulation/img).

```console
$ sudo emerge -a app-emulation/img
```

## CLI Reference

Img provides a `-h`, or `--help` flag to guide usage of the CLI and any commands. Img provides a subset of the most
important commands for building images found in the [Docker CLI].

```console
$ img -h
img -  Standalone, daemon-less, unprivileged Dockerfile and OCI compatible container image builder

Usage: img [OPTIONS] COMMAND [ARG...]

Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -h, --help             help for img
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
  -v, --version          Print version information and quit

Commands:
  build       Build an image from a Dockerfile
  du          Show image disk usage.
  help        Help about any command
  login       Log in to a Docker registry.
  logout      Log out from a Docker registry.
  ls          List images and digests.
  prune       Prune and clean up the build cache.
  pull        Pull an image or a repository from a registry.
  push        Push an image or a repository to a registry.
  rm          Remove one or more images.
  save        Save an image to a tar archive (streamed to STDOUT by default).
  tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.
  unpack      Unpack an image to a rootfs directory.
  version     Show the version information.

Use "img [command] --help" for more information about a command.
```

### `build`

Build an image from a Dockerfile. **Use just like you would `docker build`.**

#### Usage

```console
img build [OPTIONS] PATH
```

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--build-arg` | | Set build-time variables |
| `--file , -f` | `PATH/Dockerfile` | Name of the Dockerfile |
| `--label` | | Set metadata for an image |
| `--no-cache`| | Do not use cache when building the image |
| `--no-console`| | Use non-console progress UI |
| `--output , -o`| | BuildKit output [specification](#exporter-types) (e.g. type=tar,dest=build.tar) |
| `--platform`| | Set [platforms](#cross-platform-builds) for which the image should be built |
| `--tag , -t`| | Name and optionally a tag in the `name:tag` format |
| `--target   `| | Set the target build stage to build |

#### Cross Platform Builds

`img` and the underlying `buildkit` library support building containers for arbitrary platforms (OS and architecture combinations). In `img` this can be achieved using the `--platform` option, but note that
using the `RUN` command during a build requires installing support for the desired platform, and any `FROM` images used must exist for the target platform as well.

Some common platforms include:
* linux/amd64
* linux/arm64
* linux/arm/v7
* linux/arm/v6
* linux/s390x
* linux/ppc64le
* darwin/amd64
* windows/amd64

If you use multiple `--platform` options for the same build, they will be included into a [manifest](https://docs.docker.com/engine/reference/commandline/manifest/) and should work for the different platforms built for.

The most common way to get `RUN` working in cross-platform builds is to install an emulator such as QEMU on the host system (static bindings are recommended to avoid shared library loading issues). To properly use the emulator inside the build environment, the kernel [binfmt_misc](https://www.kernel.org/doc/html/latest/admin-guide/binfmt-misc.html) parameters must be set with the following flags: `OCF`.
You can check the settings in `/proc` to ensure they are set correctly.
```console
$ cat /proc/sys/fs/binfmt_misc/qemu-arm | grep flags
flags: OCF
```

On Debian/Ubuntu the above should be available with the `qemu-user-static` package >= `1:2.12+dfsg-3`

NOTE: cross-OS builds are slightly more complicated to get `RUN` commands working, but follow from the same principle.

#### Exporter Types

[bkoutputs]: https://github.com/moby/buildkit/blob/master/README.md#output

`img` can also use buildkit's [exporter types][bkoutputs] directly to output the
resulting image to a Docker-type bundle or a rootfs tar without saving the image
itself locally. Builds will still benefit from caching.

The output type and destination are specified with the `--output` flag. The list
of valid output specifications includes:

| flag | description |
|------------|-------------|
| `-o type=tar,dest=rootfs.tar` | export rootfs of target image to a tar archive |
| `-o type=tar` | output a rootfs tar to stdout, for use in piped commands |
| `-o type=docker,dest=image.tar` | save a Docker-type bundle of the image |
| `-o type=oci,dest=image.tar` | save an OCI-type bundle of the image |
| `-o type=local,dest=rootfs/` | export the target image to this directory |
| `-o type=image,name=r.j3ss.co/img` | build and tag an image and store it locally

When used in conjunction with a Dockerfile which has a final `FROM scratch` stage and
only copies files of interest from earlier stages with `COPY --from=...`, this can be
utilized to output arbitrary build artifacts for example.

### `du`

List images and disk usage.

#### Options

```console
  -f, --filter list   Filter output based on conditions provided
```

#### Example

```console
$ img du 
ID                                                                      RECLAIMABLE     SIZE            DESCRIPTION
sha256:d9a48086f223d28a838263a6c04705c8009fab1dd67cc82c0ee821545de3bf7c true            911.8KiB        pulled from docker.io/tonistiigi/copy@sha256:476e0a67a1e4650c6adaf213269a2913deb7c52cbc77f954026f769d51e1a14e
7ia86xm2e4hzn2u947iqh9ph2                                               true            203.2MiB        mount /dest from exec copy /src-0 /dest/go/src/github.com/genuinetools/img
...
sha256:9f131fba0383a6aaf25ecd78bd5f37003e41a4385d7f38c3b0cde352ad7676da true            958.6KiB        pulled from docker.io/library/golang:alpine@sha256:a0045fbb52a7ef318937e84cf7ad3301b4d2ba6cecc2d01804f428a1e39d1dfc
sha256:c4151b5a5de5b7e272b2b6a3a4518c980d6e7f580f39c85370330a1bff5821f1 true            472.3KiB        pulled from docker.io/tonistiigi/copy@sha256:476e0a67a1e4650c6adaf213269a2913deb7c52cbc77f954026f769d51e1a14e
sha256:ae4ecac23119cc920f9e44847334815d32bdf82f6678069d8a8be103c1ee2891 true            148.9MiB        pulled from docker.io/library/debian:buster@sha256:a7789365b226786a0cb9e0f142c515f9f2ede7164a6f6be4a1dc4bfe19d5ec9c
bkrjrzv3nvp7lvzd5cw9vzut7*                                              true            4.879KiB        local source for dockerfile
sha256:db193011cbfc238d622d65c4099750758df83d74571e8d7498392b17df381207 true            467.2MiB        pulled from docker.io/library/golang:alpine@sha256:a0045fbb52a7ef318937e84cf7ad3301b4d2ba6cecc2d01804f428a1e39d1dfc
wn4m5i5swdcjvt1ud5bvtr75h*                                              true            4.204KiB        local source for dockerfile
Reclaimable:    1.08GiB
Total:          1.08GiB
```

### `login`

Login to a registry. Img authentication works just like docker to login to repositories.

If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

#### Usage

```bash
img login [OPTIONS] [SERVER]
```

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--password , -p` | | Password |
| `--password-stdin` | | Take the password from stdin |
| `--user , -u` | | Username |

### `logout`

Removes credentials for a registry.

#### Usage

```console
img logout [SERVER]
```

### `ls`

List all the image layers stored in the backend.

#### Usage

```console
img ls [OPTIONS]
```

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--filter, -f` | | Filter output based on conditions provided |


### `prune`

Remove unused and dangling layers in the cache to reclaim space. Perform a `prune` after an `rm` to cleanup old images.

#### Usage

```console
img prune [OPTIONS]
```

### `pull`

Pull an Image from a registry. If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

#### Usage

```console
img pull [OPTIONS] NAME[:TAG|@DIGEST]
```

### `push`

Push an Image to a registry. If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

#### Usage

```console
img push [OPTIONS] NAME[:TAG]
```

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--insecure-registry` | | Push to insecure registry |

### `rm`

Remove images from the image store.

#### Usage

```console
img rm [OPTIONS] IMAGE [IMAGE...]
```

### `save`

Save an image to a tar archive (streamed to STDOUT by default). Provide an `--output` file to write to a file.

#### Usage

```console
img save [OPTIONS] IMAGE [IMAGE...]
```

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--format` | `docker`| image output format (`docker` | `oci`) |
| `--output , -o` | | Output file, instead of STDOUT |

#### Examples

##### Save Image via STDOUT

Send an image from img's backend to docker.

```console
$ img save jess/thing | docker load
6c3d70c8619c: Loading layer [==================================================>]  9.927MB/9.927MB                                      
7e336c441b5e: Loading layer [==================================================>]  5.287MB/5.287MB                                      
533fecff21a8: Loading layer [==================================================>]   2.56MB/2.56MB                                       
3db7019eac28: Loading layer [==================================================>]  1.679kB/1.679kB                                      
Loaded image: jess/thing
```


### `tag`

Create a tag that refers to the source image.

#### Usage


```console
$ img tag jess/thing jess/otherthing
Successfully tagged jess/thing as jess/otherthing
```


### `unpack`

Unpack the contents of an image to a rootfs. Provide an `--output` to specify where to unpack to, otherwise it saves the
image to `rootfs/` in the current directory.

#### Options

| Name, shorthand | Default | Description |
| --- | --- | --- |
| `--output , -o` | `$(pwd)/rootfs`| Directory to unpack the rootfs to. |

#### Example

```console
$ img unpack busybox
Successfully unpacked rootfs for busybox to: /home/user/rootfs
```

## Common Issues

#### Using Self-Signed Certs with a Registry

We do not allow users to pass all the custom certificate flags on commands
because it is unnecessarily messy and can be handled through Linux itself.
Which we believe is a better user experience than having to pass three
different flags just to communicate with a registry using self-signed or
private certificates.

Below are instructions on adding a self-signed or private certificate to your
trusted ca-certificates on Linux.

Make sure you have the package `ca-certificates` installed.

Copy the public half of your CA certificate (the one user to sign the CSR) into
the CA certificate directory (as root):

```console
$ cp cacert.pem /usr/share/ca-certificates
```

Rebuild the directory with your certificate included, run as root:

```console
# On debian, this will bring up a menu.
# Select the ask option, scroll to the certificate you are adding,
# 	mark it for inclusion, and select ok.
$ dpkg-reconfigure ca-certificates

# On other distros...
$ update-ca-certificates
```


## How It Works

### Unprivileged Mounting

To mount a filesystem without root accsess, `img` automatically invokes 
[`newuidmap(1)`](http://man7.org/linux/man-pages/man1/newuidmap.1.html)/[`newgidmap(1)`](http://man7.org/linux/man-pages/man1/newgidmap.1.html) 
SUID binaries to prepare SUBUIDs/SUBGIDs, which is typically required by `apt`.

Make sure you have sufficient entries (typically `>=65536`) in your 
`/etc/subuid` and `/etc/subgid`.

### High Level

<img src="contrib/how-it-works-high-level.png" width=300 />

### Low Level

<img src="contrib/how-it-works-low-level.png" width=300 />

### Snapshotter Backends

#### auto (default)

The `auto` backend is resolved into either `native` or `overlayfs`, depending on
the availability of `overlayfs` on the system.

#### native

The `native` backends creates image layers by simply copying files.
`copy_file_range(2)` is used when available.

#### overlayfs

You can also use `overlayfs` 
backend, but that requires a kernel patch from Ubuntu to be unprivileged, 
see [#22](https://github.com/genuinetools/img/issues/22).


## Contributing

Please do! This is a new project and can use some love <3. Check out the [issues](https://github.com/genuinetools/img/issues).

The local directories are mostly re-implementations of `buildkit` interfaces to
be unprivileged.

### Quick-Start With Docker

A [Dockerfile](Dockerfile.dev) is provided as a build environment for this project. This is a simple way to begin
contributing for all users without modifying local system versions, or if running on a Mac or Windows machine but need a 
Linux environment to build and test `img`.

Utilize the same security options present in the [Run in Docker](#run-in-docker) section when running this container.

The steps below will provide you with an environment with all the correct prerequisites installed. Since this is an 
Ubuntu image, you can augment this image with whatever development tools you might need. This is a simple way to get a 
basic development environment up and running.

1. Clone and `cd` into the `img` directory.
2. Build the development image with Docker. This is an Ubuntu-based image.
    ```bash
    $ docker build -t img.dev -f Dockerfile.dev .
    ```
3. Run the image via Docker, mounting the `img` filesystem. You can modify files and the running container will see updates.
   ```bash
   $ docker run --rm -it \
         --name img \
         --volume $(pwd):/home/user/img \
         --workdir /home/user/img \
         --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
         --security-opt systempaths=unconfined \
         img.dev
   ```
4. Run `make` to build. This will build an `img` binary in the current directory. You can also explore the other 
   targets available in the [Makefile](Makefile) or [basic.mk](basic.mk).
   ```bash
   $ make
   ```
5. Test the built binary. Since we are in the `img` project, we can test building `img` with it's [Dockerfile](Dockerfile)!
   ```bash
   $ ./img build -t test .
   ```
6. Alright! You've built img (twice!) and can start contributing.

Since your local filesystem is mounted in the container, you can use any IDE or text editor you are comfortable with on 
your host system, and run builds within the dev container.

## Acknowledgements

A lot of this is based on the work of [moby/buildkit](https://github.com/moby/buildkit). 
Thanks [@tonistiigi](https://github.com/tonistiigi) and
[@AkihiroSuda](https://github.com/AkihiroSuda)!
