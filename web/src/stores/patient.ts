import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Patient } from './types'
import * as patientService from '@/services/patientService'

export const usePatientStore = defineStore('patient', () => {
  // State
  const patients = ref<Patient[]>([])
  const currentPatient = ref<Patient | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const patientCount = computed(() => patients.value.length)
  const hasPatients = computed(() => patients.value.length > 0)
  const sortedPatients = computed(() => {
    return [...patients.value].sort((a, b) => {
      const nameA = a.name[0]?.family || ''
      const nameB = b.name[0]?.family || ''
      return nameA.localeCompare(nameB)
    })
  })

  // Actions
  const fetchPatients = async () => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      patients.value = await patientService.listPatients()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch patients'
      console.error('Error fetching patients:', err)
    } finally {
      loading.value = false
    }
  }

  const fetchPatientById = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      currentPatient.value = await patientService.getPatient(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch patient'
      console.error('Error fetching patient:', err)
    } finally {
      loading.value = false
    }
  }

  const createPatient = async (patient: Omit<Patient, 'id' | 'meta'>) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      const newPatient = await patientService.createPatient(patient)
      patients.value.push(newPatient)
      return newPatient
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create patient'
      console.error('Error creating patient:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const updatePatient = async (id: string, updates: Partial<Patient>) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      const updated = await patientService.updatePatient(id, updates)
      const index = patients.value.findIndex(p => p.id === id)
      if (index !== -1) {
        patients.value[index] = updated
      }
      if (currentPatient.value?.id === id) {
        currentPatient.value = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update patient'
      console.error('Error updating patient:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const deletePatient = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      await patientService.deletePatient(id)
      patients.value = patients.value.filter(p => p.id !== id)
      if (currentPatient.value?.id === id) {
        currentPatient.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete patient'
      console.error('Error deleting patient:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const clearCurrentPatient = () => {
    currentPatient.value = null
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // State
    patients,
    currentPatient,
    loading,
    error,
    // Computed
    patientCount,
    hasPatients,
    sortedPatients,
    // Actions
    fetchPatients,
    fetchPatientById,
    createPatient,
    updatePatient,
    deletePatient,
    clearCurrentPatient,
    clearError,
  }
})
