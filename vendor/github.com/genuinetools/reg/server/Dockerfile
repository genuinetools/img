FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
	&& go get -v github.com/genuinetools/reg \
	&& cd /go/src/github.com/genuinetools/reg \
	&& CGO_ENABLED=0 go build -a -tags netgo -ldflags '-extldflags "-static"' -o /usr/bin/reg-server ./server \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/reg-server /usr/bin/reg-server
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

COPY static /src/static
COPY templates /src/templates

WORKDIR /src

ENTRYPOINT [ "reg-server" ]
CMD [ "--help" ]
