<template>
  <v-card class="system-status" elevation="0" :loading="loading">
    <header class="system-status__header">
      <div>
        <p class="system-status__eyebrow">Runtime health</p>
        <h3>CareFlow Microservices</h3>
      </div>
      <v-chip size="small" class="system-status__chip" color="primary" variant="tonal">
        {{ healthyCount }} / {{ services.length }} healthy
      </v-chip>
    </header>

    <v-divider class="my-4"></v-divider>

    <div class="system-status__grid">
      <div
        v-for="service in services"
        :key="service.id"
        class="system-status__row"
      >
        <div class="system-status__identity">
          <div class="system-status__avatar">
            <v-icon size="18" color="primary">{{ service.icon || 'mdi-server' }}</v-icon>
          </div>
          <div>
            <p class="system-status__name">{{ service.name }}</p>
            <p class="system-status__description">{{ service.description }}</p>
          </div>
        </div>

        <div class="system-status__meta">
          <span v-if="service.latencyMs" class="system-status__latency">
            {{ service.latencyMs }} ms
          </span>
          <v-chip
            size="small"
            :color="statusColor(service.status)"
            variant="flat"
            class="system-status__state"
          >
            {{ formatStatus(service.status) }}
          </v-chip>
        </div>
      </div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ServiceStatus, ServiceHealthStatus } from '@/stores/types'

const props = withDefaults(
  defineProps<{
    services: ServiceStatus[]
    loading?: boolean
  }>(),
  {
    services: () => [],
    loading: false,
  },
)

const healthyCount = computed(() => props.services.filter(s => s.status === 'healthy').length)

const statusColor = (status: ServiceHealthStatus) => {
  switch (status) {
    case 'healthy':
      return 'success'
    case 'degraded':
      return 'warning'
    case 'offline':
      return 'error'
    default:
      return 'info'
  }
}

const formatStatus = (status: ServiceHealthStatus) => {
  return status.charAt(0).toUpperCase() + status.slice(1)
}
</script>

<style scoped>
.system-status {
  border-radius: var(--cf-radius-lg);
  border: 1px solid rgba(15, 23, 42, 0.06);
  padding: 1.75rem;
  background: var(--cf-surface);
  box-shadow: 0 15px 40px rgba(15, 23, 42, 0.08);
}

.system-status__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.system-status__eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.75rem;
  color: var(--cf-text-muted);
  margin: 0;
}

.system-status__header h3 {
  margin: 0.2rem 0 0;
  font-size: 1.25rem;
}

.system-status__chip {
  font-weight: 600;
}

.system-status__grid {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.system-status__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.25rem 0;
}

.system-status__identity {
  display: flex;
  gap: 0.85rem;
  align-items: center;
}

.system-status__avatar {
  width: 40px;
  height: 40px;
  border-radius: 14px;
  background: rgba(10, 132, 255, 0.1);
  display: grid;
  place-items: center;
}

.system-status__name {
  margin: 0;
  font-weight: 600;
}

.system-status__description {
  margin: 0;
  color: var(--cf-text-muted);
  font-size: 0.9rem;
}

.system-status__meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.system-status__latency {
  font-variant-numeric: tabular-nums;
  color: var(--cf-text-muted);
}

.system-status__state {
  text-transform: capitalize;
  font-weight: 600;
}

@media (max-width: 720px) {
  .system-status__row {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
  .system-status__meta {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
