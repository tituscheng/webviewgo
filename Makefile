.PHONY: all build test test-unit test-integration lint fmt vet cover clean

GO ?= go
GOLANGCI_LINT ?= golangci-lint

all: fmt vet lint test

build:
	$(GO) build ./...

test:
	$(GO) test -race -count=1 ./...

test-unit:
	$(GO) test -race -count=1 -short ./...

test-integration:
	$(GO) test -race -count=1 -tags=integration ./...

lint:
	$(GOLANGCI_LINT) run ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

clean:
	$(GO) clean ./...
	rm -f coverage.out
