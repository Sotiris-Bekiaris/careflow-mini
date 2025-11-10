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
    endpoint: '/health',
    probe: '/health',
  },
  {
    id: 'patient-svc',
    name: 'Patient Service',
    description: 'gRPC CRUD for FHIR Patient',
    kind: 'gRPC',
    icon: 'mdi-account-heart',
    endpoint: '/ready',
    probe: '/ready',
  },
  {
    id: 'appointment-svc',
    name: 'Appointment Service',
    description: 'gRPC scheduler + NATS events',
    kind: 'gRPC',
    icon: 'mdi-calendar-clock',
    fallbackStatus: 'unknown',
  },
  {
    id: 'lab-adapter',
    name: 'Lab Adapter',
    description: 'HL7 → FHIR worker pipeline',
    kind: 'Worker',
    icon: 'mdi-flask',
    fallbackStatus: 'unknown',
  },
  {
    id: 'notify-svc',
    name: 'Notify Service',
    description: 'Event consumer + alerts',
    kind: 'Worker',
    icon: 'mdi-bell-ring',
    fallbackStatus: 'unknown',
  },
]

type ProbeResult = {
  ok: boolean
  latencyMs: number
}

const getNow = () => (typeof performance !== 'undefined' ? performance.now() : Date.now())

const runProbe = async (endpoint: string): Promise<ProbeResult> => {
  const started = getNow()
  try {
    await apiClient.get(endpoint)
    return {
      ok: true,
      latencyMs: Math.max(1, Math.round(getNow() - started)),
    }
  } catch (error) {
    return {
      ok: false,
      latencyMs: Math.max(1, Math.round(getNow() - started)),
    }
  }
}

export const fetchHealthSnapshot = async () => {
  const timestamp = new Date().toISOString()
  const probeResults: Record<string, ProbeResult> = {}

  await Promise.all(
    SERVICE_BLUEPRINTS.filter(service => service.probe).map(async service => {
      const probe = await runProbe(service.probe!)
      probeResults[service.id] = probe
    }),
  )

  const statuses: ServiceStatus[] = SERVICE_BLUEPRINTS.map(service => {
    const probe = probeResults[service.id]
    let status: ServiceHealthStatus = service.fallbackStatus ?? 'offline'
    if (probe) {
      status = probe.ok ? 'healthy' : 'degraded'
    }
    const { probe: _probe, fallbackStatus, ...descriptor } = service
    return {
      ...descriptor,
      status,
      latencyMs: probe?.latencyMs,
      lastChecked: timestamp,
    }
  })

  return { statuses, timestamp }
}
