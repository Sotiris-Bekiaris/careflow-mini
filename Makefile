.PHONY: help build test lint clean run-dev proto docker-up docker-down

# Default target
help:
	@echo "CareFlow-Mini Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build all services"
	@echo "  make test        - Run all tests"
	@echo "  make lint        - Run linters"
	@echo "  make proto       - Generate protobuf code"
	@echo "  make run-dev     - Start local dev environment"
	@echo "  make docker-up   - Start Docker Compose stack"
	@echo "  make docker-down - Stop Docker Compose stack"
	@echo "  make clean       - Clean build artifacts"

# Build all services
build:
	@echo "Building all services..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./cmd/api-gateway
	@go build -o bin/patient-svc ./cmd/patient-svc
	@go build -o bin/appointment-svc ./cmd/appointment-svc
	@go build -o bin/lab-adapter ./cmd/lab-adapter
	@go build -o bin/notify-svc ./cmd/notify-svc
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run linters
lint:
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@golangci-lint run ./... || true

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@cd proto && buf generate

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Start Docker Compose stack
docker-up:
	@echo "Starting Docker Compose stack..."
	@docker-compose -f deploy/compose/docker-compose.yml up -d

# Stop Docker Compose stack
docker-down:
	@echo "Stopping Docker Compose stack..."
	@docker-compose -f deploy/compose/docker-compose.yml down

# Run local dev environment
run-dev: docker-up
	@echo "Starting local development environment..."
	@echo "Services will be available soon..."
