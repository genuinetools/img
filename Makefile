# Setup name variables for the package/tool
NAME := img
PKG := github.com/genuinetools/$(NAME)

CGO_ENABLED := 1

include basic.mk

RUNCBUILDDIR=$(BUILDDIR)/src/github.com/opencontainers/runc
$(RUNCBUILDDIR):
	git clone --depth 1 https://github.com/opencontainers/runc.git "$@"

$(RUNCBUILDDIR)/runc: $(RUNCBUILDDIR)
	GOPATH=$(BUILDDIR) $(MAKE) -C "$(RUNCBUILDDIR)" static BUILDTAGS="seccomp apparmor"

internal/binutils/runc.go: $(RUNCBUILDDIR)/runc
	@$(GO) get -u github.com/jteeuwen/go-bindata/... # update go-bindata tool
	go-bindata -tags \!noembed -pkg binutils -prefix "$(RUNCBUILDDIR)" -o $(CURDIR)/internal/binutils/runc.go $(RUNCBUILDDIR)/runc
	gofmt -s -w $(CURDIR)/internal/binutils/runc.go

.PHONY: runc
ifneq (,$(findstring noembed,$(BUILDTAGS)))
runc: ## No-op when not embedding runc.
else
runc: internal/binutils/runc.go ## Builds runc locally so it can be embedded in the resulting binary.
	$(RM) -r $(RUNCBUILDDIR)
endif
