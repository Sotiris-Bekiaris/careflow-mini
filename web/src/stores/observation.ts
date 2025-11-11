import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Observation } from './types'
import * as observationService from '@/services/observationService'

export const useObservationStore = defineStore('observation', () => {
  const observations = ref<Observation[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentPatientId = ref<string | null>(null)

  const fetchObservations = async (patientId: string) => {
    if (!patientId) return
    loading.value = true
    error.value = null
    currentPatientId.value = patientId
    try {
      const data = await observationService.listObservations(patientId)
      observations.value = data.sort((a, b) => {
        const dateA = new Date(a.effectiveDateTime || '').getTime()
        const dateB = new Date(b.effectiveDateTime || '').getTime()
        return dateB - dateA
      })
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load observations'
      console.error('Observation fetch failed:', err)
    } finally {
      loading.value = false
    }
  }

  const setObservationStatus = async (id: string, status: Observation['status']) => {
    const updated = await observationService.updateObservationStatus(id, status)
    const index = observations.value.findIndex(obs => obs.id === id)
    if (index !== -1) {
      observations.value[index] = updated
      observations.value = [...observations.value].sort((a, b) => {
        const dateA = new Date(a.effectiveDateTime || '').getTime()
        const dateB = new Date(b.effectiveDateTime || '').getTime()
        return dateB - dateA
      })
    }
    return updated
  }

  const latestObservation = computed(() => observations.value[0] || null)
  const hasObservations = computed(() => observations.value.length > 0)

  return {
    observations,
    loading,
    error,
    currentPatientId,
    latestObservation,
    hasObservations,
    fetchObservations,
    setObservationStatus,
  }
})
