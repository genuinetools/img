FROM golang:alpine

RUN apk add --no-cache \
	build-base \
	ca-certificates \
	git

ENV DOCKER_BUCKET get.docker.com
ENV DOCKER_VERSION 1.11.1
ENV DOCKER_SHA256 893e3c6e89c0cd2c5f1e51ea41bc2dd97f5e791fcfa3cee28445df277836339d

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		curl \
	&& curl -fSL "https://${DOCKER_BUCKET}/builds/Linux/x86_64/docker-$DOCKER_VERSION.tgz" -o docker.tgz \
	&& echo "${DOCKER_SHA256} *docker.tgz" | sha256sum -c - \
	&& tar -xzvf docker.tgz \
	&& mv docker/docker /usr/local/bin/ \
	&& rm -rf docker \
	&& rm docker.tgz \
	&& docker -v \
	&& apk del .build-deps

