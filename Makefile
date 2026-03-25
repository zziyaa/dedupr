
# APP_NAME is read from wails.json to keep a single source of truth.
# Requires jq: brew install jq
APP_NAME := $(shell jq -r '.name' wails.json)

VERSION := $(shell cat VERSION)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -X 'main.version=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)'

LDFLAGS_DEV := $(LDFLAGS) -X 'main.buildType=dev'
LDFLAGS_RELEASE := $(LDFLAGS) -X 'main.buildType=prod'

OUTPUT_DIR := build/bin
DIST_DIR   := dist
DMG_PATH   := $(DIST_DIR)/$(APP_NAME).dmg

.PHONY: all build dev clean fmt lint test check dmg

all: build

build:
	@echo "Building with release flags..."
	wails build -ldflags "$(LDFLAGS_RELEASE)" -platform darwin/arm64
	@/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $(VERSION)" \
		$(OUTPUT_DIR)/$(APP_NAME).app/Contents/Info.plist

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

dmg: build
	@echo "Creating DMG..."
	@mkdir -p $(DIST_DIR)
	@rm -f $(DMG_PATH)
	create-dmg --overwrite --no-version-in-filename \
		--identity="$(DEVELOPER_ID_APP)" \
		$(OUTPUT_DIR)/$(APP_NAME).app \
		$(DIST_DIR)/
	@mv $(DIST_DIR)/$(APP_NAME).dmg $(DMG_PATH) 2>/dev/null || true
	@echo "DMG ready: $(DMG_PATH)"
