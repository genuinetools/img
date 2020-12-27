# img

[![make-all](https://github.com/genuinetools/img/workflows/make%20all/badge.svg)](https://github.com/genuinetools/img/actions?query=workflow%3A%22make+all%22)
[![make-image](https://github.com/genuinetools/img/workflows/make%20image/badge.svg)](https://github.com/genuinetools/img/actions?query=workflow%3A%22make+image%22)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/img)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/img/total.svg?style=for-the-badge)](https://github.com/genuinetools/img/releases)

Standalone, daemon-less, unprivileged Dockerfile and OCI compatible
container image builder.

`img` is more cache-efficient than Docker and can also execute multiple build stages concurrently, 
as it internally uses [BuildKit](https://github.com/moby/buildkit)'s DAG solver.

The commands/UX are the same as `docker {build,tag,push,pull,login,logout,save}` so all you 
have to do is replace `docker` with `img` in your scripts, command line, and/or life.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Goals](#goals)
      - [Upstream Patches](#upstream-patches)
      - [Benchmarks](#benchmarks)
- [Installation](#installation)
    - [Binaries](#binaries)
    - [From Source](#from-source)
    - [Alpine Linux](#alpine-linux)
    - [Arch Linux](#arch-linux)
    - [Gentoo](#gentoo)
    - [Running with Docker](#running-with-docker)
  - [Running with Kubernetes](#running-with-kubernetes)
- [Usage](#usage)
  - [Build an Image](#build-an-image)
    - [Cross Platform](#cross-platform)
    - [Exporter Types](#exporter-types)
  - [List Image Layers](#list-image-layers)
  - [Pull an Image](#pull-an-image)
  - [Push an Image](#push-an-image)
  - [Tag an Image](#tag-an-image)
  - [Export an Image to Docker](#export-an-image-to-docker)
  - [Unpack an Image to a rootfs](#unpack-an-image-to-a-rootfs)
  - [Remove an Image](#remove-an-image)
  - [Disk Usage](#disk-usage)
  - [Prune and Cleanup the Build Cache](#prune-and-cleanup-the-build-cache)
  - [Login to a Registry](#login-to-a-registry)
  - [Logout from a Registry](#logout-from-a-registry)
  - [Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry)
- [How It Works](#how-it-works)
  - [Unprivileged Mounting](#unprivileged-mounting)
  - [High Level](#high-level)
  - [Low Level](#low-level)
  - [Snapshotter Backends](#snapshotter-backends)
    - [auto (default)](#auto-default)
    - [native](#native)
    - [overlayfs](#overlayfs)
    - [fuse-overlayfs](#fuse-overlayfs)
- [Contributing](#contributing)
- [Acknowledgements](#acknowledgements)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Goals

This a glorified cli tool built on top of
[buildkit](https://github.com/moby/buildkit). The goal of this project is to be
able to build container images as an unprivileged user.

Running unprivileged allows companies who use LDAP and other login mechanisms
to use `img` without needing root. This is very important in HPC environments
and academia as well.

Currently this works out of the box on a Linux machine if you install via 
the directions covered in [installing from binaries](#binaries). This
installation will ensure you have the correct version of `img` and also `runc`.

##### Upstream Patches

The ultimate goal is to also have this work inside a container. There are
patches being made to container runtimes and Kubernetes to make this possible. 
For the on-going work toward getting patches into container runtimes and
Kubernetes, see:

- [moby/moby#36644](https://github.com/moby/moby/pull/36644) **merged**
- [docker/cli#1347](https://github.com/docker/cli/pull/1347) **merged**
- [kubernetes/community#1934](https://github.com/kubernetes/community/pull/1934) **merged**
- [kubernetes/kubernetes#64283](https://github.com/kubernetes/kubernetes/pull/64283) **merged** 

The patches for runc has been merged into the upstream since `ecd55a4135e0a26de884ce436442914f945b1e76` (May 30, 2018).
The upstream BuildKit can also run in rootless mode since `65b526438b86a17cf35042011051ce15c8bfb92a` (June 1, 2018).

You might also be interested in reading: 
* [the original design doc](https://docs.google.com/document/d/1rT2GUSqDGcI2e6fD5nef7amkW0VFggwhlljrKQPTn0s/edit?usp=sharing)
* [a blog post on building images securely in Kubernetes](https://blog.jessfraz.com/post/building-container-images-securely-on-kubernetes/)

##### Benchmarks

If you are curious about benchmarks comparing various container builders, check
out [@AkihiroSuda's buildbench](https://github.com/AkihiroSuda/buildbench) 
[results](https://github.com/AkihiroSuda/buildbench/issues/1).


## Installation

You need to have `newuidmap` installed. On Ubuntu, `newuidmap` is provided by the `uidmap` package.

You also need to have `seccomp` installed. On Ubuntu, `seccomp` is provided by the `libseccomp-dev` package.

`runc` will be installed on start from an embedded binary if it is not already
available locally. If you would like to disable the embedded runc you can use `BUILDTAGS="seccomp
noembed"` while building from source with `make`. Or the environment variable
`IMG_DISABLE_EMBEDDED_RUNC=1` on execution of the `img` binary.

NOTE: These steps work only for Linux. Compile and run in a container 
(explained below) if you're on Windows or MacOS.

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/img/releases).

#### From Source

```bash
$ mkdir -p $GOPATH/src/github.com/genuinetools
$ git clone https://github.com/genuinetools/img $GOPATH/src/github.com/genuinetools/img
$ cd !$
$ make
$ sudo make install

# For packagers if you would like to disable the embedded `runc`, please use:
$ make BUILDTAGS="seccomp noembed"
```

#### Alpine Linux

There is an [APKBUILD](https://pkgs.alpinelinux.org/package/edge/community/x86_64/img).

```console
$ apk add img
```

#### Arch Linux

There is an [AUR build](https://aur.archlinux.org/packages/img/).

```console
# Use whichever AUR helper you prefer
$ yay -S img

# Or build from the source PKGBUILD
$ git clone https://aur.archlinux.org/packages/img.git
$ cd img
$ makepkg -si
```

#### Gentoo

There is an [ebuild](https://github.com/gentoo/gentoo/tree/master/app-emulation/img).

```console
$ sudo emerge -a app-emulation/img
```

#### Running with Docker

Docker image `r.j3ss.co/img` is configured to be executed as an unprivileged user with UID 1000 and it does not need `--privileged` since `img` v0.5.11.

```console
$ docker run --rm -it \
    --name img \
    --volume $(pwd):/home/user/src:ro \ # for the build context and dockerfile, can be read-only since we won't modify it
    --workdir /home/user/src \ # set the builder working directory
    --volume "${HOME}/.docker:/root/.docker:ro" \ # for credentials to push to docker hub or a registry
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \ # required by runc
    r.j3ss.co/img build -t user/myimage .
```

To enable PID namespace isolation (which disallows build containers to `kill(2)` the `img` process), you need to specify
`--privileged` so that build containers can mount `/proc` with unshared PID namespaces.
Note that even with `--privileged`, `img` works as an unprivileged user with UID 1000.

See [docker/cli patch](#upstream-patches) for how to allow mounting `/proc` without `--privileged`.

### Running with Kubernetes

Since `img` v0.5.11, you don't need to specify any `securityContext` for running `img` as a Kubernetes container.

However the following security annotations are needed:
```
container.apparmor.security.beta.kubernetes.io/img: unconfined
container.seccomp.security.alpha.kubernetes.io/img: unconfined
```

To enable PID namespace isolation, you need to set `securityContext.procMount` to `Unmasked` (or simply set
`securityContext.privileged` to `true`).
`securityContext.procMount` is available since Kubernetes 1.12 with Docker 18.06/containerd 1.2/CRI-O 1.12.

## Usage

Make sure you have user namespace support enabled. On some distros (Debian and
Arch Linux) this requires running `echo 1 > /proc/sys/kernel/unprivileged_userns_clone`.


```console
$ img -h
img -  Standalone, daemon-less, unprivileged Dockerfile and OCI compatible container image builder

Usage: img [OPTIONS] COMMAND [ARG...]

Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
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

### Build an Image

```console
$ img build -h
build -  Build an image from a Dockerfile

Usage: img build [OPTIONS] PATH

Flags:
      --build-arg list    Set build-time variables
      --cache-from list   Buildkit import-cache or Buildx cache-from specification
      --cache-to list     Buildx cache-to specification
  -f, --file string       Name of the Dockerfile (Default is 'PATH/Dockerfile')
  -h, --help              help for build
      --label list        Set metadata for an image
      --no-cache          Do not use cache when building the image
      --no-console        Use non-console progress UI
  -o, --output string     BuildKit output specification (e.g. type=tar,dest=build.tar)
      --platform list     Set platforms for which the image should be built
  -t, --tag list          Name and optionally a tag in the 'name:tag' format
      --target string     Set the target build stage to build

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

**Use just like you would `docker build`.**

```console
$ img build -t r.j3ss.co/img .
Building r.j3ss.co/img:latest
Setting up the rootfs... this may take a bit.
[+] Building 44.7s (16/16) FINISHED                                                        
 => local://dockerfile (Dockerfile)                                                   0.0s
 => => transferring dockerfile: 1.15kB                                                0.0s
 => local://context (.dockerignore)                                                   0.0s
 => => transferring context: 02B                                                      0.0s
 => CACHED docker-image://docker.io/tonistiigi/copy:v0.1.1@sha256:854cee92ccab4c6d63  0.0s
 => => resolve docker.io/tonistiigi/copy:v0.1.1@sha256:854cee92ccab4c6d63183d147389e  0.0s
 => CACHED docker-image://docker.io/library/alpine@sha256:e1871801d30885a610511c867d  0.0s
 => => resolve docker.io/library/alpine@sha256:e1871801d30885a610511c867de0d6baca7ed  0.0s
 => docker-image://docker.io/library/golang:1.10-alpine@sha256:98c1f3458b21f50ac2e58  5.5s
 => => resolve docker.io/library/golang:1.10-alpine@sha256:98c1f3458b21f50ac2e5896d1  0.0s
 => => sha256:866414f805391b58973d4e3d76e5d32ae51baecb1c93762c9751b9d6c5 126B / 126B  0.0s
 => => sha256:ae8dbf6f23bf1c326de78fc780c6a870bf11eb86b45a7dc567 308.02kB / 308.02kB  0.0s
 => => sha256:44ccce322b34208317d748e998212cd677c16f1a58c2ff5e59578c 3.86kB / 3.86kB  0.0s
 => => sha256:0d01df27c53e651ecfa5c689dafb8c63c759761a757cc37e30eccc5e3a 153B / 153B  0.0s
 => => sha256:ff3a5c916c92643ff77519ffa742d3ec61b7f591b6b7504599d95a 2.07MB / 2.07MB  0.0s
 => => sha256:4be696a8d726150ed9636ea7156edcaa9ba8293df1aae49f9e 113.26MB / 113.26MB  0.0s
 => => sha256:98c1f3458b21f50ac2e5896d14a644eadb3adcae5afdceac0cc9c2 2.04kB / 2.04kB  0.0s
 => => sha256:bb31085d5c5db578edf3d4e5541cfb949b713bb7018bbac4dfd407 1.36kB / 1.36kB  0.0s
 => => unpacking docker.io/library/golang:1.10-alpine@sha256:98c1f3458b21f50ac2e5896  5.4s
 => local://context                                                                   0.8s
 => => transferring context: 116.83MB                                                 0.8s
 => /bin/sh -c apk add --no-cache  bash  build-base  gcc  git  libseccomp-dev  linux  3.8s
 => copy /src-0 go/src/github.com/genuinetools/img/                                   1.5s
 => /bin/sh -c go get -u github.com/jteeuwen/go-bindata/...                           7.3s
 => /bin/sh -c make static && mv img /usr/bin/img                                    15.2s
 => /bin/sh -c git clone https://github.com/opencontainers/runc.git "$GOPATH/src/git  7.6s
 => /bin/sh -c apk add --no-cache  bash  git  shadow  shadow-uidmap  strace           2.3s
 => copy /src-0/img usr/bin/img                                                       0.5s
 => copy /src-0/runc usr/bin/runc                                                     0.4s
 => /bin/sh -c useradd --create-home --home-dir $HOME user  && chown -R user:user $H  0.4s
 => exporting to image                                                                1.5s
 => => exporting layers                                                               1.4s
 => => exporting manifest sha256:03e034afb839fe6399a271efc972da823b1b6297ea792ec94fa  0.0s
 => => exporting config sha256:92d033f9575176046db41f4f1feacc0602c8f2811f59d59f8e7b6  0.0s
 => => naming to r.j3ss.co/img:latest                                                 0.0s
Successfully built r.j3ss.co/img:latest
```

#### Cross Platform

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

### List Image Layers

```console
$ img ls -h
ls -  List images and digests.

Usage: img ls [OPTIONS]

Flags:
  -f, --filter list   Filter output based on conditions provided
  -h, --help          help for ls

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img ls
NAME                    SIZE            CREATED AT      UPDATED AT      DIGEST
jess/img:latest         1.534KiB        9 seconds ago   9 seconds ago   sha256:27d862ac32022946d61afbb91ddfc6a1fa2341a78a0da11ff9595a85f651d51e
jess/thing:latest       591B            30 minutes ago  30 minutes ago  sha256:d664b4e9b9cd8b3067e122ef68180e95dd4494fd4cb01d05632b6e77ce19118e
```

### Pull an Image

If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

```console
$ img pull -h
pull -  Pull an image or a repository from a registry.

Usage: img pull [OPTIONS] NAME[:TAG|@DIGEST]

Flags:
  -h, --help   help for pull

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img pull r.j3ss.co/stress
Pulling r.j3ss.co/stress:latest...
Snapshot ref: sha256:2bb7a0a5f074ffe898b1ef64b3761e7f5062c3bdfe9947960e6db48a998ae1d6
Size: 365.9KiB
```

### Push an Image

If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

```console
$ img push -h
push -  Push an image or a repository to a registry.

Usage: img push [OPTIONS] NAME[:TAG]

Flags:
  -h, --help                help for push
      --insecure-registry   Push to insecure registry

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img push jess/thing
Pushing jess/thing:latest...
Successfully pushed jess/thing:latest
```

### Tag an Image

```console
$ img tag -h
tag -  Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.

Usage: img tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]

Flags:
  -h, --help   help for tag

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img tag jess/thing jess/otherthing
Successfully tagged jess/thing as jess/otherthing
```

### Export an Image to Docker

```console
$ img save -h
save -  Save an image to a tar archive (streamed to STDOUT by default).

Usage: img save [OPTIONS] IMAGE [IMAGE...]

Flags:
      --format string   image output format (docker|oci) (default "docker")
  -h, --help            help for save
  -o, --output string   write to a file, instead of STDOUT

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img save jess/thing | docker load
6c3d70c8619c: Loading layer [==================================================>]  9.927MB/9.927MB                                      
7e336c441b5e: Loading layer [==================================================>]  5.287MB/5.287MB                                      
533fecff21a8: Loading layer [==================================================>]   2.56MB/2.56MB                                       
3db7019eac28: Loading layer [==================================================>]  1.679kB/1.679kB                                      
Loaded image: jess/thing
```

### Unpack an Image to a rootfs

```console
$ img unpack -h
unpack -  Unpack an image to a rootfs directory.

Usage: img unpack [OPTIONS] IMAGE

Flags:
  -h, --help            help for unpack
  -o, --output string   Directory to unpack the rootfs to. (defaults to rootfs/ in the current working directory)

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img unpack busybox
Successfully unpacked rootfs for busybox to: /home/user/rootfs
```

### Remove an Image

```console
$ img rm -h
rm -  Remove one or more images.

Usage: img rm [OPTIONS] IMAGE [IMAGE...]

Flags:
  -h, --help   help for rm

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

### Disk Usage

```console
$ img du -h
du -  Show image disk usage.

Usage: img du [OPTIONS]

Flags:
  -f, --filter list   Filter output based on conditions provided
  -h, --help          help for du

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

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

### Prune and Cleanup the Build Cache

```console
$ img prune -h
prune -  Prune and clean up the build cache.

Usage: img prune [OPTIONS]

Flags:
  -h, --help   help for prune

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

```console
$ img prune
ID                                                                      RECLAIMABLE     SIZE            DESCRIPTION
j1yil8bdz35eyxp0m17tggknd                                               true            5.08KiB         local source for dockerfile
je23wfyz2apii1au38occ8zag                                               true            52.95MiB        mount / from exec /bin/sh -c useradd --create-home...
sha256:74906c0186257f2897c5fba99e1ea87eb8b2ee0bb03b611f5e866232bfbf6739 true            2.238MiB        pulled from docker.io/tonistiigi/copy:v0.1.1@sha25...
vr2pvhmrt1sjs8n7jodesrvnz*                                              true            572.6MiB        mount / from exec /bin/sh -c git clone https://git...
afn0clz11yphlv6g8golv59c8                                               true            4KiB            local source for context
qx5yql370piuscuczutrnansv*                                              true            692.4MiB        mount / from exec /bin/sh -c make static && mv img...
uxocruvniojl1jqlm8gs3ds1e*                                              true            113.8MiB        local source for context
sha256:0b9cfed6a170b357c528cd9dfc104d8b404d08d84152b38e98c60f50d2ae718b true            1.449MiB        pulled from docker.io/tonistiigi/copy:v0.1.1@sha25...
vz0716utmnlmya1vhkojyxd4o                                               true            55.39MiB        mount /dest from exec copy /src-0/runc usr/bin/run...
a0om6hwulbf9gd2jfgmxsyaoa                                               true            646.5MiB        mount / from exec /bin/sh -c go get -u github.com/...
ys8y9ixi3didtbpvwbxuptdfq                                               true            641.2MiB        mount /dest from exec copy /src-0 go/src/github.co...
sha256:f64a552a56ce93b6e389328602f2cd830280fd543ade026905e69895b5696b7a true            1.234MiB        pulled from docker.io/tonistiigi/copy:v0.1.1@sha25...
05wxxnq6yu5nssn3bojsz2mii                                               true            52.4MiB         mount /dest from exec copy /src-0/img usr/bin/img
wlrp1nxsa37cixf127bh6w2sv                                               true            35.11MiB        mount / from exec /bin/sh -c apk add --no-cache  b...
wy0173xa6rkoq49tf9g092r4z                                               true            527.4MiB        mount / from exec /bin/sh -c apk add --no-cache  b...
Reclaimed:      4.148GiB
Total:          4.148GiB
```

### Login to a Registry

If you need to use self-signed certs with your registry, see 
[Using Self-Signed Certs with a Registry](#using-self-signed-certs-with-a-registry).

```console
$ img login -h
login -  Log in to a Docker registry.

Usage: img login [OPTIONS] [SERVER]

Flags:
  -h, --help              help for login
  -p, --password string   Password
      --password-stdin    Take the password from stdin
  -u, --user string       Username

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

### Logout from a Registry

```console
$ img logout -h
logout -  Log out from a Docker registry.

Usage: img logout [SERVER]

Flags:
  -h, --help   help for logout

Global Flags:
  -b, --backend string   backend for snapshots ([auto native overlayfs fuse-overlayfs]) (default "auto")
  -d, --debug            enable debug logging
  -s, --state string     directory to hold the global state (default "/home/user/.local/share/img")
```

### Using Self-Signed Certs with a Registry

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

The `auto` backend selects a backend based on what the current system supports,
preferring `overlayfs`, then `fuse-overlayfs`, then `native`.

#### native

The `native` backend creates image layers by simply copying files.
`copy_file_range(2)` is used when available.

#### overlayfs

The `overlayfs` backend uses the kernel's native overlayfs support. It requires
a kernel patch from Ubuntu to be unprivileged, see
[#22](https://github.com/genuinetools/img/issues/22).

#### fuse-overlayfs

The `fuse-overlayfs` backend provides overlay support without any kernel
patches. It requires a Linux kernel >= 4.18 and for
[fuse-overlayfs](https://github.com/containers/fuse-overlayfs) to be installed.


## Contributing

Please do! This is a new project and can use some love <3. Check out the [issues](https://github.com/genuinetools/img/issues).

The local directories are mostly re-implementations of `buildkit` interfaces to
be unprivileged.

## Acknowledgements

A lot of this is based on the work of [moby/buildkit](https://github.com/moby/buildkit). 
Thanks [@tonistiigi](https://github.com/tonistiigi) and
[@AkihiroSuda](https://github.com/AkihiroSuda)!
