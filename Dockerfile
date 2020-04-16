FROM golang:1.14-alpine AS gobuild-base
RUN apk add --no-cache \
	bash \
	build-base \
	gcc \
	git \
	libseccomp-dev \
	linux-headers \
	make

FROM gobuild-base AS img
WORKDIR /img
COPY . .
RUN cd / && go get github.com/jteeuwen/go-bindata/go-bindata && cd /img
RUN make static && mv img /usr/bin/img

FROM alpine:3.11 AS base
MAINTAINER Jessica Frazelle <jess@linux.com>
RUN apk add --no-cache git shadow-uidmap
COPY --from=img /usr/bin/img /usr/bin/img
RUN chmod u+s /usr/bin/newuidmap /usr/bin/newgidmap \
  && adduser -D -u 1000 user \
  && mkdir -p /run/user/1000 \
  && chown -R user /run/user/1000 /home/user \
  && echo user:100000:65536 | tee /etc/subuid | tee /etc/subgid

FROM base AS debug
RUN apk add --no-cache bash strace

FROM base AS release
USER user
ENV USER user
ENV HOME /home/user
ENV XDG_RUNTIME_DIR=/run/user/1000
ENTRYPOINT [ "img" ]
CMD [ "--help" ]
