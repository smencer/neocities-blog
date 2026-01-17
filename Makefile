BINARY_NAME := neocities
BUILD_DIR := build
GO := go

.PHONY: all build clean install fmt vet test

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...
