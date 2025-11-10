import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ServiceStatus } from './types'
import { fetchHealthSnapshot } from '@/services/systemService'

export const useSystemStore = defineStore('system', () => {
  const services = ref<ServiceStatus[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdated = ref<string | null>(null)

  const healthyCount = computed(() => services.value.filter(s => s.status === 'healthy').length)

  const refresh = async () => {
    loading.value = true
    error.value = null
    try {
      const { statuses, timestamp } = await fetchHealthSnapshot()
      services.value = statuses
      lastUpdated.value = timestamp
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unable to load service health'
      console.error('System health fetch failed:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    services,
    loading,
    error,
    lastUpdated,
    healthyCount,
    refresh,
  }
})
