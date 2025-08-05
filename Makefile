# Variables
BINARY_DIR := bin
DOCKER_REGISTRY := your-registry
VERSION := $(shell git describe --tags --always --dirty)
SERVICES := api-gateway queue-manager worker monitor
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: all build test clean

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed 's/## //'

## build: Build all services
build: $(SERVICES)

$(SERVICES):
	@echo "Building $@..."
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$@ ./cmd/$@

## test: Run all tests with coverage
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@$(GOLINT) run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -w $(GO_FILES)

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "Tidying modules..."
	@$(GOMOD) tidy

## mod-download: Download dependencies
mod-download:
	@echo "Downloading dependencies..."
	@$(GOMOD) download

## proto-gen: Generate protobuf code
proto-gen:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

## docker-build: Build Docker images for all services
docker-build:
	@for service in $(SERVICES); do \
		echo "Building Docker image for $$service..."; \
		docker build -f build/docker/$$service.Dockerfile -t $(DOCKER_REGISTRY)/$$service:$(VERSION) .; \
	done

## docker-push: Push Docker images
docker-push:
	@for service in $(SERVICES); do \
		echo "Pushing Docker image for $$service..."; \
		docker push $(DOCKER_REGISTRY)/$$service:$(VERSION); \
	done

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@docker-compose -f deployments/docker-compose.yml up -d

## docker-compose-down: Stop services
docker-compose-down:
	@docker-compose -f deployments/docker-compose.yml down

## migrate-up: Run database migrations
migrate-up:
	@echo "Running migrations..."
	@./scripts/migrate.sh up

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@./scripts/migrate.sh down

## setup: Setup development environment
setup: mod-download
	@echo "Setting up development environment..."
	@./scripts/setup.sh

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR) coverage.* vendor/

## ci: Run CI pipeline
ci: fmt lint vet test

## run-local: Run all services locally
run-local: build docker-compose-up
	@echo "Services are running. Check http://localhost:8080/health"

.DEFAULT_GOAL := help