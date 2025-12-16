.PHONY: build run clean test install help

# Binary name
BINARY_NAME=trash

# Build variables
VERSION?=0.1.0
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/artemisfowl/trash/cmd.Version=$(VERSION) -X github.com/artemisfowl/trash/cmd.BuildDate=$(BUILD_DATE) -X github.com/artemisfowl/trash/cmd.GitCommit=$(GIT_COMMIT)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

run: ## Run the application
	go run main.go

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

test: ## Run tests
	go test -v ./...

install: ## Install the binary to $GOPATH/bin
	go install $(LDFLAGS) .

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

.DEFAULT_GOAL := help
