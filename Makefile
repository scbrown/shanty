.PHONY: build install clean test

BINARY := shanty
BUILD_DIR := .
INSTALL_DIR := $(HOME)/.local/bin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/scbrown/shanty/internal/cmd.Version=$(VERSION) \
           -X github.com/scbrown/shanty/internal/cmd.Commit=$(COMMIT) \
           -X github.com/scbrown/shanty/internal/cmd.BuildTime=$(BUILD_TIME)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/shanty

install: build
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_DIR)"

test:
	go test ./...

clean:
	rm -f $(BUILD_DIR)/$(BINARY)
