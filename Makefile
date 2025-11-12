.PHONY: help build test lint clean run-dev stop-dev proto docker-up docker-down build-ui dev-ui lint-ui clean-ui

# Default target
help:
	@echo "CareFlow-Mini Makefile"
	@echo ""
	@echo "Backend Targets:"
	@echo "  make build       - Build all backend services"
	@echo "  make test        - Run all backend tests"
	@echo "  make lint        - Run backend linters"
	@echo "  make proto       - Generate protobuf code"
	@echo "  make run-dev     - Start local dev environment (backend + docker)"
	@echo "  make stop-dev    - Stop local dev services"
	@echo "  make docker-up   - Start Docker Compose stack"
	@echo "  make docker-down - Stop Docker Compose stack"
	@echo "  make clean       - Clean backend build artifacts"
	@echo ""
	@echo "Frontend Targets (Vue 3 UI):"
	@echo "  make dev-ui      - Start Vue dev server (port 3000)"
	@echo "  make build-ui    - Build Vue production bundle"
	@echo "  make lint-ui     - Lint Vue code"
	@echo "  make clean-ui    - Clean Vue artifacts"

# Build all services
build:
	@echo "Building all services..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./cmd/api-gateway
	@go build -o bin/patient-svc ./cmd/patient-svc
	@go build -o bin/appointment-svc ./cmd/appointment-svc
	@go build -o bin/lab-adapter ./cmd/lab-adapter
	@go build -o bin/observation-svc ./cmd/observation-svc
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

# Stop local dev services
stop-dev:
	@echo "Stopping local development services..."
	@chmod +x scripts/stop-services.sh
	@./scripts/stop-services.sh

# Run local dev environment
run-dev: docker-up build
	@echo "Starting local development environment..."
	@chmod +x scripts/run-services.sh scripts/stop-services.sh
	@./scripts/run-services.sh
	@echo ""
	@echo "Development environment is ready!"
	@echo "API Gateway: http://localhost:8080"
	@echo "To stop services: make stop-dev"

# Frontend (Vue 3) targets

# Start Vue development server
dev-ui:
	@echo "Starting Vue development server..."
	@cd web && npm run dev

# Build Vue production bundle
build-ui:
	@echo "Building Vue production bundle..."
	@cd web && npm run build

# Lint Vue code
lint-ui:
	@echo "Linting Vue code..."
	@cd web && npm run lint

# Clean Vue build artifacts
clean-ui:
	@echo "Cleaning Vue artifacts..."
	@cd web && npm run clean || true
	@rm -rf web/dist web/node_modules/.vite web/coverage
