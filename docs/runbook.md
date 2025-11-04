# CareFlow-Mini Runbook

## Getting Started

### Prerequisites
- Go 1.22+
- Docker and Docker Compose
- Make
- Buf CLI (for Protocol Buffers)

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/careflow-mini.git
   cd careflow-mini
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start infrastructure**
   ```bash
   make docker-up
   ```

4. **Generate Protocol Buffer code**
   ```bash
   make proto
   ```

5. **Build services**
   ```bash
   make build
   ```

6. **Run services locally**
   ```bash
   # Terminal 1: API Gateway
   ./bin/api-gateway

   # Terminal 2: Patient Service
   ./bin/patient-svc

   # Terminal 3: Appointment Service
   ./bin/appointment-svc

   # Terminal 4: Lab Adapter
   ./bin/lab-adapter

   # Terminal 5: Notify Service
   ./bin/notify-svc
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -cover ./...

# Run specific package tests
go test -v ./pkg/fhir/...
```

### Linting

```bash
# Run linters
make lint

# Format code
go fmt ./...
```

## Operations

### Health Checks

- **API Gateway**: http://localhost:8080/health
- **Patient Service**: gRPC health check on port 50051
- **Appointment Service**: gRPC health check on port 50052

### Observability URLs

- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **NATS Monitoring**: http://localhost:8222

### Database Access

```bash
# Connect to PostgreSQL
docker exec -it careflow-postgres psql -U careflow -d careflow

# Common queries
SELECT * FROM patients LIMIT 10;
SELECT * FROM appointments WHERE status = 'booked';
```

### NATS Management

```bash
# Check NATS status
docker exec careflow-nats nats server info

# List streams
docker exec careflow-nats nats stream list

# Subscribe to events
docker exec careflow-nats nats sub "patient.>"
```

## Common Issues

### Issue: Services can't connect to database

**Symptom**: Connection refused errors in logs

**Solution**:
1. Check if PostgreSQL is running: `docker ps | grep postgres`
2. Verify connection string in environment variables
3. Check PostgreSQL logs: `docker logs careflow-postgres`

### Issue: gRPC connection failures

**Symptom**: "connection refused" or timeout errors

**Solution**:
1. Verify service is listening on correct port
2. Check firewall rules
3. Ensure services started in correct order

### Issue: Events not being processed

**Symptom**: Events published but not consumed

**Solution**:
1. Check NATS is running: `docker ps | grep nats`
2. Verify subscriber is connected: Check NATS monitoring UI
3. Check for subscription errors in service logs

## Deployment

### Docker Compose (Local/Dev)

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
docker-compose -f deploy/compose/docker-compose.yml logs -f
```

### Kubernetes (Production)

```bash
# Install with Helm
helm install careflow ./deploy/helm

# Upgrade release
helm upgrade careflow ./deploy/helm

# Uninstall
helm uninstall careflow

# Check pod status
kubectl get pods -l app=careflow-mini

# View logs
kubectl logs -l app=careflow-mini --tail=100 -f
```

## Monitoring & Alerts

### Key Metrics to Monitor

1. **Request Rate**: Requests per second per service
2. **Error Rate**: 4xx and 5xx response rates
3. **Latency**: p50, p95, p99 response times
4. **Database Connection Pool**: Active connections
5. **Event Queue Depth**: NATS queue backlog

### Common Alerts

- High error rate (>5% 5xx errors)
- High latency (p95 > 1s)
- Database connection pool exhausted
- NATS queue depth growing
- Pod crash loops

## Backup & Recovery

### Database Backup

```bash
# Backup PostgreSQL
docker exec careflow-postgres pg_dump -U careflow careflow > backup.sql

# Restore
cat backup.sql | docker exec -i careflow-postgres psql -U careflow careflow
```

### Disaster Recovery

1. Restore database from latest backup
2. Redeploy services from known good image
3. Verify health checks pass
4. Resume event processing from last checkpoint

## Security

### Secrets Management

- Never commit secrets to git
- Use environment variables for configuration
- In production, use secrets management (Vault, K8s Secrets, etc.)

### PII Handling

- All patient data is considered PII
- Logs should redact sensitive fields
- Use encryption at rest for database
- Use TLS for all network communication

## Troubleshooting Commands

```bash
# Check service logs
docker logs careflow-api-gateway
kubectl logs -l app=patient-svc

# Check resource usage
docker stats
kubectl top pods

# Network debugging
docker exec careflow-api-gateway netstat -tlnp
kubectl exec -it pod-name -- sh

# Database queries
docker exec -it careflow-postgres psql -U careflow -c "SELECT version();"

# NATS debugging
docker exec careflow-nats nats stream info
```

## References

- [Architecture Documentation](./architecture.md)
- [ADR-0001: Technology Choices](./ADR-0001-technology-choices.md)
- [FHIR R4 API Documentation](https://hl7.org/fhir/R4/)
