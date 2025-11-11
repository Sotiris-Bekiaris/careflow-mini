# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CareFlow-Mini is a healthcare-focused microservices demonstration built in Go 1.24+. It showcases FHIR R4 compliance, HL7 v2.x message processing, event-driven architecture, and comprehensive observability in a production-ready pattern.

**Core Patient Journey**: Register patient → Create appointment → Ingest lab result (HL7→FHIR) → Unified patient view → Event streaming

## Essential Commands

### Building and Testing
```bash
# Build all services
make build

# Run all tests with race detection and coverage
make test

# Run tests for a specific package
go test -v ./internal/patient/...

# Run a single test
go test -v -run TestServiceName_CreatePatient ./internal/patient/

# Lint and format
make lint
```

### Development Environment
```bash
# Start infrastructure (Postgres, NATS, observability stack)
make docker-up

# Start all services in development mode (builds + starts infrastructure)
make run-dev

# Stop all local development services
make stop-dev

# Stop Docker Compose stack
make docker-down
```

### Protocol Buffers
```bash
# Generate protobuf code (requires buf CLI)
make proto
```

### Database Operations
```bash
# Connect to PostgreSQL
docker exec -it careflow-postgres psql -U careflow -d careflow

# Initialize database schema
docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/init.sql

# Seed sample data
docker exec -i careflow-postgres psql -U careflow -d careflow < scripts/db/seed.sql
```

### Observability and Monitoring
- **Jaeger (Traces)**: http://localhost:16686
- **Prometheus (Metrics)**: http://localhost:9090
- **Grafana (Dashboards)**: http://localhost:3000 (admin/admin)
- **NATS Monitoring**: http://localhost:8222
- **API Gateway**: http://localhost:8080
- **Patient Service Health**: gRPC health check on port 50051
- **Appointment Service Health**: gRPC health check on port 50052

## Architecture

### Service Communication Pattern
- **External API**: REST/JSON following FHIR R4 shapes via API Gateway (port 8080)
- **Internal Services**: gRPC using Protocol Buffers
- **Event Bus**: NATS JetStream for asynchronous domain events

### Microservices Structure
```
api-gateway (REST) → gRPC services → Postgres
                  ↓
                NATS (events)
                  ↓
          Event Consumers (lab-adapter, notify-svc)
```

**Services**:
- `api-gateway`: Public REST API, translates REST→gRPC (port 8080)
- `patient-svc`: Patient CRUD operations (gRPC, port 50051)
- `appointment-svc`: Appointment scheduling (gRPC, port 50052)
- `lab-adapter`: HL7 ORU^R01 parser → FHIR Observation mapper (worker)
- `notify-svc`: Event consumer for notifications (subscriber)

### Data Storage Strategy
- **PostgreSQL**: Primary store for FHIR resources (Patient, Appointment, Observation) stored as JSONB
- **NATS JetStream**: Persistent event streaming with subjects like `CAREFLOW_EVENTS.patient.created`
- **Event Pattern**: Services publish domain events; consumers subscribe with durable consumers

### Key Packages
- `pkg/db`: PostgreSQL connection pooling with pgx/v5
- `pkg/events`: NATS publisher/subscriber with JetStream integration
- `pkg/fhir`: FHIR R4 resource types (Patient, Appointment, Observation)
- `pkg/hl7`: HL7 v2.x message parsing and FHIR mapping
- `pkg/observability`: OpenTelemetry tracing and metrics initialization
- `internal/patient`: Patient service business logic (service, repository, handler layers)

### Module Path
All imports use: `github.com/Sotiris-Bekiaris/careflow-mini`

## Code Architecture Patterns

### Service Layer Pattern
Services follow a three-layer architecture:
1. **Handler** (`internal/*/handler.go`): gRPC request/response handling
2. **Service** (`internal/*/service.go`): Business logic with OpenTelemetry tracing
3. **Repository** (`internal/*/repository.go`): Database operations with pgx

Example from patient service:
- Service publishes events via `events.Publisher` after successful operations
- Each service method creates spans: `tracer.Start(ctx, "patient.CreatePatient")`
- Repository methods accept context for tracing and transactions

### Event-Driven Communication
Events are published after successful database writes:
```go
// Pattern used in service layer
event := events.Event{
    Type: events.EventTypePatientCreated,
    AggregateID: patientID,
    AggregateType: "Patient",
    Data: patientData,
}
publisher.Publish(event)
```

Subscribers use durable consumers with manual acknowledgment:
- ACK on successful processing
- NACK for redelivery on errors

### Configuration Pattern
Services use environment variables with sensible defaults:
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- NATS: `NATS_URL` (default: `nats://localhost:4222`)
- OpenTelemetry: `OTEL_EXPORTER_OTLP_ENDPOINT` (default: `http://localhost:4318`)

### Testing Patterns
- **Table-driven tests**: All test files use table-driven approach
- **Testcontainers**: Integration tests spin up real Postgres/NATS containers
- **Mocks**: Repository interfaces are mocked for service layer tests
- **Golden tests**: HL7→FHIR mapping uses golden file comparisons

## Healthcare-Specific Context

### FHIR R4 Resources
The system implements simplified FHIR R4 resource types:
- **Patient**: Demographics, identifiers, contact information
- **Appointment**: Scheduling with status (proposed, pending, booked, cancelled)
- **Observation**: Lab results mapped from HL7

Resources are stored as JSONB in Postgres for schema flexibility while maintaining queryability.

### HL7 v2.x Integration
Lab Adapter processes HL7 ORU^R01 (Observation Result) messages:
1. Parse HL7 segments (MSH, PID, OBR, OBX)
2. Map to FHIR Observation resource
3. Store in database
4. Publish `observation.created` event

### Observability Requirements
All services must:
- Create spans for each operation using OpenTelemetry tracer
- Propagate context through gRPC calls (automatic with interceptors)
- Expose health check endpoints
- Record custom metrics for business operations

## Development Workflow

### Adding a New Service
1. Create `cmd/service-name/main.go` with service initialization
2. Create `internal/service-name/` with handler, service, repository layers
3. Define protobuf contract in `proto/service-name/v1/`
4. Run `make proto` to generate gRPC code
5. Add service to Docker Compose in `deploy/compose/docker-compose.yml`
6. Update Makefile build target
7. Add health checks and observability

### Adding a New Event Type
1. Define event type constant in `pkg/events/events.go`
2. Publish event in service layer after successful operation
3. Create subscriber in consumer service
4. Ensure NATS stream subjects cover the new event pattern

### Working with Protocol Buffers
- Contracts defined in `proto/` directory
- Use `buf` for linting and breaking change detection
- Generated code goes to `proto/*/v1/*.pb.go`
- Never edit generated files directly

## Important Notes

- All database operations use `context.Context` for tracing and cancellation
- NATS JetStream requires manual ACK/NACK - always acknowledge messages
- Patient data is PII - logs should redact sensitive fields
- Services should gracefully handle shutdown signals (SIGTERM, SIGINT)
- Connection pools (DB, NATS) must be closed on shutdown
- Test files include comprehensive table-driven tests for all service methods
