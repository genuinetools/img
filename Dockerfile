FROM golang:alpine
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates \
	fuse \
	git

ENV DIND_COMMIT 3b5fac462d21ca164b3778647420016315289034

RUN set -ex; \
	apk add --no-cache --virtual .fetch-deps libressl; \
	wget -O /usr/local/bin/dind "https://raw.githubusercontent.com/docker/docker/${DIND_COMMIT}/hack/dind"; \
	chmod +x /usr/local/bin/dind; \
	apk del .fetch-deps

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
