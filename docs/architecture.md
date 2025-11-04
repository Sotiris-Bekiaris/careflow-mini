# CareFlow-Mini Architecture

## Overview

CareFlow-Mini is a microservices-based healthcare demo system demonstrating modern Go backend engineering practices with healthcare-specific features including FHIR R4 compliance and HL7 message processing.

## System Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  API Gateway    │ ◄── REST/JSON (FHIR-style)
│   (Port 8080)   │
└────────┬────────┘
         │ gRPC
    ┌────┴────┬────────────┐
    │         │            │
    ▼         ▼            ▼
┌──────┐  ┌──────┐    ┌──────┐
│Patient│  │Appt. │    │Other │
│  Svc  │  │ Svc  │    │Svcs  │
└───┬───┘  └───┬──┘    └───┬──┘
    │          │            │
    └──────────┴────────────┘
               │
        ┌──────┴──────┐
        │             │
        ▼             ▼
   ┌────────┐    ┌──────┐
   │Postgres│    │ NATS │
   └────────┘    └──┬───┘
                    │
           ┌────────┴────────┐
           │                 │
           ▼                 ▼
      ┌─────────┐      ┌──────────┐
      │   Lab   │      │  Notify  │
      │ Adapter │      │   Svc    │
      └─────────┘      └──────────┘
```

## Components

### Services

#### API Gateway
- **Purpose**: Public-facing REST API
- **Port**: 8080
- **Protocol**: HTTP/REST with FHIR R4 JSON
- **Responsibilities**:
  - Request routing
  - Authentication/Authorization
  - Rate limiting
  - REST to gRPC translation

#### Patient Service
- **Purpose**: Patient CRUD operations
- **Port**: 50051
- **Protocol**: gRPC
- **Data**: Patient resources (FHIR-compliant)

#### Appointment Service
- **Purpose**: Appointment scheduling
- **Port**: 50052
- **Protocol**: gRPC
- **Data**: Appointment resources (FHIR-compliant)

#### Lab Adapter
- **Purpose**: HL7 message processing
- **Type**: Worker/Consumer
- **Responsibilities**:
  - Parse HL7 ORU^R01 messages
  - Map to FHIR Observation resources
  - Emit domain events

#### Notify Service
- **Purpose**: Event-driven notifications
- **Type**: Consumer
- **Responsibilities**:
  - Subscribe to domain events
  - Send notifications (email, SMS, etc.)

### Data Stores

#### PostgreSQL
- Primary data store
- Stores FHIR resources as JSONB
- Schema per service

#### Redis (Optional)
- Cache layer
- Session storage
- Patient summary cache

### Messaging

#### NATS
- Event bus for domain events
- JetStream for persistent messaging
- Pub/Sub pattern

### Observability Stack

#### OpenTelemetry
- Distributed tracing
- Metrics collection
- Unified observability

#### Jaeger
- Trace visualization
- Performance analysis

#### Prometheus
- Metrics storage
- Alerting

#### Grafana
- Metrics visualization
- Dashboard creation

## Data Flow

### Patient Registration Flow
1. Client sends POST to `/fhir/Patient`
2. API Gateway validates and forwards to Patient Service (gRPC)
3. Patient Service stores to Postgres
4. Patient Service publishes `patient.created` event to NATS
5. Response flows back to client

### HL7 Lab Result Flow
1. HL7 ORU^R01 message received by Lab Adapter
2. Message parsed and validated
3. Mapped to FHIR Observation resource
4. Stored in Postgres
5. `observation.created` event published to NATS
6. Notify Service receives event and sends notification

## Technology Stack

- **Language**: Go 1.22+
- **API**: REST (external), gRPC (internal)
- **Data Format**: JSON (FHIR R4), Protocol Buffers
- **Database**: PostgreSQL 15+
- **Messaging**: NATS with JetStream
- **Observability**: OpenTelemetry, Prometheus, Grafana, Jaeger
- **Deployment**: Docker Compose, Kubernetes (Helm)

## Design Principles

1. **Microservices Architecture**: Loosely coupled, independently deployable services
2. **API-First**: Well-defined contracts using Protocol Buffers and FHIR
3. **Event-Driven**: Asynchronous communication via NATS
4. **Observability**: Built-in tracing, metrics, and logging
5. **Healthcare Standards**: FHIR R4 and HL7 v2.x compliance
6. **Testing**: Comprehensive unit, integration, and contract tests
7. **Security**: Authentication, authorization, and PII protection

## Next Steps

- Complete service implementations
- Add authentication/authorization
- Implement full FHIR resource support
- Add comprehensive test coverage
- Set up CI/CD pipeline
- Deploy to Kubernetes
