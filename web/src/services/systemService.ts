import { apiClient } from './api'
import type { ServiceStatus, ServiceHealthStatus } from '@/stores/types'

type ServiceBlueprint = Omit<ServiceStatus, 'status'> & {
  probe?: string
  fallbackStatus?: ServiceHealthStatus
}

const SERVICE_BLUEPRINTS: ServiceBlueprint[] = [
  {
    id: 'api-gateway',
    name: 'API Gateway',
    description: 'REST facade + tracing middleware',
    kind: 'REST',
    icon: 'mdi-access-point-network',
  },
  {
    id: 'patient-svc',
    name: 'Patient Service',
    description: 'gRPC CRUD for FHIR Patient',
    kind: 'gRPC',
    icon: 'mdi-account-heart',
  },
  {
    id: 'appointment-svc',
    name: 'Appointment Service',
    description: 'gRPC scheduler + NATS events',
    kind: 'gRPC',
    icon: 'mdi-calendar-clock',
  },
  {
    id: 'observation-svc',
    name: 'Observation Service',
    description: 'Lab results + FHIR observations',
    kind: 'gRPC',
    icon: 'mdi-test-tube',
  },
  {
    id: 'lab-adapter',
    name: 'Lab Adapter',
    description: 'HL7 → FHIR worker pipeline',
    kind: 'Worker',
    icon: 'mdi-flask',
  },
  {
    id: 'notify-svc',
    name: 'Notify Service',
    description: 'Event consumer + alerts',
    kind: 'Worker',
    icon: 'mdi-bell-ring',
  },
]

type HealthServicesResponse = {
  timestamp: string
  summary: {
    total: number
    healthy: number
    degraded: number
    offline: number
  }
  services: Array<{
    id: string
    name: string
    status: string
    latencyMs: number
    message?: string
  }>
}

export const fetchHealthSnapshot = async () => {
  try {
    const response = await apiClient.get<HealthServicesResponse>('/health/services')
    const timestamp = response.data.timestamp

    // Create a map of backend service data
    const backendServices = new Map(response.data.services.map(s => [s.id, s]))

    // Merge backend data with frontend blueprints
    const statuses: ServiceStatus[] = SERVICE_BLUEPRINTS.map(blueprint => {
      const backendService = backendServices.get(blueprint.id)

      if (backendService) {
        // Use backend data
        return {
          id: blueprint.id,
          name: blueprint.name,
          description: blueprint.description,
          kind: blueprint.kind,
          icon: blueprint.icon,
          status: backendService.status as ServiceHealthStatus,
          latencyMs: backendService.latencyMs,
          lastChecked: timestamp,
        }
      } else {
        // Service not found in backend response - mark as offline
        return {
          id: blueprint.id,
          name: blueprint.name,
          description: blueprint.description,
          kind: blueprint.kind,
          icon: blueprint.icon,
          status: 'offline',
          lastChecked: timestamp,
        }
      }
    })

    return { statuses, timestamp }
  } catch (error) {
    console.error('Failed to fetch health snapshot:', error)
    // Return all services as offline on error
    const timestamp = new Date().toISOString()
    const statuses: ServiceStatus[] = SERVICE_BLUEPRINTS.map(blueprint => ({
      id: blueprint.id,
      name: blueprint.name,
      description: blueprint.description,
      kind: blueprint.kind,
      icon: blueprint.icon,
      status: 'offline' as ServiceHealthStatus,
      lastChecked: timestamp,
    }))
    return { statuses, timestamp }
  }
}
