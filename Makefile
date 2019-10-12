# Setup name variables for the package/tool
NAME := img
PKG := github.com/genuinetools/$(NAME)

CGO_ENABLED := 1

# Set any default go build tags
BUILDTAGS ?= seccomp

include basic.mk

.PHONY: prebuild
prebuild: runc

RUNCBUILDDIR=$(BUILDDIR)/src/github.com/opencontainers/runc
RUNCCOMMIT=7cb3cde1f49eae53fb8fff5012c0750a64eb928b
$(RUNCBUILDDIR):
	git clone https://github.com/opencontainers/runc.git "$@"
	

$(BUILDDIR)/bin/runc: $(RUNCBUILDDIR)
	( cd $(RUNCBUILDDIR) ; git checkout $(RUNCCOMMIT) )
	GOPATH=$(BUILDDIR) $(MAKE) -C "$(RUNCBUILDDIR)" static BUILDTAGS="seccomp apparmor"
	mkdir -p $(BUILDDIR)/bin
	mv $(RUNCBUILDDIR)/runc $(BUILDDIR)/bin/runc

internal/binutils/runc.go: $(BUILDDIR)/bin/runc
	@$(GO) get -u github.com/go-bindata/go-bindata # update go-bindata tool
	go-bindata -tags \!noembed -pkg binutils -prefix "$(BUILDDIR)/bin" -o $(CURDIR)/internal/binutils/runc.go $(BUILDDIR)/bin
	gofmt -s -w $(CURDIR)/internal/binutils/runc.go

.PHONY: runc
ifneq (,$(findstring noembed,$(BUILDTAGS)))
runc: ## No-op when not embedding runc.
else
runc: internal/binutils/runc.go ## Builds runc locally so it can be embedded in the resulting binary.
	$(RM) -r $(RUNCBUILDDIR)
runc-install: $(RUNCBUILDDIR)/runc
	sudo cp $(RUNCBUILDDIR)/runc /usr/bin/runc
	$(RM) -r $(RUNCBUILDDIR)
endif
