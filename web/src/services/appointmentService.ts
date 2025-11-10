import { apiClient } from './api'
import type { Appointment } from '@/stores/types'

const APPOINTMENT_ENDPOINT = '/fhir/Appointment'

/**
 * List all appointments
 * TODO: Implement actual API call
 */
export const listAppointments = async (): Promise<Appointment[]> => {
  try {
    const response = await apiClient.get<{ entry: Array<{ resource: Appointment }> }>(
      APPOINTMENT_ENDPOINT,
    )
    // FHIR bundle response handling
    return response.data.entry?.map(entry => entry.resource) || []
  } catch (error) {
    console.error('Failed to list appointments:', error)
    // Placeholder: Return empty array for now
    return []
  }
}

/**
 * Get a single appointment by ID
 * TODO: Implement actual API call
 */
export const getAppointment = async (id: string): Promise<Appointment> => {
  try {
    const response = await apiClient.get<Appointment>(`${APPOINTMENT_ENDPOINT}/${id}`)
    return response.data
  } catch (error) {
    console.error(`Failed to fetch appointment ${id}:`, error)
    throw error
  }
}

/**
 * Create a new appointment
 * TODO: Implement actual API call
 */
export const createAppointment = async (
  appointment: Omit<Appointment, 'id' | 'meta'>,
): Promise<Appointment> => {
  try {
    const response = await apiClient.post<Appointment>(APPOINTMENT_ENDPOINT, appointment)
    return response.data
  } catch (error) {
    console.error('Failed to create appointment:', error)
    throw error
  }
}

/**
 * Update an existing appointment
 * TODO: Implement actual API call
 */
export const updateAppointment = async (
  id: string,
  updates: Partial<Appointment>,
): Promise<Appointment> => {
  try {
    const response = await apiClient.put<Appointment>(`${APPOINTMENT_ENDPOINT}/${id}`, updates)
    return response.data
  } catch (error) {
    console.error(`Failed to update appointment ${id}:`, error)
    throw error
  }
}

/**
 * Delete an appointment
 * TODO: Implement actual API call
 */
export const deleteAppointment = async (id: string): Promise<void> => {
  try {
    await apiClient.delete(`${APPOINTMENT_ENDPOINT}/${id}`)
  } catch (error) {
    console.error(`Failed to delete appointment ${id}:`, error)
    throw error
  }
}

/**
 * Get appointments for a patient
 * TODO: Implement actual API call
 */
export const getPatientAppointments = async (patientId: string): Promise<Appointment[]> => {
  try {
    const response = await apiClient.get<{ entry: Array<{ resource: Appointment }> }>(
      `${APPOINTMENT_ENDPOINT}?actor=${patientId}`,
    )
    return response.data.entry?.map(entry => entry.resource) || []
  } catch (error) {
    console.error(`Failed to fetch appointments for patient ${patientId}:`, error)
    return []
  }
}
