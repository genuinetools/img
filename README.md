# img

[![Build Status](https://travis-ci.org/jessfraz/img.svg?branch=master)](https://travis-ci.org/jessfraz/img)

Standalone, daemon-less, unprivileged Dockerfile and OCI compatible
container image builder.

`img` is more cache-efficient than Docker and can also execute multiple build stages concurrently, 
as it internally uses [BuildKit](https://github.com/moby/buildkit)'s DAG solver.

The commands/UX are the same as `docker {build,push,pull,login}` so all you 
have to do is replace `docker` with `img` in your scripts, command line, and/or life.

Currently you can run it as unprivileged if you follow the instructions in the
[unprivileged mounting](#unprivileged-mounting) section below.

You might also be interested in reading the 
[original design doc](https://docs.google.com/document/d/1rT2GUSqDGcI2e6fD5nef7amkW0VFggwhlljrKQPTn0s/edit?usp=sharing).

#### Snapshotter Backends

The default backend is currently set to `naive` and requires privileges, but 
it can be made unprivileged and that work is being done, see the 
[unprivileged mounting](#unprivileged-mounting) section below.
It is a lot more stable than the `fuse` backend. You can also use `overlayfs` 
backend, but that requires a kernel patch from Ubuntu to be unprivileged, 
see [#22](https://github.com/jessfraz/img/issues/22).

The `fuse` backend runs completely in userspace. It is a bit buggy and a work
in progress so hang tight.

#### Unprivileged Mounting

To mount a filesystem without root access you need to do it from a mount and
user namespace.

Make sure you have user namespace support enabled. On some distros (Debian and
Arch Linux) this requires running `echo 1 > /proc/sys/kernel/unprivileged_ns_clone`.

You also need a version of `runc` with the patches from
[opencontainers/runc#1688](https://github.com/opencontainers/runc/pull/1688).

Example:

```console
# unshare a mountns and userns 
# and remap the user inside the namespaces to your current user
$ unshare -m -U --map-root-user

# then you can run img
$ img build -t user/myimage .
```

Note that `unshare -m -U --map-root-user` does not make use of [`subuid(5)`](http://man7.org/linux/man-pages/man5/subuid.5.html)/[`subgid(5)`](http://man7.org/linux/man-pages/man5/subgid.5.html), and also, it disables [`setgroups(2)`](http://man7.org/linux/man-pages/man2/setgroups.2.html), which is typically required by `apt`.

So we might want to use [`newuidmap(1)`](http://man7.org/linux/man-pages/man1/newuidmap.1.html)/[`newgidmap(1)`](http://man7.org/linux/man-pages/man1/newgidmap.1.html) SUID binaries to enable these features. See [opencontainers/runc#1692](https://github.com/opencontainers/runc/pull/1692) and [opencontainers/runc#1693](https://github.com/opencontainers/runc/pull/1693).

If depending on these SUID binaries is problematic, we could use ptrace hacks such as PRoot, although its performance overhead is not negligible. ([#15](https://github.com/jessfraz/img/issues/15) and [AkihiroSuda/runrootless](https://github.com/AkihiroSuda/runrootless))

For the on-going work toward integrating runc with these patches to `buildkit`, please refer to [moby/buildkit#252](https://github.com/moby/buildkit/issues/252#issuecomment-359696630) 
and [AkihiroSuda/buildkit_poc@511c7e71](https://github.com/AkihiroSuda/buildkit_poc/commit/511c7e71156fb349dca52475d8c0dc0946159b7b).


#### Goals

This project is not trying to reinvent `buildkit`. If anything all changes and
modifications for `buildkit` without privileges are being done upstream. The
goal of this project is to in the future just be a glorified cli tool on top of
`buildkit`.

**Table of Contents**

* [Installation](#installation)
    - [Binaries](#binaries)
    - [Via Go](#via-go)
    - [Running with Docker](#running-with-docker)
* [Usage](#usage)
    + [Build an Image](#build-an-image)
    + [List Image Layers](#list-image-layers)
    + [Pull an Image](#pull-an-image)
    + [Push an Image](#push-an-image)
    + [Disk Usage](#disk-usage)
    + [Login to a Registry](#login-to-a-registry)
* [Contributing](#contributing)
* [Acknowledgements](#acknowledgements)

## Installation

You need to have `runc` installed.

For the FUSE backend, you will also need `fusermount` installed.

#### Binaries

- **linux** [amd64](https://github.com/jessfraz/img/releases/download/v0.2.2/img-linux-amd64)

#### Via Go

```bash
$ go get github.com/jessfraz/img
```

#### Running with Docker

```console
$ docker run --rm -it \
    --name img \
    --volume /tmp/state:/root/.img \
    --volume $(pwd):/src \
    --workdir /src \
    --privileged \
    --volume "${HOME}/.docker:/root/.docker:ro" \
    r.j3ss.co/img build -t user/myimage .
```

## Usage

```console
$ img -h
Usage: img <command>

Commands:

  build    Build an image from a Dockerfile.
  du       Show image disk usage.
  ls       List images and digests.
  login    Log in to a Docker registry.
  pull     Pull an image or a repository from a registry.
  push     Push an image or a repository to a registry.
  version  Show the version information.
```

### Build an Image

```console
$ img build -h
Usage: img build [OPTIONS] PATH

Build an image from a Dockerfile.

Flags:

  -backend    backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -build-arg  Set build-time variables (default: [])
  -d          enable debug logging (default: false)
  -f          Name of the Dockerfile (Default is 'PATH/Dockerfile') (default: <none>)
  -state      directory to hold the global state (default: /tmp/img)
  -t          Name and optionally a tag in the 'name:tag' format (default: <none>)
  -target     Set the target build stage to build (default: <none>)
```

**Use just like you would `docker build`.**

```console
$ sudo img build -t jess/img .
Building jess/img
Setting up the rootfs... this may take a bit.
RUN [/bin/sh -c apk add --no-cache      ca-certificates]
--->
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/community/x86_64/APKINDEX.tar.gz
OK: 5 MiB in 12 packages
<--- 5e433zdbh8eosea0u9b70axb3 0 <nil>
RUN [copy /src-0 /dest/go/src/github.com/jessfraz/img]
--->
<--- rqku3imaivvjpgl676se1gupc 0 <nil>
RUN [/bin/sh -c set -x  && apk add --no-cache --virtual .build-deps             bash            git             gcc             libc-dev      libgcc           libseccomp-dev          linux-headers           make    && cd /go/src/github.com/jessfraz/img   && make static  && mv img /usr/bin/img         && mkdir -p /go/src/github.com/opencontainers   && git clone https://github.com/opencontainers/runc /go/src/github.com/opencontainers/runc     && cd /go/src/github.com/opencontainers/runc    && make static BUILDTAGS="seccomp" EXTRA_FLAGS="-buildmode pie" EXTRA_LDFLAGS="-extldflags \\\"-fno-PIC -static\\\""   && mv runc /usr/bin/runc        && apk del .build-deps  && rm -rf /go   && echo "Build complete."]
--->
+ apk add --no-cache --virtual .build-deps bash git gcc libc-dev libgcc libseccomp-dev linux-headers make
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/community/x86_64/APKINDEX.tar.gz
(1/28) Installing pkgconf (1.3.10-r0)
(2/28) Installing ncurses-terminfo-base (6.0_p20171125-r0)
(3/28) Installing ncurses-terminfo (6.0_p20171125-r0)
(4/28) Installing ncurses-libs (6.0_p20171125-r0)
(5/28) Installing readline (7.0.003-r0)
(6/28) Installing bash (4.4.19-r1)
....
....
RUN [copy /src-0/certs /dest/etc/ssl/certs]
--->
<--- 6ljir2x800w6deqlradhw0dy2 0 <nil>
Successfully built jess/img
```

### List Image Layers

```console
$ img ls -h
Usage: img ls [OPTIONS]

List images and digests.

Flags:

  -backend  backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -d        enable debug logging (default: false)
  -f        Filter output based on conditions provided (default: [])
  -state    directory to hold the global state (default: /tmp/img)
```

```console
$ img ls --backend=fuse
NAME                    SIZE            CREATED AT      UPDATED AT      DIGEST
jess/img:latest         1.534KiB        9 seconds ago   9 seconds ago   sha256:27d862ac32022946d61afbb91ddfc6a1fa2341a78a0da11ff9595a85f651d51e
jess/thing:latest       591B            30 minutes ago  30 minutes ago  sha256:d664b4e9b9cd8b3067e122ef68180e95dd4494fd4cb01d05632b6e77ce19118e
```

### Pull an Image

```console
$ img pull -h
Usage: img pull [OPTIONS] NAME[:TAG|@DIGEST]

Pull an image or a repository from a registry.

Flags:

  -backend  backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -d        enable debug logging (default: false)
  -state    directory to hold the global state (default: /tmp/img)
```

```console
$ img pull --backend=fuse r.j3ss.co/stress
Pulling r.j3ss.co/stress:latest...
Snapshot ref: sha256:2bb7a0a5f074ffe898b1ef64b3761e7f5062c3bdfe9947960e6db48a998ae1d6
Size: 365.9KiB
```

### Push an Image

```console
$ img push -h
Usage: img push [OPTIONS] NAME[:TAG]

Push an image or a repository to a registry.

Flags:

  -backend  backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -d        enable debug logging (default: false)
  -state    directory to hold the global state (default: /tmp/img)
```

```console
$ img push --backend=fuse jess/thing
Pushing jess/thing:latest...
Successfully pushed jess/thing:latest
```

### Disk Usage

```console
$ img du -h
Usage: img du [OPTIONS]

Show image disk usage.

Flags:

  -backend  backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -d        enable debug logging (default: false)
  -f        Filter output based on conditions provided (snapshot ID supported) (default: <none>)
  -state    directory to hold the global state (default: /tmp/img)
```

```console
$ img du --backend=fuse
ID                                                                      RECLAIMABLE     SIZE            DESCRIPTION
sha256:d9a48086f223d28a838263a6c04705c8009fab1dd67cc82c0ee821545de3bf7c true            911.8KiB        pulled from docker.io/tonistiigi/copy@sha256:476e0a67a1e4650c6adaf213269a2913deb7c52cbc77f954026f769d51e1a14e
7ia86xm2e4hzn2u947iqh9ph2                                               true            203.2MiB        mount /dest from exec copy /src-0 /dest/go/src/github.com/jessfraz/img
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

### Login to a Registry

```console
$ img login -h
Usage: img login [OPTIONS] [SERVER]

Log in to a Docker registry.
If no server is specified, the default (https://index.docker.io/v1/) is used.

Flags:

  -backend         backend for snapshots ([fuse naive overlayfs]) (default: naive)
  -d               enable debug logging (default: false)
  -p               Password (default: <none>)
  -password-stdin  Take the password from stdin (default: false)
  -state           directory to hold the global state (default: /tmp/img)
  -u               Username (default: <none>)
```

## Contributing

Please do! This is a new project and can use some love <3. Check out the [issues](https://github.com/jessfraz/img/issues).

The local directories are mostly re-implementations of `buildkit` interfaces to
be unprivileged.

## Acknowledgements

A lot of this is based on the work of [moby/buildkit](https://github.com/moby/buildkit). 
Thanks [@tonistiigi](https://github.com/tonistiigi) and
[@AkihiroSuda](https://github.com/AkihiroSuda)!
