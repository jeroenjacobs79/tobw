GOBUILD=go build
GOX=gox
BINARY_NAME=tobw

VERSION=$(shell git describe --tags --always --dirty)
# linker flags for stripping debug info and injecting version info
LD_FLAGS="-s -w -X main.Version=$(VERSION)"
BIN_TARGETS="windows/386 windows/amd64 darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64 freebsd/386 freebsd/amd64 freebsd/arm openbsd/386 openbsd/amd64 netbsd/386 netbsd/amd64 netbsd/arm"


# Used for help output
HELP_SPACING=15
HELP_COLOR=33
HELP_FORMATSTRING="\033[$(HELP_COLOR)m%-$(HELP_SPACING)s \033[00m%s.\n"

GO_FILES?=$$(find . -name '*.go' | grep -v vendor)
EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/goimports \
	github.com/golang/dep/cmd/dep \
	github.com/client9/misspell/cmd/misspell \
	github.com/mitchellh/gox

.PHONY: build local clean fmt check_spelling fix_spelling vet dep bootstrap help

build:
	@echo "*** Building binaries for supported architectures... ***"
	$(GOX) -osarch=$(BIN_TARGETS) -ldflags=$(LD_FLAGS) -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}"
	@echo "*** Done ***"
    
local:
	@echo "*** Building local binary... ***"
	$(GOBUILD) -o $(BINARY_NAME) ./
	@echo "*** Done ***"

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

bootstrap:
	@echo "*** Installing required tools for building... ***"
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing/Updating $$tool" ; \
		GO111MODULE=off go get -u $$tool; \
	done
	@echo "*** Done ***"

help:
	@printf "\n*** Available make targets ***\n\n"
	@printf $(HELP_FORMATSTRING) "help" "This message"
	@printf $(HELP_FORMATSTRING) "bootstrap" "Install tools needed for build"
	@printf $(HELP_FORMATSTRING) "dep" "Install libraries needed for compilation"
	@printf $(HELP_FORMATSTRING) "build" "Compile for all targets"
	@printf $(HELP_FORMATSTRING) "vet" "Checks code for common mistakes"
	@printf $(HELP_FORMATSTRING) "fmt" "Fix formatting on .go files"
	@printf $(HELP_FORMATSTRING) "check_spelling" "Show potential spelling mistakes"
	@printf $(HELP_FORMATSTRING) "fix_spelling" "Correct detected spelling mistakes"
	@printf $(HELP_FORMATSTRING) "local" "Build executable for your OS, for testing purposes"
	@printf $(HELP_FORMATSTRING) "clean" "Clean your working directory"
	@printf "\n*** End ***\n\n"
