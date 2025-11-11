import { apiClient } from './api'
import type { Observation } from '@/stores/types'

const OBSERVATION_ENDPOINT = '/fhir/Observation'

export const listObservations = async (patientId: string): Promise<Observation[]> => {
  if (!patientId) return []
  const response = await apiClient.get<{ entry?: Array<{ resource: Observation }> }>(
    `${OBSERVATION_ENDPOINT}?patient=${patientId}`,
  )
  return response.data.entry?.map(item => item.resource) ?? []
}

export const getObservation = async (id: string): Promise<Observation> => {
  const response = await apiClient.get<Observation>(`${OBSERVATION_ENDPOINT}/${id}`)
  return response.data
}

export const createObservation = async (observation: Observation): Promise<Observation> => {
  const response = await apiClient.post<Observation>(OBSERVATION_ENDPOINT, observation)
  return response.data
}

export const updateObservationStatus = async (
  id: string,
  status: Observation['status'],
): Promise<Observation> => {
  const response = await apiClient.post<Observation>(`${OBSERVATION_ENDPOINT}/${id}`, {
    status,
  })
  return response.data
}
