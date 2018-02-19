# img

Standalone, Daemonless Dockerfile and OCI compatible container image builder.

**Goals: Runs completely in userspace. Currently not possible with FUSE problems, but working on it.**

### Usage

```console
$ img -h
Usage: img <command>

Commands:

  build    Build an image from a Dockerfile.
  version  Show the version information
```

```console
$ img build -h
Usage: build [OPTIONS] PATH img



Flags:

  -f       Name of the Dockerfile (Default is 'PATH/Dockerfile') (default: <none>)
  -t       Name and optionally a tag in the 'name:tag' format (default: <none>)
  -target  Set the target build stage to build (default: <none>)
  -v       enable verbose logging (default: false)
```

**Use just like you would `docker build`, currently it automatically pushes your image.**

```console
$ sudo img  build -t jess/img .
Building jess/img...
Setting up the rootfs... this may take a bit.
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc420404d00), Root:(*specs.Root)(0xc420345c00), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc4201cba40), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
RUN [/bin/sh -c apk add --no-cache      ca-certificates]
--->
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/community/x86_64/APKINDEX.tar.gz
OK: 5 MiB in 12 packages
<--- 5e433zdbh8eosea0u9b70axb3 0 <nil>
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc421b285b0), Root:(*specs.Root)(0xc421fb2a20), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/dest", Type:"overlay", Source:"overlay", Options:[]string{"workdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/16/work", "upperdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/16/fs", "lowerdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/14/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/13/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/12/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/11/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/10/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/9/fs"}}, specs.Mount{Destination:"/src-0", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/3/fs", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc421fce700), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
RUN [copy /src-0 /dest/go/src/github.com/jessfraz/img]
--->
<--- rqku3imaivvjpgl676se1gupc 0 <nil>
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc4218145b0), Root:(*specs.Root)(0xc422346800), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc421a0ea80), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
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
Executing bash-4.4.19-r1.post-install
(7/28) Installing libcurl (7.58.0-r0)
(8/28) Installing expat (2.2.5-r0)
(9/28) Installing pcre2 (10.30-r0)
(10/28) Installing git (2.15.0-r1)
(11/28) Installing binutils-libs (2.28-r3)
(12/28) Installing binutils (2.28-r3)
(13/28) Installing gmp (6.1.2-r1)
(14/28) Installing isl (0.18-r0)
(15/28) Installing libgomp (6.4.0-r5)
(16/28) Installing libatomic (6.4.0-r5)
(17/28) Installing libgcc (6.4.0-r5)
(18/28) Installing mpfr3 (3.1.5-r1)
(19/28) Installing mpc1 (1.0.3-r1)
(20/28) Installing libstdc++ (6.4.0-r5)
(21/28) Installing gcc (6.4.0-r5)
(22/28) Installing musl-dev (1.1.18-r3)
(23/28) Installing libc-dev (0.7.1-r0)
(24/28) Installing libseccomp (2.3.2-r0)
(25/28) Installing libseccomp-dev (2.3.2-r0)
(26/28) Installing linux-headers (4.4.6-r2)
(27/28) Installing make (4.2.1-r0)
(28/28) Installing .build-deps (0)
Executing busybox-1.27.2-r7.trigger
OK: 127 MiB in 40 packages
+ cd /go/src/github.com/jessfraz/img
+ make static
+ static
CGO_ENABLED=0 go build \
                        -tags " static_build" \
                        -ldflags "-w -X github.com/jessfraz/img/version.GITCOMMIT=0dddf56-dirty -X github.com/jessfraz/img/version.VERSION=v0.0.0 -extldflags -static" -o img .
+ mv img /usr/bin/img
+ mkdir -p /go/src/github.com/opencontainers
+ git clone https://github.com/opencontainers/runc /go/src/github.com/opencontainers/runc
Cloning into '/go/src/github.com/opencontainers/runc'...
+ cd /go/src/github.com/opencontainers/runc
+ make static BUILDTAGS=seccomp EXTRA_FLAGS=-buildmode pie EXTRA_LDFLAGS=-extldflags \"-fno-PIC -static\"
CGO_ENABLED=1 go build -buildmode pie -tags "seccomp netgo cgo static_build" -installsuffix netgo -ldflags "-w -extldflags -static -X main.gitCommit="a618ab5a0186905949ee463dbb762c3d23e12a80" -X main.version=1.0.0-rc4+dev -extldflags \"-fno-PIC -static\"" -o runc .
CGO_ENABLED=1 go build -buildmode pie -tags "seccomp netgo cgo static_build" -installsuffix netgo -ldflags "-w -extldflags -static -X main.gitCommit="a618ab5a0186905949ee463dbb762c3d23e12a80" -X main.version=1.0.0-rc4+dev -extldflags \"-fno-PIC -static\"" -o contrib/cmd/recvtty/recvtty ./contrib/cmd/recvtty
+ mv runc /usr/bin/runc
+ apk del .build-deps
WARNING: Ignoring APKINDEX.70c88391.tar.gz: No such file or directory
WARNING: Ignoring APKINDEX.5022a8a2.tar.gz: No such file or directory
(1/28) Purging .build-deps (0)
(2/28) Purging bash (4.4.19-r1)
Executing bash-4.4.19-r1.pre-deinstall
(3/28) Purging git (2.15.0-r1)
(4/28) Purging gcc (6.4.0-r5)
(5/28) Purging binutils (2.28-r3)
(6/28) Purging libatomic (6.4.0-r5)
(7/28) Purging libgomp (6.4.0-r5)
(8/28) Purging libc-dev (0.7.1-r0)
(9/28) Purging musl-dev (1.1.18-r3)
(10/28) Purging libseccomp-dev (2.3.2-r0)
(11/28) Purging libseccomp (2.3.2-r0)
(12/28) Purging linux-headers (4.4.6-r2)
(13/28) Purging make (4.2.1-r0)
(14/28) Purging pkgconf (1.3.10-r0)
(15/28) Purging readline (7.0.003-r0)
(16/28) Purging ncurses-libs (6.0_p20171125-r0)
(17/28) Purging ncurses-terminfo (6.0_p20171125-r0)
(18/28) Purging ncurses-terminfo-base (6.0_p20171125-r0)
(19/28) Purging libcurl (7.58.0-r0)
(20/28) Purging expat (2.2.5-r0)
(21/28) Purging pcre2 (10.30-r0)
(22/28) Purging binutils-libs (2.28-r3)
(23/28) Purging mpc1 (1.0.3-r1)
(24/28) Purging mpfr3 (3.1.5-r1)
(25/28) Purging isl (0.18-r0)
(26/28) Purging gmp (6.1.2-r1)
(27/28) Purging libstdc++ (6.4.0-r5)
(28/28) Purging libgcc (6.4.0-r5)
Executing busybox-1.27.2-r7.trigger
OK: 5 MiB in 12 packages
+ rm -rf /go
Build complete.
+ echo Build complete.
<--- ljjwzov00yx5wbpegv1zmvett 0 <nil>
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc42225cb60), Root:(*specs.Root)(0xc420b033c0), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/dest", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/19/fs", Options:[]string{"rw", "rbind"}}, specs.Mount{Destination:"/src-0/img", Type:"bind", Source:"/tmp/buildkit-mount908096361/usr/bin/img", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc4219b7880), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
RUN [copy /src-0/img /dest/usr/bin/img]
--->
<--- gqebdkrq8myjuf8zuecu5c5rw 0 <nil>
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc421815110), Root:(*specs.Root)(0xc422347260), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/dest", Type:"overlay", Source:"overlay", Options:[]string{"workdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/21/work", "upperdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/21/fs", "lowerdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/19/fs"}}, specs.Mount{Destination:"/src-0/runc", Type:"bind", Source:"/tmp/buildkit-mount642716333/usr/bin/runc", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc421a0f180), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
RUN [copy /src-0/runc /dest/usr/bin/runc]
--->
<--- su5420q73ry8lrz3j22yk9wgg 0 <nil>
spec: &specs.Spec{Version:"1.0.1", Process:(*specs.Process)(0xc422652ea0), Root:(*specs.Root)(0xc420564c80), Hostname:"", Mounts:[]specs.Mount{specs.Mount{Destination:"/proc", Type:"proc", Source:"proc", Options:[]string(nil)}, specs.Mount{Destination:"/dev", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/dev/pts", Type:"devpts", Source:"devpts", Options:[]string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"}}, specs.Mount{Destination:"/dev/shm", Type:"tmpfs", Source:"shm", Options:[]string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"}}, specs.Mount{Destination:"/dev/mqueue", Type:"mqueue", Source:"mqueue", Options:[]string{"nosuid", "noexec", "nodev"}}, specs.Mount{Destination:"/sys", Type:"sysfs", Source:"sysfs", Options:[]string{"nosuid", "noexec", "nodev", "ro"}}, specs.Mount{Destination:"/run", Type:"tmpfs", Source:"tmpfs", Options:[]string{"nosuid", "strictatime", "mode=755", "size=65536k"}}, specs.Mount{Destination:"/etc/resolv.conf", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/resolv.conf", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/etc/hosts", Type:"bind", Source:"/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/executor/hosts", Options:[]string{"rbind", "ro"}}, specs.Mount{Destination:"/dest", Type:"overlay", Source:"overlay", Options:[]string{"workdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/23/work", "upperdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/23/fs", "lowerdir=/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/21/fs:/tmp/img/xlxkzz3yxd6nfmsr91u723gxt/runc-overlayfs/snapshots/snapshots/19/fs"}}, specs.Mount{Destination:"/src-0/certs", Type:"bind", Source:"/tmp/buildkit-mount530478641/etc/ssl/certs", Options:[]string{"rbind", "ro"}}}, Hooks:(*specs.Hooks)(nil), Annotations:map[string]string(nil), Linux:(*specs.Linux)(0xc422733ce0), Solaris:(*specs.Solaris)(nil), Windows:(*specs.Windows)(nil)}
RUN [copy /src-0/certs /dest/etc/ssl/certs]
--->
<--- 6ljir2x800w6deqlradhw0dy2 0 <nil>
Built and pushed image: jess/img
```
