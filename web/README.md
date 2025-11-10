# CareFlow Web Experience

This Vite/Vue 3 UI is the visual companion to the Go microservices in this repo. It exposes a minimal, Apple-inspired light theme that showcases live metrics from:

- `api-gateway` (REST/FHIR facade)
- `patient-svc` (FHIR Patient CRUD)
- `appointment-svc` (scheduler + events)

The dashboard hits `/health`, `/ready`, and the `/fhir/*` endpoints exposed by the gateway so the UI always reflects the current behaviour of the stack.

## Quick start

```bash
cd web
npm install
npm run dev
```

Key scripts:

- `npm run dev` – start Vite with HMR
- `npm run build` – type-check via `vue-tsc` and emit the production bundle
- `npm run preview` – preview the production build locally
- `npm run lint` – Vue + TypeScript linting

## Config

Use `VITE_API_BASE_URL` to point the UI at a running gateway (defaults to `http://localhost:8080`). Example `.env.local`:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

## Design system

- Vuetify 3 custom theme with SF Pro/Inter stack, glassmorphic surfaces, and soft elevation.
- Purpose-built UI components in `src/components/ui` (MetricCard, SystemStatusCard, SectionHeader, EmptyState) keep the layout DRY.
- Shared formatters for FHIR resources live in `src/utils/formatters.ts`.

## Data flow

```
Vue components → Pinia stores → services/*.ts → api-gateway → gRPC services
```

The `systemService` probes `/health` and `/ready` to infer the health of the Go services, while the patient and appointment stores call the `/fhir/*` routes the gateway already exposes.
