# Makefile for idle-svc – convenience targets for development

BINARY := idle-svc
PKG := github.com/0xabrar/idle-svc

.PHONY: all build install test lint clean docker helm-chart tidy

all: build

## Build the binary for the host platform
build:
	@echo "--> building $(BINARY)"
	go mod tidy
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BINARY) .

## Install the binary to $(shell go env GOBIN) (or GOPATH/bin)
install:
	@echo "--> installing $(BINARY) to $$(go env GOBIN)"
	GO111MODULE=on go install .

## Run unit tests (none yet, placeholder for future)
test: tidy
	@echo "--> running go test"
	go test ./...

## Run basic linters/vet (requires 'go vet' and optional 'staticcheck')
lint: tidy
	@echo "--> running go vet"
	go vet ./...

## Remove the compiled binary and other build cache
clean:
	@echo "--> cleaning"
	go clean
	@rm -f $(BINARY)

docker:
	@echo "--> building multi-arch docker image"
	docker build -t idle-svc:latest -f Dockerfile .

helm-chart:
	@echo "--> packaging helm chart"
	helm lint chart/idle-svc
	helm package chart/idle-svc -d ./dist

tidy:
	go mod tidy 