GOBUILD=go build
BINARY_NAME=tobw
VERSION=$(shell git describe --tags --always --long --dirty)
LD_FLAGS="-s -w -X main.Version=$(VERSION)"

GO_FILES?=$$(find . -name '*.go' | grep -v vendor)
EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/goimports \
	github.com/golang/dep/cmd/dep \
	github.com/client9/misspell/cmd/misspell

all: build

build: $(BINARY_NAME)_linux_amd64 \
	$(BINARY_NAME)_darwin_amd64 \
	$(BINARY_NAME)_windows_amd64 \
	$(BINARY_NAME)_freebsd_amd64 \
	$(BINARY_NAME)_openbsd_amd64 \
	$(BINARY_NAME)_openbsd_amd64 \
	$(BINARY_NAME)_linux_386 \
	$(BINARY_NAME)_darwin_386 \
	$(BINARY_NAME)_windows_386 \
	$(BINARY_NAME)_freebsd_386 \
	$(BINARY_NAME)_openbsd_386 \
	$(BINARY_NAME)_netbsd_386 \
	$(BINARY_NAME)_linux_arm \
	$(BINARY_NAME)_linux_arm64 \
	$(BINARY_NAME)_freebsd_arm


clean:
	@echo "*** Cleaning up object files... ***"
	rm -r -f bin/*
	rm -f tobw
	rm -f tobw.exe
	go clean
	@echo "*** Done ***"

fmt:
	@echo "*** Applying gofmt on all .go files (excluding vendor)... ***"
	@goimports -w $(GO_FILES)
	@echo "*** Done ***"

check_spelling:
	@echo "*** Check for common spelling mistakes in .go files... ***"
	@misspell -error $(GO_FILES)
	@echo "*** Done ***"

fix_spelling:
	@echo "*** Fix any encountered spelling mistakes in .go files... ***"
	@misspell -w $(GO_FILES)
	@echo "*** Done ***"

vet:
	@echo "*** Running vet on package directories... ***"
	@go list ./... | grep -v /vendor/ | xargs go vet
	@echo "*** Done ***"

dep:
	@echo "*** Installing all dependencies... ***"
	@dep ensure
	@echo "*** Done ***"

local:
	@echo "*** Building local binary... ***"
	$(GOBUILD) -o $(BINARY_NAME) ./
	@echo "*** Done ***"

bootstrap:
	@echo "*** Installing required tools for building... ***"
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing/Updating $$tool" ; \
		GO111MODULE=off go get -u $$tool; \
	done
	@echo "*** Done ***"

help:
	@echo "*** Supported commands ***"
	@echo "make bootstrap:        Install tools needed for build."
	@echo "make dep:              Install libraries needed for compilation."
	@echo "make build:            Compile for all targets."
	@echo "make vet:              Checks code for common mistakes."
	@echo "make fmt:              Fix formatting on .go files"
	@echo "make check_spelling:   Show potential spelling mistakes."
	@echo "make fix_spelling:     Correct detected spelling mistakes."
	@echo "make local:            Build executable for your OS, for testing purposes."
	@echo "make clean:            Clean your working directory."
	@echo "*** Done ***"


# Compile common amd64 platforms.
$(BINARY_NAME)_darwin_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_windows_amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$@.exe -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_freebsd_amd64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_netbsd_amd64:
	CGO_ENABLED=0 GOOS=netbsd GOARCH=amd64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_openbsd_amd64:
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

# Compile common 386 platforms.
$(BINARY_NAME)_darwin_386:
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_windows_386:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o bin/$@.exe -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_linux_386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_freebsd_386:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_netbsd_386:
	CGO_ENABLED=0 GOOS=netbsd GOARCH=386 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_openbsd_386:
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

# For ARM targets, only Linux and FreeBSD support for now.
$(BINARY_NAME)_linux_arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

$(BINARY_NAME)_freebsd_arm:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm $(GOBUILD) -o bin/$@ -v -ldflags=$(LD_FLAGS)

