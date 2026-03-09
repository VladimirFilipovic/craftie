VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BINARY := craftie
INSTALL_DIR := $(HOME)/.local/bin
BUILD_DIR := ./_builds

.PHONY: build install clean test run

run: build
	$(BUILD_DIR)/$(BINARY) $(ARGS)

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/...

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...
