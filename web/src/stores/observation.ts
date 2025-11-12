import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Observation } from './types'
import * as observationService from '@/services/observationService'

export const useObservationStore = defineStore('observation', () => {
  // State
  const observations = ref<Observation[]>([])
  const currentObservation = ref<Observation | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const patientObservations = ref<Map<string, Observation[]>>(new Map())

  // Computed
  const observationCount = computed(() => observations.value.length)
  const hasObservations = computed(() => observations.value.length > 0)
  const finalObservations = computed(() => {
    return observations.value.filter(o => o.status === 'final')
  })

  // Actions
  const fetchPatientObservations = async (patientId: string) => {
    loading.value = true
    error.value = null
    try {
      const results = await observationService.listPatientObservations(patientId)
      patientObservations.value.set(patientId, results)
      observations.value = results
      return results
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch observations'
      console.error('Error fetching observations:', err)
      return []
    } finally {
      loading.value = false
    }
  }

  const fetchObservationById = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      currentObservation.value = await observationService.getObservation(id)
      return currentObservation.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch observation'
      console.error('Error fetching observation:', err)
      return null
    } finally {
      loading.value = false
    }
  }

  const createObservation = async (observation: Partial<Observation>) => {
    loading.value = true
    error.value = null
    try {
      const newObservation = await observationService.createObservation(observation)
      if (newObservation) {
        observations.value.push(newObservation)
      }
      return newObservation
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create observation'
      console.error('Error creating observation:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const generateLabData = async (patientId: string) => {
    loading.value = true
    error.value = null
    try {
      const generated = await observationService.generateLabData(patientId)
      // Update the observations for this patient
      patientObservations.value.set(patientId, generated)
      observations.value = generated
      return generated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to generate lab data'
      console.error('Error generating lab data:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const clearError = () => {
    error.value = null
  }

  const getPatientObservationsFromCache = (patientId: string): Observation[] => {
    return patientObservations.value.get(patientId) || []
  }

  return {
    // State
    observations,
    currentObservation,
    loading,
    error,
    patientObservations,

    // Computed
    observationCount,
    hasObservations,
    finalObservations,

    // Actions
    fetchPatientObservations,
    fetchObservationById,
    createObservation,
    generateLabData,
    clearError,
    getPatientObservationsFromCache,
  }
})
