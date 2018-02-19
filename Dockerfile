FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/jessfraz/img

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		bash \
		git \
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

FROM scratch

COPY --from=builder /usr/bin/img /usr/bin/img
COPY --from=builder /usr/bin/runc /usr/bin/runc
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "img" ]
CMD [ "--help" ]
