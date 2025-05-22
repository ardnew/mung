SHELL := $(firstword $(shell which bash sh))

modpkg := $(shell go list -f '{{.Name}}' .)
moddir := $(shell go list -f '{{.Dir}}' .)
modimp := $(shell go list -f '{{.ImportPath}}' .)
cmdpkg := $(shell go list -f '{{.Name}}' ./cmd/$(modpkg))
cmddir := $(shell go list -f '{{.Dir}}' ./cmd/$(modpkg))
cmdimp := $(shell go list -f '{{.ImportPath}}' ./cmd/$(modpkg))

tagsemver := $(or $(VERSION),$(shell git describe --tags --abbrev=0 2>/dev/null))
modsemver := $(or $(MODVERSION),$(tagsemver:v%=%))
cmdsemver := $(or $(CMDVERSION),$(tagsemver:v%=%))

output := dist
assets := README.md LICENSE

platforms := $(foreach os,linux darwin windows,$(foreach arch,amd64 arm64,$(os)-$(arch)))
platform  := $(or $(PLATFORM),$(platforms))

dist      := $(foreach p,$(platform),dist-$(p))
clean     := $(foreach p,$(platform),clean-$(p))
distclean := $(foreach p,$(platform),distclean-$(p))

os = $(firstword $(subst -, ,$(1)))
arch = $(lastword $(subst -, ,$(1)))

.PHONY: all generate version dist clean distclean $(platform) $(dist) $(clean) $(distclean)

all: $(platform)

# double-hyphen prevents usage from command-line.
# Make will interpret it as an invalid option and exit.
.PHONY: --force

# An empty recipe is always considered out of date.
# Any targets that depend on it will always be rebuilt.
--force:

generate: version
	go generate -v ./...

version: $(moddir)/VERSION $(cmddir)/VERSION
	@echo "$(moddir) version $(shell cat $(moddir)/VERSION)"
	@echo "$(cmddir) version $(shell cat $(cmddir)/VERSION)"

dist: $(dist)

clean: $(clean)

distclean: $(distclean)

$(moddir)/VERSION: --force
ifeq ($(strip $(modsemver)),)
	$(error unknown module version: set MODVERSION or tag the repository)
endif
	@echo ${modsemver} > $@

$(cmddir)/VERSION: --force
ifeq ($(strip $(cmdsemver)),)
	$(error unknown command version: set CMDVERSION or tag the repository)
endif
	@echo ${cmdsemver} > $@

$(platform): generate
	@echo
	@echo build $@
	@echo
	@mkdir -p $(output)/$(modpkg)$(cmdsemver).$@
	GOOS=$(call os,$@) GOARCH=$(call arch,$@) go build -v -ldflags="-s -w" -o $(output)/$(modpkg)$(cmdsemver).$@/$(modpkg) ./cmd/$(modpkg)

.SECONDEXPANSION:
$(dist): $$(subst dist-,,$$@)
	@echo
	@echo dist $<
	@echo
	@cp $(assets) $(output)/$(modpkg)$(cmdsemver).$</
	tar -czf $(output)/$(modpkg)$(cmdsemver).$<.tar.gz -C $(output) $(modpkg)$(cmdsemver).$<

$(clean):
	@echo
	@echo clean $(subst clean-,,$@)
	@echo
	GOOS=$(call os,$@) GOARCH=$(call arch,$@) go clean -i -r $(modimp) $(cmdimp)

.SECONDEXPANSION:
$(distclean): $$(subst distclean-,clean-,$$@)
	@echo
	@echo distclean $(subst clean-,,$<)
	@echo
	rm -rf $(output)/$(modpkg)$(cmdsemver).$(subst clean-,,$<)
	rm -f $(output)/$(modpkg)$(cmdsemver).$(subst clean-,,$<).tar.gz
