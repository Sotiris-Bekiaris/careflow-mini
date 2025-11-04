Project: CareFlow-Mini

Goal:
Demonstrate real-world Go backend engineering with healthcare flavor — clean microservices, gRPC/REST, observability, CI/CD, and testing discipline.

⸻

Core Concept

A minimal “patient journey” demo: 1. Register a patient 2. Create an appointment 3. Ingest a lab result (HL7 → FHIR) 4. Expose a unified patient view via REST 5. Stream domain events to other services

⸻

Services (all Go)
• api-gateway (REST) – Public REST → forwards to gRPC services (FHIR-style routes).
• patient-svc (gRPC) – CRUD for Patients.
• appointment-svc (gRPC) – Create & list appointments.
• lab-adapter (worker) – Parses HL7 ORU^R01 → maps to FHIR Observation → emits events.
• notify-svc (consumer) – Subscribes to ObservationCreated → logs notifications.

⸻

Data & Messaging
• Postgres – Stores FHIR-shaped Patient / Appointment / Observation as JSONB.
• NATS – Lightweight event bus for domain events.
• Redis (optional) – Cache patient summaries.

⸻

APIs & Contracts
• External: REST/JSON following simplified FHIR R4 shapes.
• Internal: gRPC + Protocol Buffers.
• Buf used for lint & breaking-change detection.

⸻

Observability
• OpenTelemetry for traces, metrics, and logs.
• Prometheus + Grafana + Jaeger included in Docker Compose.
• Expose latency metrics & health/readiness probes.

⸻

Testing & Quality
• Table-driven unit tests and httptest for REST.
• HL7→FHIR golden tests for mapping correctness.
• Benchmarks (testing.B) for hot paths.
• golangci-lint, gofumpt, pre-commit hooks.
• Makefile for build/test/lint/dev stack.
• ADRs for key design choices.

⸻

CI/CD

GitHub Actions pipeline 1. Lint 2. Unit + integration tests 3. Buf breaking-check 4. Build & Trivy scan 5. Push Docker images to GHCR 6. Smoke test via Kind cluster

⸻

Deployment
• Docker Compose – one-command local stack:
Postgres, NATS, Grafana, Prometheus, Jaeger.
• Helm charts (minimal) for K8s preview.
• Google Kubernetes Engine (GKE) – manifests and Helm values for deploying the full stack to GKE with Workload Identity and Cloud SQL proxy (optional demo flavor).

⸻

Test-Driven Deployment
• Each service includes automated integration and contract tests run in CI before deployment.
• Kind-based smoke tests validate the Helm and GKE manifests.
• CI/CD enforces green tests and Buf compatibility before pushing images to GHCR.
• This ensures consistent, reliable deployments across environments.

⸻

Security & Supply Chain (lightweight)
• Distroless Docker images
• SBOM (Syft) + Trivy scan
• Basic rate limiting & PII redaction middleware

⸻

Repo Structure

/cmd/api-gateway
/cmd/patient-svc
/cmd/appointment-svc
/cmd/lab-adapter
/cmd/notify-svc
/pkg/fhir
/pkg/hl7
/pkg/events
/pkg/observability
/proto
/deploy/compose
/deploy/helm
/docs (architecture.md, ADR-0001.md, runbook.md)
/scripts (db, certs, demo bootstrap)
/.github/workflows (ci.yml)

⸻

Demo Scenarios 1. Patient + Appointment flow
→ REST call → gRPC chain traced in Jaeger → metrics visible in Grafana. 2. HL7 ingestion
→ post ORU message → FHIR Observation appears → notify-svc reacts. 3. CI run
→ Buf detects intentional proto breaking change → fails pipeline.
