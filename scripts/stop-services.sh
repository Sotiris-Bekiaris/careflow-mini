#!/bin/bash
# Stop all CareFlow-Mini services

echo "Stopping CareFlow-Mini services..."

# Check if logs directory exists
if [ ! -d "logs" ]; then
    echo "No services appear to be running (no logs directory found)"
    exit 0
fi

# Stop all services by reading PID files
for pidfile in logs/*.pid; do
    if [ -f "$pidfile" ]; then
        service_name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile")

        if kill -0 "$pid" 2>/dev/null; then
            echo "Stopping $service_name (PID: $pid)..."
            kill "$pid"
        else
            echo "$service_name (PID: $pid) is not running"
        fi

        rm "$pidfile"
    fi
done

echo ""
echo "All services stopped!"
echo ""
echo "Logs are preserved in the logs/ directory"
