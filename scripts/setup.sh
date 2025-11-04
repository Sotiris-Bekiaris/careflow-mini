#!/bin/bash
# CareFlow-Mini Development Environment Setup Script

set -e

echo "========================================="
echo "CareFlow-Mini Environment Setup"
echo "========================================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "✗ Go is not installed. Please install Go 1.22 or later."
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo "✓ Go installed: ${GO_VERSION}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "✗ Docker is not installed. Please install Docker."
    exit 1
fi
echo "✓ Docker installed"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "✗ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi
echo "✓ Docker Compose installed"

# Check Make
if ! command -v make &> /dev/null; then
    echo "✗ Make is not installed. Please install Make."
    exit 1
fi
echo "✓ Make installed"

# Optional: Check Buf
if ! command -v buf &> /dev/null; then
    echo "⚠ Buf CLI is not installed. Protocol Buffer generation may not work."
    echo "  Install from: https://docs.buf.build/installation"
else
    echo "✓ Buf CLI installed"
fi

echo ""
echo "Installing Go dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies installed"

echo ""
echo "Starting infrastructure services..."
docker-compose -f deploy/compose/docker-compose.yml up -d
echo "✓ Infrastructure services started"

echo ""
echo "Waiting for services to be ready..."
sleep 5

# Check PostgreSQL
if docker exec careflow-postgres pg_isready -U careflow > /dev/null 2>&1; then
    echo "✓ PostgreSQL is ready"
else
    echo "⚠ PostgreSQL is not ready yet, you may need to wait a bit longer"
fi

# Check NATS
if curl -s http://localhost:8222/healthz > /dev/null 2>&1; then
    echo "✓ NATS is ready"
else
    echo "⚠ NATS is not ready yet, you may need to wait a bit longer"
fi

echo ""
echo "Initializing database..."
if [ -f scripts/db/init.sql ]; then
    docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/init.sql
    echo "✓ Database schema created"
else
    echo "⚠ Database initialization script not found"
fi

echo ""
echo "========================================="
echo "Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Build services:    make build"
echo "  2. Run tests:         make test"
echo "  3. Generate protos:   make proto"
echo "  4. Start services:    Run each service binary in bin/"
echo ""
echo "Useful URLs:"
echo "  API Gateway:    http://localhost:8080"
echo "  Jaeger UI:      http://localhost:16686"
echo "  Prometheus:     http://localhost:9090"
echo "  Grafana:        http://localhost:3000"
echo "  NATS Monitor:   http://localhost:8222"
echo ""
