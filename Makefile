# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

.PHONY: build build-cli build-daemon clean test version release-major release-minor release-patch release-beta help ci

# Version info
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -s -w \
	-X github.com/dployr-io/github.com/dployr-io/dployr/pkg/version.Version=$(VERSION) \
	-X github.com/dployr-io/github.com/dployr-io/dployr/pkg/version.GitCommit=$(COMMIT) \
	-X github.com/dployr-io/github.com/dployr-io/dployr/pkg/version.BuildDate=$(DATE)

CLI_LDFLAGS = $(LDFLAGS) -X github.com/dployr-io/github.com/dployr-io/dployr/pkg/version.Component=dployr
DAEMON_LDFLAGS = $(LDFLAGS) -X github.com/dployr-io/github.com/dployr-io/dployr/pkg/version.Component=dployrd

# Build directory
BUILD_DIR = ./dist

## Build both binaries
build: build-dployr build-dployrd

## Build CLI binary
build-dployr:
	@echo "Building dployr CLI..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(CLI_LDFLAGS)" -o $(BUILD_DIR)/dployr ./cmd/dployr

## Build daemon binary  
build-dployrd:
	@echo "Building dployrd daemon..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(DAEMON_LDFLAGS)" -o $(BUILD_DIR)/dployrd ./cmd/dployrd

## Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./dployr ./dployrd

## Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## Local CI parity
ci:
	@echo "Checking formatting (gofmt -s)..."
	@fmt_out=$$(gofmt -s -l .); if [ -n "$$fmt_out" ]; then echo "Files not formatted:" && echo "$$fmt_out" && exit 1; fi
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running staticcheck..."
	@if ! command -v staticcheck >/dev/null 2>&1; then echo "Installing staticcheck..."; GO111MODULE=on go install honnef.co/go/tools/cmd/staticcheck@latest; fi
	@"$$(go env GOPATH)"/bin/staticcheck ./...
	@echo "Building..."
	@go build ./...
	@echo "Running tests (race)..."
	@CGO_ENABLED=1 go test -race -count=1 ./...

## Show version info 
version: build
	@echo "CLI version:"
	@$(BUILD_DIR)/dployr version
	@echo ""
	@echo "Daemon version:"
	@$(BUILD_DIR)/dployrd --version

## Create major release
release-major:
	@./scripts/release.sh major

## Create minor release  
release-minor:
	@./scripts/release.sh minor

## Create patch release
release-patch:
	@./scripts/release.sh patch

## Create beta release
release-beta:
	@./scripts/release.sh patch --beta

## Show help
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'