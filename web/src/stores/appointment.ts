import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Appointment } from './types'
import * as appointmentService from '@/services/appointmentService'

export const useAppointmentStore = defineStore('appointment', () => {
  // State
  const appointments = ref<Appointment[]>([])
  const currentAppointment = ref<Appointment | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filterStatus = ref<string>('all')

  // Computed
  const appointmentCount = computed(() => appointments.value.length)
  const hasAppointments = computed(() => appointments.value.length > 0)
  const filteredAppointments = computed(() => {
    if (filterStatus.value === 'all') {
      return appointments.value
    }
    return appointments.value.filter(a => a.status === filterStatus.value)
  })
  const upcomingAppointments = computed(() => {
    const now = new Date()
    return appointments.value
      .filter(a => new Date(a.start) > now && a.status !== 'cancelled')
      .sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime())
  })

  // Actions
  const fetchAppointments = async () => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      appointments.value = await appointmentService.listAppointments()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch appointments'
      console.error('Error fetching appointments:', err)
    } finally {
      loading.value = false
    }
  }

  const fetchAppointmentById = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      currentAppointment.value = await appointmentService.getAppointment(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch appointment'
      console.error('Error fetching appointment:', err)
    } finally {
      loading.value = false
    }
  }

  const createAppointment = async (appointment: Omit<Appointment, 'id' | 'meta'>) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      const newAppointment = await appointmentService.createAppointment(appointment)
      appointments.value.push(newAppointment)
      return newAppointment
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create appointment'
      console.error('Error creating appointment:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateAppointment = async (id: string, updates: Partial<Appointment>) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      const updated = await appointmentService.updateAppointment(id, updates)
      const index = appointments.value.findIndex(a => a.id === id)
      if (index !== -1) {
        appointments.value[index] = updated
      }
      if (currentAppointment.value?.id === id) {
        currentAppointment.value = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update appointment'
      console.error('Error updating appointment:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteAppointment = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      // TODO: Implement actual API call
      await appointmentService.deleteAppointment(id)
      appointments.value = appointments.value.filter(a => a.id !== id)
      if (currentAppointment.value?.id === id) {
        currentAppointment.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete appointment'
      console.error('Error deleting appointment:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const setFilterStatus = (status: string) => {
    filterStatus.value = status
  }

  const clearCurrentAppointment = () => {
    currentAppointment.value = null
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // State
    appointments,
    currentAppointment,
    loading,
    error,
    filterStatus,
    // Computed
    appointmentCount,
    hasAppointments,
    filteredAppointments,
    upcomingAppointments,
    // Actions
    fetchAppointments,
    fetchAppointmentById,
    createAppointment,
    updateAppointment,
    deleteAppointment,
    setFilterStatus,
    clearCurrentAppointment,
    clearError,
  }
})
