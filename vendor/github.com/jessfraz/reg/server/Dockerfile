FROM alpine:latest
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY static /src/static
COPY templates /src/templates

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		go \
		git \
		gcc \
		libc-dev \
		libgcc \
	&& go get -v github.com/jessfraz/reg \
	&& cd /go/src/github.com/jessfraz/reg \
	&& go build -o /usr/bin/reg-server ./server \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

WORKDIR /src

ENTRYPOINT [ "reg-server" ]
