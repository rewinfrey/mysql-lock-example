APPVERSION=$(or $(shell git rev-parse HEAD 2>/dev/null),"unknown")
EXAMPLEBIN = bin/example

GO = go
V  = 0
Q  = $(if $(filter 1,$V),,@)
M  = $(shell printf "\033[34;1m▶\033[0m")
OS = $(shell uname -s | tr A-Z a-z)

.PHONY: all
all:

all: build

build: example

BUILDFLAGS = GOFLAGS=-mod=vendor
example:
	$Q $(BUILDFLAGS) $(GO) build \
		-ldflags '-X main.BuildVersion=$(APPVERSION)' \
		-o $(EXAMPLEBIN) ./cmd/

# Tests
TESTFLAGS = -race -v
TESTSUITE = ./...
.PHONY: test
test:
	$Q $(BUILDFLAGS) $(GO) test $(TESTFLAGS) `$(BUILDFLAGS) $(GO) list $(TESTSUITE)`

# Misc
.PHONY: clean
clean: ; $(info $(M) cleaning...) @ ## Clean up everything
	$Q rm -f $(EXAMPLEBIN)
	$Q rm -f bin/example
