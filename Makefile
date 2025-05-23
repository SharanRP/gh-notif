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
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY?=$(shell whoami)
LDFLAGS=-ldflags "-s -w -X main.versionString=$(VERSION) -X main.commitString=$(COMMIT) -X main.dateString=$(DATE) -X main.builtByString=$(BUILT_BY)"

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

# Additional targets for CI/CD
.PHONY: security vuln completions docs docker-build release-dry-run help

security: ## Run security scan
	gosec ./...

vuln: ## Check for vulnerabilities
	govulncheck ./...

completions: build ## Generate shell completions
	mkdir -p completions
	./$(BINARY_NAME) completion bash > completions/$(BINARY_NAME).bash
	./$(BINARY_NAME) completion zsh > completions/$(BINARY_NAME).zsh
	./$(BINARY_NAME) completion fish > completions/$(BINARY_NAME).fish
	./$(BINARY_NAME) completion powershell > completions/$(BINARY_NAME).ps1

docs: build ## Generate documentation
	mkdir -p docs/man
	./$(BINARY_NAME) man --output-dir docs/man

docker-build: ## Build Docker image
	docker build -t gh-notif:latest .

docker-build-dev: ## Build development Docker image
	docker build -f Dockerfile.dev -t gh-notif:dev .

release-dry-run: ## Run GoReleaser in dry-run mode
	goreleaser release --snapshot --skip-publish --clean

check: lint test security vuln ## Run all checks

help: ## Show this help message
	@echo "gh-notif Makefile"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
