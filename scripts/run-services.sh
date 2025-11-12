#!/bin/bash
# Run all CareFlow-Mini services in the background

set -e

echo "Starting CareFlow-Mini services..."

# Set environment variables for local development

# API Gateway
export PORT=8080
export PATIENT_SVC_ADDR=localhost:50051
export APPOINTMENT_SVC_ADDR=localhost:50052
export LAB_ADAPTER_ADDR=localhost:50053
export NOTIFY_SVC_ADDR=localhost:50054
export OBSERVATION_SVC_ADDR=localhost:50055

# Database (default credentials from docker-compose)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=careflow
export DB_PASSWORD=careflow_dev
export DB_NAME=careflow

# NATS
export NATS_URL=nats://localhost:4222

# Observability
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# CORS
export CORS_ALLOWED_ORIGINS=http://localhost:3001,http://127.0.0.1:3001

# Create logs directory if it doesn't exist
mkdir -p logs

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=0
until docker exec careflow-postgres pg_isready -U careflow -d careflow > /dev/null 2>&1; do
    if [ $attempt -ge $max_attempts ]; then
        echo "PostgreSQL failed to become ready after $max_attempts attempts"
        exit 1
    fi
    attempt=$((attempt + 1))
    echo "  Waiting... ($attempt/$max_attempts)"
    sleep 1
done
echo "PostgreSQL is ready!"

# Initialize database schema if not already done
echo "Ensuring database schema is initialized..."
docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/init.sql > /dev/null 2>&1 || true

# Function to start a service
start_service() {
    local name=$1
    local binary=$2
    local port=$3

    echo "Starting $name on port $port..."
    env DB_HOST=$DB_HOST DB_PORT=$DB_PORT DB_USER=$DB_USER DB_PASSWORD=$DB_PASSWORD DB_NAME=$DB_NAME \
        NATS_URL=$NATS_URL OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT \
        CORS_ALLOWED_ORIGINS=$CORS_ALLOWED_ORIGINS \
        nohup "$binary" > "logs/${name}.log" 2>&1 &
    echo $! > "logs/${name}.pid"
}

# Start backend services (gRPC)
start_service "patient-svc" "./bin/patient-svc" "50051"
start_service "appointment-svc" "./bin/appointment-svc" "50052"
start_service "lab-adapter" "./bin/lab-adapter" "50053"
start_service "observation-svc" "./bin/observation-svc" "50055"
start_service "notify-svc" "./bin/notify-svc" "50054"

# Wait a moment for backend services to start
sleep 3

# Start API Gateway (HTTP)
start_service "api-gateway" "./bin/api-gateway" "8080"

echo ""
echo "All services started!"
echo ""
echo "Service PIDs:"
for pidfile in logs/*.pid; do
    if [ -f "$pidfile" ]; then
        service_name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile")
        echo "  $service_name: $pid"
    fi
done

echo ""
echo "Logs are available in the logs/ directory"
echo "API Gateway: http://localhost:8080"
echo ""
echo "To stop all services, run: ./scripts/stop-services.sh"
