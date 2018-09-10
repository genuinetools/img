# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Setup name variables for the package/tool
NAME := reg
PKG := github.com/genuinetools/$(NAME)

# Set any default go build tags
BUILDTAGS :=

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/cross

# Populate version variables
# Add to compile time flags
VERSION := $(shell cat VERSION.txt)
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# Set our default go compiler
GO := go

# List the GOOS and GOARCH to build
GOOSARCHES = $(shell cat .goosarch)

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package

$(NAME): generated-assets $(wildcard *.go) $(wildcard */*.go) VERSION.txt
	@echo "+ $@"
	$(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

.PHONY: static
static: generated-assets ## Builds a static executable
	@echo "+ $@"
	CGO_ENABLED=0 $(GO) build \
				-tags "$(BUILDTAGS) static_build" \
				${GO_LDFLAGS_STATIC} -o $(NAME) .

all: clean build fmt lint test staticcheck vet install ## Runs a clean, build, fmt, lint, test, staticcheck, vet and install

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | grep -v internal | tee /dev/stderr

.PHONY: test
test: generated-assets ## Runs the go tests
	@echo "+ $@"
	@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes
	@echo "+ $@"
	@$(GO) vet $(shell $(GO) list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: staticcheck
staticcheck: ## Verifies `staticcheck` passes
	@echo "+ $@"
	@staticcheck $(shell $(GO) list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: cover
cover: generated-assets ## Runs go test with coverage
	@echo "" > coverage.txt
	@for d in $(shell $(GO) list ./... | grep -v vendor); do \
		$(GO) test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

.PHONY: install
install: generated-assets ## Installs the executable or package
	@echo "+ $@"
	$(GO) install -a -tags "$(BUILDTAGS)" ${GO_LDFLAGS} .

define buildpretty
mkdir -p $(BUILDDIR)/$(1)/$(2);
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 $(GO) build \
	 -o $(BUILDDIR)/$(1)/$(2)/$(NAME) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).md5;
sha256sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).sha256;
endef

.PHONY: cross
cross: generated-assets *.go VERSION.txt ## Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildpretty,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 $(GO) build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).md5;
sha256sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: release
release: generated-assets *.go VERSION.txt ## Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildrelease,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

.PHONY: bump-version
BUMP := patch
bump-version: ## Bump the version in the version file. Set BUMP to [ patch | major | minor ]
	@$(GO) get -u github.com/jessfraz/junk/sembump # update sembump tool
	$(eval NEW_VERSION = $(shell sembump --kind $(BUMP) $(VERSION)))
	@echo "Bumping VERSION.txt from $(VERSION) to $(NEW_VERSION)"
	echo $(NEW_VERSION) > VERSION.txt
	@echo "Updating links to download binaries in README.md"
	sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
	git add VERSION.txt README.md
	git commit -vsam "Bump version to $(NEW_VERSION)"
	@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"

.PHONY: tag
tag: ## Create a new git tag to prepare to build a release
	git tag -sa $(VERSION) -m "$(VERSION)"
	@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a travis build."

.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

SERVER_ASSETS_DIR := $(CURDIR)/server
BINDATA_DIR := $(CURDIR)/internal/binutils

.PHONY: generated-assets
generated-assets: $(BINDATA_DIR)/templates.go $(BINDATA_DIR)/templates.go

$(BINDATA_DIR):
	@mkdir -p $@

$(BINDATA_DIR)/templates.go: $(BINDATA_DIR) $(wildcard *.go) $(wildcard server/templates/*)
	@$(GO) get -u github.com/jteeuwen/go-bindata/... # update go-bindata tool
	go-bindata -pkg binutils -prefix "$(SERVER_ASSETS_DIR)" -o $@ $(SERVER_ASSETS_DIR)/templates
	gofmt -s -w $@

$(BINDATA_DIR)/static.go: $(BINDATA_DIR) $(wildcard *.go) $(wildcard server/static/*)
	@$(GO) get -u github.com/jteeuwen/go-bindata/... # update go-bindata tool
	go-bindata -pkg binutils -prefix "$(SERVER_ASSETS_DIR)" -o $@ $(SERVER_ASSETS_DIR)/static
	gofmt -s -w $@

.PHONY: clean
clean: ## Cleanup any build binaries or packages
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)
	$(RM) -r $(BINDATA_DIR)
	sudo $(RM) -r $(CURDIR)/.certs

# set the graph driver as the current graphdriver if not set
DOCKER_GRAPHDRIVER := $(if $(DOCKER_GRAPHDRIVER),$(DOCKER_GRAPHDRIVER),$(shell docker info 2>&1 | grep "Storage Driver" | sed 's/.*: //'))
export DOCKER_GRAPHDRIVER

# if this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

.PHONY: dind
DIND_CONTAINER=reg-dind
DIND_DOCKER_IMAGE=r.j3ss.co/docker:userns
dind: stop-dind ## Starts a docker-in-docker container for running the tests with
	docker run -d  \
		--name $(DIND_CONTAINER) \
		--privileged \
		-v $(CURDIR)/.certs:/etc/docker/ssl \
		-v $(CURDIR):/go/src/github.com/genuinetools/reg \
		-v /tmp:/tmp \
		$(DIND_DOCKER_IMAGE) \
		dockerd -D --storage-driver $(DOCKER_GRAPHDRIVER) \
		-H tcp://127.0.0.1:2375 \
		--host=unix:///var/run/docker.sock \
		--exec-opt=native.cgroupdriver=cgroupfs \
		--insecure-registry localhost:5000 \
		--tlsverify \
		--tlscacert=/etc/docker/ssl/cacert.pem \
		--tlskey=/etc/docker/ssl/server.key \
		--tlscert=/etc/docker/ssl/server.cert

.PHONY: stop-dind
stop-dind: ## Stops the docker-in-docker container
	@docker rm -f $(DIND_CONTAINER) >/dev/null 2>&1 || true

.PHONY: dtest
DOCKER_IMAGE := reg-dev
dtest: ## Run the tests in a docker container
	docker build --rm --force-rm -f Dockerfile.dev -t $(DOCKER_IMAGE) .
	docker run --rm -i $(DOCKER_FLAGS) \
		-v $(CURDIR):/go/src/github.com/genuinetools/reg \
		--workdir /go/src/github.com/genuinetools/reg \
		-v $(CURDIR)/.certs:/etc/docker/ssl:ro \
		-v /tmp:/tmp \
		--disable-content-trust=true \
		--net container:$(DIND_CONTAINER) \
		-e DOCKER_HOST=tcp://127.0.0.1:2375 \
		-e DOCKER_TLS_VERIFY=true \
		-e DOCKER_CERT_PATH=/etc/docker/ssl \
		-e DOCKER_API_VERSION \
		$(DOCKER_IMAGE) \
		make test

.PHONY: snakeoil
snakeoil: ## Update snakeoil certs for testing
	go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost,127.0.0.1 --ca
	mv $(CURDIR)/key.pem $(CURDIR)/testutils/snakeoil/key.pem
	mv $(CURDIR)/cert.pem $(CURDIR)/testutils/snakeoil/cert.pem

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
