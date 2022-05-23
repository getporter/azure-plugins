PLUGIN = azure
PKG = get.porter.sh/plugin/$(PLUGIN)
SHELL = bash

PORTER_HOME ?= $(HOME)/.porter

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git describe --tags --exact-match &> /dev/null && echo latest || echo canary)

GO = GO111MODULE=on go
RECORDTEST = RECORDER_MODE=record $(GO)
LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
XBUILD = CGO_ENABLED=0 $(GO) build -a -tags netgo -ldflags '$(LDFLAGS)'
BINDIR = bin/plugins/$(PLUGIN)

CLIENT_PLATFORM ?= $(shell go env GOOS)
CLIENT_ARCH ?= $(shell go env GOARCH)
SUPPORTED_PLATFORMS = linux darwin windows
SUPPORTED_ARCHES = amd64

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

debug: build-for-debug install

build-for-debug:
	mkdir -p $(BINDIR)
	$(GO) build -o $(BINDIR)/$(PLUGIN)$(FILE_EXT) ./cmd/$(PLUGIN)

.PHONY: build
build:
	mkdir -p $(BINDIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(PLUGIN)$(FILE_EXT) ./cmd/$(PLUGIN)

xbuild-all:
	$(foreach OS, $(SUPPORTED_PLATFORMS), \
		$(foreach ARCH, $(SUPPORTED_ARCHES), \
				$(MAKE) $(MAKE_OPTS) CLIENT_PLATFORM=$(OS) CLIENT_ARCH=$(ARCH) PLUGIN=$(PLUGIN) xbuild; \
		))

xbuild: $(BINDIR)/$(VERSION)/$(PLUGIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(PLUGIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(CLIENT_PLATFORM) GOARCH=$(CLIENT_ARCH) $(XBUILD) -o $@ ./cmd/$(PLUGIN)

test: test-unit
	$(BINDIR)/$(PLUGIN)$(FILE_EXT) version

test-unit: build
	$(GO) test ./...

test-recorder: build
	# AZURE_STORAGE_CONNECTION_STRING is required to be set for this command
	$(RECORDTEST) test ./pkg/azure/table/...

publish: bin/porter$(FILE_EXT)
	go run mage.go -v Publish

bin/porter$(FILE_EXT):
	curl -fsSLo bin/porter$(FILE_EXT) https://cdn.porter.sh/canary/porter-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
	chmod +x bin/porter$(FILE_EXT)

install:
	mkdir -p $(PORTER_HOME)/plugins/$(PLUGIN)
	install $(BINDIR)/$(PLUGIN)$(FILE_EXT) $(PORTER_HOME)/plugins/$(PLUGIN)/$(PLUGIN)$(FILE_EXT)

clean:
	-rm -fr bin/
