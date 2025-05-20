.PHONY: build clean test test-coverage test-coverage-html lint run install

# Binary name
BINARY_NAME=gh-notif

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w"

# Coverage directory
COVERAGE_DIR=coverage

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf $(COVERAGE_DIR)

test:
	$(GOTEST) -v ./...

test-coverage:
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v ./... -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out

test-coverage-html: test-coverage
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o=$(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

lint:
	$(GOLINT) run

run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

deps:
	$(GOGET) -u
	$(GOMOD) tidy

install:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	mv $(BINARY_NAME) $(GOPATH)/bin/

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe -v

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 -v

build-all: build-linux build-windows build-darwin
