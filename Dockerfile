FROM golang:alpine
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates \
	fuse \
	git

COPY . /go/src/github.com/jessfraz/img

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		bash \
		gcc \
		libc-dev \
		libgcc \
		libseccomp-dev \
		linux-headers \
		make \
	&& cd /go/src/github.com/jessfraz/img \
	&& make static \
	&& mv img /usr/bin/img \
	&& mkdir -p /go/src/github.com/opencontainers \
	&& git clone https://github.com/opencontainers/runc /go/src/github.com/opencontainers/runc \
	&& cd /go/src/github.com/opencontainers/runc \
	&& make static BUILDTAGS="seccomp" EXTRA_FLAGS="-buildmode pie" EXTRA_LDFLAGS="-extldflags \\\"-fno-PIC -static\\\"" \
	&& mv runc /usr/bin/runc \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

ENTRYPOINT [ "img" ]
CMD [ "--help" ]
