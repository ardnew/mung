SHELL := $(firstword $(shell which bash sh))

modpkg := $(shell go list -f '{{.Name}}' .)
moddir := $(shell go list -f '{{.Dir}}' .)
modimp := $(shell go list -f '{{.ImportPath}}' .)
cmdpkg := $(shell go list -f '{{.Name}}' ./cmd/$(modpkg))
cmddir := $(shell go list -f '{{.Dir}}' ./cmd/$(modpkg))
cmdimp := $(shell go list -f '{{.ImportPath}}' ./cmd/$(modpkg))

tagsemver := $(shell git describe --tags --abbrev=0 2>/dev/null)
modsemver := $(or $(MODVERSION),$(tagsemver:v%=%))
cmdsemver := $(or $(CMDVERSION),$(tagsemver:v%=%))

output := dist
assets := README.md LICENSE

platforms := $(foreach os,linux darwin windows,$(foreach arch,amd64 arm64,$(os)-$(arch)))
platform := $(or $(PLATFORM),$(platforms))
os = $(firstword $(subst -, ,$(1)))
arch = $(lastword $(subst -, ,$(1)))

.PHONY: all overwrite version clean

all: version $(platform)

version: $(moddir)/VERSION $(cmddir)/VERSION
	@echo "$(moddir) version $(shell cat $(moddir)/VERSION)"
	@echo "$(cmddir) version $(shell cat $(cmddir)/VERSION)"

$(moddir)/VERSION: overwrite
ifeq ($(strip $(modsemver)),)
	$(error unknown module version: set MODVERSION or tag the repository)
endif
	@echo ${modsemver} > $@

$(cmddir)/VERSION: overwrite
ifeq ($(strip $(cmdsemver)),)
	$(error unknown command version: set CMDVERSION or tag the repository)
endif
	@echo ${cmdsemver} > $@

$(platform): GOOS   = $(call os,$@)
$(platform): GOARCH = $(call arch,$@)
$(platform):
	@echo building $@
	@mkdir -p $(output)/$(modpkg)$(cmdsemver).$@
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-s -w" \
		-o $(output)/$(modpkg)$(cmdsemver).$@/$(modpkg) \
		./cmd/$(modpkg)
	@cp $(assets) $(output)/$(modpkg)$(cmdsemver).$@/
	tar -czf $(output)/$(modpkg)$(cmdsemver).$@.tar.gz \
		-C $(output) $(modpkg)$(cmdsemver).$@

clean:
	rm -rf $(output)
	go clean -i -r $(modimp) $(cmdimp)

overwrite:

