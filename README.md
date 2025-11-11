# CareFlow-Mini

A production-ready healthcare microservices reference architecture demonstrating modern cloud-native engineering patterns, FHIR R4 compliance, and comprehensive observability.

## Overview

CareFlow-Mini implements a patient journey from registration through appointment scheduling and lab result ingestion. It showcases real-world patterns for building scalable, observable microservices: gRPC for internal communication, REST API for external clients, PostgreSQL for storage, NATS for events, and OpenTelemetry for tracing.

Ideal for learning distributed systems design or as a reference for healthcare platform development.

## Quick Start

### Backend Services

```bash
git clone https://github.com/Sotiris-Bekiaris/careflow-mini.git
cd careflow-mini

make docker-up    # Start infrastructure (Postgres, NATS, observability)
make run-dev      # Build and run all backend services
```

### Frontend (Vue 3)

In a **new terminal**, from the project root:

```bash
cd web
npm install      # Install dependencies
npm run dev      # Start development server on http://localhost:3000
```

### Access the System

| Component | URL | Credentials |
|-----------|-----|-------------|
| **Frontend** | http://localhost:3000 | - |
| **API Gateway** | http://localhost:8080 | - |
| **Jaeger (Tracing)** | http://localhost:16686 | - |
| **Prometheus (Metrics)** | http://localhost:9090 | - |
| **Grafana (Dashboards)** | http://localhost:3000 | admin/admin |

> **Note**: The frontend uses CORS to communicate with the API. During development, the API Gateway allows requests from `http://localhost:3000` by default. See [Configuration](#configuration) to customize allowed origins.

## Architecture

```mermaid
graph TB
    subgraph clients["Client Applications"]
        rest["REST/FHIR JSON"]
    end

    subgraph api["API Gateway (8080)"]
        gateway["HTTP Router<br/>FHIR ↔ Proto Converter<br/>OpenTelemetry Middleware"]
    end

    subgraph services["Microservices"]
        patient["Patient Service<br/>Port 50051"]
        appt["Appointment Service<br/>Port 50052"]
        obs["Observation Service<br/>Port 50055"]
        lab["Lab Adapter<br/>Port 50053"]
        notify["Notify Service<br/>Port 50054"]
    end

    subgraph data["Data & Events"]
        pg["PostgreSQL 15<br/>FHIR Resources"]
        nats["NATS JetStream<br/>Domain Events"]
    end

    subgraph observability["Observability"]
        otel["OpenTelemetry"]
        jaeger["Jaeger"]
        prom["Prometheus"]
        grafana["Grafana"]
    end

    clients -->|REST| api
    api -->|gRPC| patient
    api -->|gRPC| appt
    api -->|gRPC| obs

    patient -->|Read/Write| pg
    appt -->|Read/Write| pg
    obs -->|Read/Write| pg
    lab -->|Read/Write| pg

    patient -->|Events| nats
    appt -->|Events| nats
    obs -->|Events| nats
    notify -->|Subscribe| nats

    patient -.->|Traces| otel
    appt -.->|Traces| otel
    obs -.->|Traces| otel
    lab -.->|Traces| otel
    notify -.->|Traces| otel

    otel -->|OTLP| jaeger
    otel -->|Metrics| prom
    prom -->|Scrape| grafana
```

## Services

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8080 | REST API, routes to gRPC services, traces requests |
| Patient Service | 50051 | CRUD operations for patients |
| Appointment Service | 50052 | Appointment scheduling and cancellation |
| Observation Service | 50055 | Lab results and clinical measurements |
| Lab Adapter | 50053 | HL7 v2.x parser, maps to FHIR Observation |
| Notify Service | 50054 | Event consumer for notifications |

## Technology Stack

- **Language**: Go 1.24+
- **Service Communication**: gRPC + Protocol Buffers (internal), REST + JSON (external)
- **Database**: PostgreSQL 15 with JSONB columns
- **Events**: NATS JetStream with durable consumers
- **Tracing**: OpenTelemetry → Jaeger
- **Metrics**: Prometheus + Grafana
- **Testing**: testcontainers-go for integration tests
- **Frontend**: Vue 3 + TypeScript + Pinia + Vuetify
- **Containerization**: Docker, Docker Compose, Helm

## Development

### Requirements

- **Backend**: Go 1.24+, Docker, Docker Compose, Make, buf (for protobuf)
- **Frontend**: Node.js 18+, npm or yarn

### Backend Build Targets

```bash
make build        # Compile all services
make test         # Run tests with race detection
make lint         # Format and lint
make proto        # Generate protobuf code
make run-dev      # Build and run all backend services
make stop-dev     # Stop all services and clean up
```

### Frontend Development

```bash
cd web

# Install dependencies
npm install

# Development server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build locally
npm run preview

# Linting and type checking
npm run lint
npm run type-check
```

### Full Local Setup

```bash
# Terminal 1: Start infrastructure and backend services
make docker-up
make run-dev

# Terminal 2: Start frontend dev server
cd web
npm install
npm run dev

# Access at http://localhost:3000
```

### Testing

```bash
# All tests
go test -v -race -cover ./...

# Specific package
go test -v ./internal/patient/...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

```
cmd/                    # Service entry points
internal/               # Service layers (handler, service, repository)
pkg/                    # Shared packages (db, events, fhir, hl7, observability)
proto/                  # Protocol Buffer definitions
web/                    # Vue 3 frontend
deploy/                 # Docker Compose, Helm charts
scripts/                # DB initialization, demo scripts
.github/workflows/      # CI/CD pipeline
```

## Data Model

**FHIR Resources** (stored as JSONB in PostgreSQL):
- Patient - Demographics, identifiers, contact info
- Appointment - Scheduled encounters with status
- Observation - Lab results and clinical measurements

**Domain Events** (published to NATS):
- `patient.created`, `patient.updated`, `patient.deleted`
- `appointment.created`, `appointment.cancelled`
- `observation.created`

## API

REST endpoints follow FHIR conventions:

```bash
# Patients
GET/POST        /fhir/Patient
GET/PUT/DELETE  /fhir/Patient/{id}

# Appointments
GET/POST        /fhir/Appointment
DELETE          /fhir/Appointment/{id}

# Observations
GET/POST        /fhir/Observation
GET             /fhir/Observation/{id}

# Health
GET             /health    # Liveness
GET             /ready     # Readiness
```

Full API docs in `docs/` directory.

## Deployment

### Local Development
```bash
make docker-up    # Start infrastructure (Postgres, NATS, observability)
make run-dev      # Build and run services
```

### Kubernetes
```bash
helm install careflow deploy/helm/ \
  --namespace careflow \
  --create-namespace
```

### Configuration

All services use environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=careflow
DB_PASSWORD=careflow_dev
DB_NAME=careflow

# NATS
NATS_URL=nats://localhost:4222

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# Service addresses (for API Gateway)
PATIENT_SVC_ADDR=localhost:50051
APPOINTMENT_SVC_ADDR=localhost:50052
OBSERVATION_SVC_ADDR=localhost:50055

# CORS (for API Gateway frontend communication)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
```

### Frontend Configuration

The frontend uses Vite and environment variables prefixed with `VITE_`:

```bash
# .env or environment
VITE_API_BASE_URL=http://localhost:8080
```

See `web/.env.example` for all available options.

## Observability

- **Jaeger** (http://localhost:16686) - Distributed tracing across services
- **Prometheus** (http://localhost:9090) - Metrics scraping and querying
- **Grafana** (http://localhost:3000) - Dashboards and visualization
- **Health Checks** - `/health` and `/ready` endpoints, gRPC health protocol

All services automatically instrumented with OpenTelemetry. Traces show request paths through multiple services with latency breakdowns.

## Testing & Quality

- Table-driven unit tests for all business logic
- Integration tests with testcontainers for real dependencies
- CI/CD pipeline: lint → test → buf validation → build → security scan → smoke tests
- 80%+ test coverage across packages

## Security

**Development Focus**: This is a reference implementation. For production healthcare systems:

- Implement TLS/mTLS for all service communication
- Add authentication (OAuth 2.0/OIDC) and authorization (RBAC)
- Enable audit logging for data access
- Encrypt data at rest and in transit
- Implement rate limiting and input validation
- Use secrets management systems (not env vars for production)
- Ensure HIPAA compliance for real patient data

See `docs/security.md` for detailed security architecture.

## Troubleshooting

### Frontend cannot connect to API (CORS errors)

**Problem**: Browser console shows `Access to XMLHttpRequest blocked by CORS policy`

**Solution**:
- Ensure API Gateway is running on port 8080
- Verify frontend origin is in `CORS_ALLOWED_ORIGINS` environment variable
- Default allows `http://localhost:3000` and `http://127.0.0.1:3000`
- For production, set: `CORS_ALLOWED_ORIGINS=https://your-domain.com`

### Frontend shows "Cannot GET /" or 404 errors

**Problem**: Frontend assets not served after `npm run build`

**Solution**:
- In development: Use `npm run dev` (Vite dev server)
- For production: Serve `web/dist/` directory with your web server
- Configure web server to redirect all routes to `index.html` (Vue Router requirement)

### Services fail to start

**Problem**: `Connection refused` or `port already in use`

**Solution**:
```bash
# Stop all running services
make stop-dev

# Clean up and restart
make docker-down
make docker-up
make run-dev
```

### Database connection issues

**Problem**: Services timeout connecting to PostgreSQL

**Solution**:
```bash
# Check if Postgres is running
docker ps | grep careflow-postgres

# View logs
docker logs careflow-postgres

# Verify connection
docker exec -it careflow-postgres psql -U careflow -d careflow -c "SELECT 1"
```

### NATS connection issues

**Problem**: `Failed to connect to NATS`

**Solution**:
```bash
# Check if NATS is running
curl http://localhost:8222/varz | jq .

# View NATS logs
docker logs careflow-nats
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow Go conventions, write tests
4. Ensure `make test` and `make lint` pass
5. Submit PR with clear description

Use conventional commits for messages.

## Documentation

- **Architecture Decision Records** - `docs/adr/` - Design rationales and trade-offs
- **Runbook** - `docs/runbook.md` - Operational procedures
- **Deployment** - `docs/deployment.md` - Cloud deployment patterns
- **Security** - `docs/security.md` - Security architecture and recommendations

## Performance

- Simple operations (patient lookup): 20-50ms
- Complex workflows (create + publish event): 100-200ms
- HL7 parsing: 5-10ms per message

Horizontal scaling: Deploy multiple service instances behind a load balancer. Stateless services scale linearly. Database connection pooling and NATS handle increased throughput.

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

## Support

- **Issues**: Report bugs and request features on GitHub
- **Discussions**: Ask questions and discuss architecture on GitHub Discussions
- **Security**: Report vulnerabilities privately to maintainers

## Roadmap

- Expanded FHIR resource support (Medication, Procedure, Condition, etc.)
- Additional HL7 message types (ADT, ORM, etc.)
- Multi-tenant support
- Machine learning for anomaly detection
- Reference implementations in Rust and Python