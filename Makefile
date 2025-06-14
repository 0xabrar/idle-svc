# Makefile for idle-svc – convenience targets for development

BINARY := idle-svc
PKG := github.com/0xabrar/idle-svc

.PHONY: all build install test lint clean

all: build

## Build the binary for the host platform
build:
	@echo "--> building $(BINARY)"
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BINARY) .

## Install the binary to $(shell go env GOBIN) (or GOPATH/bin)
install:
	@echo "--> installing $(BINARY) to $$(go env GOBIN)"
	GO111MODULE=on go install .

## Run unit tests (none yet, placeholder for future)
test:
	@echo "--> no tests yet (placeholder)"

## Run basic linters/vet (requires 'go vet' and optional 'staticcheck')
lint:
	@echo "--> running go vet"
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not found; skipping"

## Remove the compiled binary and other build cache
clean:
	@echo "--> cleaning"
	go clean
	@rm -f $(BINARY) 