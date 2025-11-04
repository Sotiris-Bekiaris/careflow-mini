# ADR-0001: Technology Stack and Architecture Choices

## Status
Accepted

## Context
We need to build a healthcare demonstration system that showcases modern backend engineering practices while handling healthcare-specific requirements including FHIR R4 compliance, HL7 message processing, and HIPAA-relevant security considerations.

## Decision

### 1. Programming Language: Go
**Chosen**: Go 1.22+

**Rationale**:
- Excellent concurrency primitives for handling multiple requests
- Strong standard library with HTTP/JSON support
- Fast compilation and execution
- Extensive healthcare libraries (HL7, FHIR)
- Good gRPC support
- Strong typing and error handling
- Popular in healthcare/enterprise systems

**Alternatives Considered**:
- Java: More verbose, heavier runtime
- Python: Slower execution, weaker typing
- Rust: Steeper learning curve

### 2. API Protocol: REST (external) + gRPC (internal)
**Chosen**: Dual protocol approach

**Rationale**:
- REST/JSON for public-facing API (industry standard, FHIR compatible)
- gRPC for inter-service communication (performance, type safety)
- Clear separation of external vs internal APIs
- Flexibility to expose both when needed

**Alternatives Considered**:
- REST only: Less efficient for internal communication
- gRPC only: Less accessible for external clients
- GraphQL: Overkill for this use case

### 3. Data Store: PostgreSQL
**Chosen**: PostgreSQL 15+

**Rationale**:
- JSONB support for flexible FHIR resource storage
- ACID compliance for healthcare data
- Mature, reliable, widely used
- Good performance for read-heavy workloads
- Rich indexing capabilities

**Alternatives Considered**:
- MongoDB: Less mature for transactions
- MySQL: Weaker JSON support
- DynamoDB: Vendor lock-in, different consistency model

### 4. Message Broker: NATS
**Chosen**: NATS with JetStream

**Rationale**:
- Lightweight and fast
- Built-in persistence with JetStream
- Simple deployment (single binary)
- Good Go client library
- Suitable for event-driven architecture

**Alternatives Considered**:
- Kafka: Heavier, more complex setup
- RabbitMQ: More features but heavier
- Redis Streams: Limited persistence guarantees

### 5. Observability: OpenTelemetry
**Chosen**: OpenTelemetry + Prometheus + Jaeger + Grafana

**Rationale**:
- OpenTelemetry is vendor-neutral standard
- Unified approach to traces, metrics, logs
- Prometheus for metrics (industry standard)
- Jaeger for distributed tracing
- Grafana for visualization
- All open-source, no vendor lock-in

**Alternatives Considered**:
- DataDog: Commercial, vendor lock-in
- New Relic: Commercial, vendor lock-in
- ELK Stack: Heavier, more complex

### 6. Healthcare Standards: FHIR R4 + HL7 v2.x
**Chosen**: FHIR R4 for REST API, HL7 v2.x for legacy integration

**Rationale**:
- FHIR R4 is current healthcare interoperability standard
- RESTful and JSON-based (modern, developer-friendly)
- HL7 v2.x still widely used for lab results
- Both standards have good Go library support

**Alternatives Considered**:
- FHIR STU3: Older version
- HL7 v3: More complex, less adopted
- Custom format: No interoperability

### 7. Deployment: Docker Compose + Kubernetes
**Chosen**: Docker Compose for local dev, Helm for K8s production

**Rationale**:
- Docker Compose: Simple local development
- Kubernetes: Industry standard for production
- Helm: Package manager for K8s
- Easy transition from local to production
- Cloud-agnostic

**Alternatives Considered**:
- Docker Swarm: Less popular, limited ecosystem
- Nomad: Less healthcare industry adoption
- Bare metal: Complex operational overhead

## Consequences

### Positive
- Modern, maintainable stack
- Healthcare industry standards compliance
- Strong observability from day one
- Fast development and deployment cycles
- Good performance characteristics
- Cloud-native and scalable

### Negative
- Learning curve for team members unfamiliar with Go
- Multiple protocols (REST + gRPC) adds complexity
- FHIR R4 is complex and large
- Operational complexity of microservices

### Neutral
- Need to maintain Protocol Buffer definitions
- Multiple datastores to manage (Postgres, NATS)
- Observability stack requires additional infrastructure

## References
- [FHIR R4 Specification](https://hl7.org/fhir/R4/)
- [HL7 v2.x Standard](https://www.hl7.org/implement/standards/product_brief.cfm?product_id=185)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Go gRPC Guide](https://grpc.io/docs/languages/go/)

## Revision History
- 2024-01-15: Initial version
