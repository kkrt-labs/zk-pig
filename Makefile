# Git based version
VERSION ?= $(shell git describe --tags)

# Veriables used for building with goreleaser
MODULE_NAME := github.com/kkrt-labs/zk-pig

GOPATH ?= $(shell go env GOPATH)

# List of effective go files
GOFILES := $(shell find . -name '*.go' -not -path "./vendor/*" -not -path "./tests/*" | egrep -v "^\./\.go" | grep -v _test.go)

# List of packages except testsutils
PACKAGES ?= $(shell go list ./... | egrep -v "testutils" )

# Build folder
BUILD_FOLDER = build

# Tools versions
GORELEASER_CROSS_VERSION = v1.24.0
GOLANGCI_VERSION = v1.62.0
MOCKGEN_VERSION = v0.5.0

# GOPRIVVATE
GOPRIVATE=

# Test coverage variables
COVERAGE_BUILD_FOLDER = $(BUILD_FOLDER)/coverage

UNIT_COVERAGE_OUT  = $(COVERAGE_BUILD_FOLDER)/ut_cov.out
UNIT_COVERAGE_HTML = $(COVERAGE_BUILD_FOLDER)/ut_index.html

.PHONY: help run mod-tidy test test-race test-lint lint generate-mocks goreleaser-snaptho goreleaser version

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@echo "  help                Show Makefile help message"
	@echo "  run                 Run the main.go file"
	@echo "  mod-tidy            Run go mod tidy command to update go.mod and go.sum files"
	@echo "  test                Run unit tests with coverage"
	@echo "  test-race           Run unit tests with race detector" 
	@echo "  test-lint           Check linting"
	@echo "  lint                Run linter to fix linting issues"
	@echo "  generate-proto      Generate protobuf files"
	@echo "  mockgen-install     Install mockgen command"
	@echo "  generate-mocks      Generate mocks"
	@echo "  goreleaser-snapshot Execute goreleaser with --snapshot flag"
	@echo "  goreleaser          Execute goreleaser"
	@echo "  version             Read version from git tags"

run:
	@go run .

# Run go mod tidy command to update go.mod and go.sum files
mod-tidy:
	@export GOPRIVATE=$(GOPRIVATE) | go mod tidy --compat 1.18

build/coverage:
	@mkdir -p $(COVERAGE_BUILD_FOLDER)

unit-test: build/coverage
	@go test -covermode=count -coverprofile $(UNIT_COVERAGE_OUT) -v $(PACKAGES)

# Run unit tests with coverage
test: unit-test
	@go tool cover -html=$(UNIT_COVERAGE_OUT) -o $(UNIT_COVERAGE_HTML)

# Run unit tests with race detector
test-race:
	@go test -race $(PACKAGES)


test-lint: ## Check linting
	@type golangci-lint >/dev/null 2>&1 && { \
		golangci-lint run -v --timeout 2m0s; \
	} || { \
		docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run -v --timeout 2m0s; \
	}

lint: ## Run linter to fix issues
	@type golangci-lint >/dev/null 2>&1 && { \
		golangci-lint run --fix; \
	} || { \
		docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run --fix; \
	}

generate-proto:
	@protoc --go_out=. --go_opt=paths=source_relative src/prover-input/proto/input.proto

# Install mockgen command
mockgen-install:
	@type mockgen >/dev/null 2>&1 || {   \
		echo "Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION);  \
	}

generate-mocks: mockgen-install
	@go generate ./...

goreleaser-snapshot:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(MODULE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(MODULE_NAME) \
		goreleaser/goreleaser-cross:${GORELEASER_CROSS_VERSION} \
		--clean --skip=publish,validate --snapshot

goreleaser:
	@if [ ! -f ".env" ]; then \
		echo "\033[91m.env is required for release\033[0m";\
		exit 1;\
	fi
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(MODULE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(MODULE_NAME) \
		goreleaser/goreleaser-cross:${GORELEASER_CROSS_VERSION} \
		--clean

# Read version from git tags
# It is used in CI to set the version
version:
	@echo $(VERSION)
