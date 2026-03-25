
# APP_NAME is read from wails.json to keep a single source of truth.
# Requires jq: brew install jq
APP_NAME := $(shell jq -r '.name' wails.json)

VERSION := $(shell cat VERSION)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -X 'main.version=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)'

LDFLAGS_DEV := $(LDFLAGS) -X 'main.buildType=dev'
LDFLAGS_RELEASE := $(LDFLAGS) -X 'main.buildType=prod'

OUTPUT_DIR := build/bin

.PHONY: all build dev clean fmt lint test check

all: build

build:
	@echo "Building with release flags..."
	wails build -ldflags "$(LDFLAGS_RELEASE)" -platform darwin/arm64

dev:
	@echo "Starting dev server..."
	wails dev -ldflags "$(LDFLAGS_DEV)"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...
	@go vet ./...

test:
	@echo "Running tests..."
	@go test -v ./...

check: fmt lint test
