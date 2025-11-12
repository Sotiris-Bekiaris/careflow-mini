import { apiClient } from './api'
import type { Patient } from '@/stores/types'

const PATIENT_ENDPOINT = '/fhir/Patient'

/**
 * List all patients
 * TODO: Implement actual API call
 */
export const listPatients = async (): Promise<Patient[]> => {
  try {
    const response = await apiClient.get<{ entry: Patient[] }>(PATIENT_ENDPOINT)
    // FHIR bundle response handling - patients are returned directly in entry array
    return response.data.entry || []
  } catch (error) {
    console.error('Failed to list patients:', error)
    // Placeholder: Return empty array for now
    return []
  }
}

/**
 * Get a single patient by ID
 * TODO: Implement actual API call
 */
export const getPatient = async (id: string): Promise<Patient> => {
  try {
    const response = await apiClient.get<Patient>(`${PATIENT_ENDPOINT}/${id}`)
    return response.data
  } catch (error) {
    console.error(`Failed to fetch patient ${id}:`, error)
    throw error
  }
}

/**
 * Create a new patient
 * TODO: Implement actual API call
 */
export const createPatient = async (patient: Omit<Patient, 'id' | 'meta'>): Promise<Patient> => {
  try {
    const response = await apiClient.post<Patient>(PATIENT_ENDPOINT, patient)
    return response.data
  } catch (error) {
    console.error('Failed to create patient:', error)
    throw error
  }
}

/**
 * Update an existing patient
 * TODO: Implement actual API call
 */
export const updatePatient = async (id: string, updates: Partial<Patient>): Promise<Patient> => {
  try {
    const response = await apiClient.put<Patient>(`${PATIENT_ENDPOINT}/${id}`, updates)
    return response.data
  } catch (error) {
    console.error(`Failed to update patient ${id}:`, error)
    throw error
  }
}

/**
 * Delete a patient
 * TODO: Implement actual API call
 */
export const deletePatient = async (id: string): Promise<void> => {
  try {
    await apiClient.delete(`${PATIENT_ENDPOINT}/${id}`)
  } catch (error) {
    console.error(`Failed to delete patient ${id}:`, error)
    throw error
  }
}

/**
 * Search patients by criteria
 * TODO: Implement actual API call
 */
export const searchPatients = async (query: string): Promise<Patient[]> => {
  try {
    const response = await apiClient.get<{ entry: Patient[] }>(
      `${PATIENT_ENDPOINT}?name=${query}`,
    )
    return response.data.entry || []
  } catch (error) {
    console.error('Failed to search patients:', error)
    return []
  }
}
