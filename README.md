# img

Standalone, Daemonless Dockerfile and OCI compatible container image builder.

**Goals: Runs completely in userspace. Currently not possible with FUSE problems, but working on it.**

## Usage

```console
Usage: img <command>

Commands:

  build    Build an image from a Dockerfile.
  ls       List images and digests.
  pull     Pull an image or a repository from a registry.
  version  Show the version information.
```

### Build an Image

```console
$ img build -h
Usage: img build [OPTIONS] PATH

Build an image from a Dockerfile.

Flags:

  -build-arg  Set build-time variables (default: [])
  -d          enable debug logging (default: false)
  -f          Name of the Dockerfile (Default is 'PATH/Dockerfile') (default: <none>)
  -t          Name and optionally a tag in the 'name:tag' format (default: <none>)
  -target     Set the target build stage to build (default: <none>)
```

**Use just like you would `docker build`.**

```console
$ sudo img build -t jess/img .
Building jess/img...
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
Built and pushed image: jess/img
```

### List Image Layers

```console
$ img ls -h
Usage: img ls [OPTIONS]

List images and digests.

Flags:

  -d  enable debug logging (default: false)
  -f  Filter output based on conditions provided (snapshot ID supported) (default: <none>)
```

```console
$ img ls
ID                                                                      RECLAIMABLE     SIZE            LAST ACCESSED
sha256:2bb7a0a5f074ffe898b1ef64b3761e7f5062c3bdfe9947960e6db48a998ae1d6 true            365.9KiB
sha256:aa74a6c91df06c8a41629caf62cc5f2dbb8f6a8f278aff042bd45ad1cc573b8d true            297.9KiB
sha256:e706942da64d71182be53550aca301898157f19999706c7f7e895151d512224f true            9.051KiB
sha256:6b0518a7493b951bc13dd85b9c27f72912fdf1871674c869cd7881c4235a29e2 true            1.44MiB
9fya41x2qhytrqkcsbd4c0qq1*                                              true            154.5MiB
sha256:cd7100a72410606589a54b932cabd804a17f9ae5b42a1882bd56d263e02b6215 true            6.258MiB
af61bkht0jb3dhfd2qkqftwg6                                               true            420.8KiB
dfh29u56whf6aem91edvume8m                                               true            154.6MiB
khp847rbw1we6uc2h8m6xlyqp*                                              true            5.039KiB
utpsw3cgpxsqg5ih7e3rqrzyg                                               true            1.075MiB
sha256:ae4ecac23119cc920f9e44847334815d32bdf82f6678069d8a8be103c1ee2891 true            148.9MiB
c1yqlb3e05crmll7v0lvltfuc*                                              true            4KiB
sha256:db193011cbfc238d622d65c4099750758df83d74571e8d7498392b17df381207 true            467.2MiB
vr535b5pvgcanvuttw8lx1lfw*                                              true            4.204KiB
sha256:c4151b5a5de5b7e272b2b6a3a4518c980d6e7f580f39c85370330a1bff5821f1 true            472.3KiB
sha256:d9a48086f223d28a838263a6c04705c8009fab1dd67cc82c0ee821545de3bf7c true            911.8KiB
vxzb27136nn8934gs88857ezj                                               true            10.48MiB
w75mt1b5ho9ck968ebgp3l27w                                               true            17.92MiB
l60avgz32ot9jw2okw7lb1u24*                                              true            73.22MiB
ne7zasdkxi41q6f0dnnclh7dj                                               true            56.38KiB
sha256:46b4a1f9227b1b8aa0948d2ed39beb3a74de0f34c4f9fd9bf7e4a5f00e781bd6 true            16.12KiB
tsvyai0nrdfweeptigv9nb3di*                                              true            4KiB
sha256:2f648eb75764cf89e8a6327da5e5b0b61c31785e7fd15f206298a42544a7e4e5 true            217.5KiB
sha256:9f131fba0383a6aaf25ecd78bd5f37003e41a4385d7f38c3b0cde352ad7676da true            958.6KiB
Reclaimable:    1.015GiB
Total:          1.015GiB
```

### Pull an Image

```console
$ img pull -h
Usage: img pull [OPTIONS] NAME[:TAG|@DIGEST]

Pull an image or a repository from a registry.

Flags:

  -d  enable debug logging (default: false)
```

```console
$ img pull r.j3ss.co/stress
Pulling r.j3ss.co/stress:latest...
Snapshot ref: sha256:2bb7a0a5f074ffe898b1ef64b3761e7f5062c3bdfe9947960e6db48a998ae1d6
Size: 365.9KiB
```

## Acknowledgements

A lot of this is based on the work of [moby/buildkit](https://github.com/moby/buildkit). Thanks!
