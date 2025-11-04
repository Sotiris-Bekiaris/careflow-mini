#!/bin/bash
# Run all CareFlow-Mini services in the background

set -e

echo "Starting CareFlow-Mini services..."

# Set environment variables for local development
export PORT=8080
export PATIENT_SVC_ADDR=localhost:50051
export APPOINTMENT_SVC_ADDR=localhost:50052
export LAB_ADAPTER_ADDR=localhost:50053
export NOTIFY_SVC_ADDR=localhost:50054

# Create logs directory if it doesn't exist
mkdir -p logs

# Function to start a service
start_service() {
    local name=$1
    local binary=$2
    local port=$3

    echo "Starting $name on port $port..."
    nohup "$binary" > "logs/${name}.log" 2>&1 &
    echo $! > "logs/${name}.pid"
}

# Start backend services (gRPC)
start_service "patient-svc" "./bin/patient-svc" "50051"
start_service "appointment-svc" "./bin/appointment-svc" "50052"
start_service "lab-adapter" "./bin/lab-adapter" "50053"
start_service "notify-svc" "./bin/notify-svc" "50054"

# Wait a moment for backend services to start
sleep 2

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
