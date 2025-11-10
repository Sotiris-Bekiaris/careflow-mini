<template>
  <v-card
    class="metric-card"
    elevation="0"
    :loading="loading"
  >
    <div class="metric-card__header">
      <div>
        <p class="metric-card__subtitle">{{ subtitle }}</p>
        <h3 class="metric-card__title">{{ title }}</h3>
      </div>
      <div v-if="icon" class="metric-card__icon">
        <v-icon size="24" color="primary">{{ icon }}</v-icon>
      </div>
    </div>

    <div class="metric-card__value">
      <slot name="value">
        <span>{{ value }}</span>
      </slot>
    </div>

    <div class="metric-card__footer">
      <slot name="footer">
        <div v-if="delta" class="metric-card__delta" :class="`metric-card__delta--${delta.trend}`">
          <v-icon size="14" :color="deltaColor">{{ deltaIcon }}</v-icon>
          <span class="metric-card__delta-value">{{ delta.value }}</span>
          <span class="metric-card__delta-label">{{ delta.label }}</span>
        </div>
        <p v-else class="metric-card__baseline">{{ baseline }}</p>
      </slot>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface DeltaMeta {
  value: string
  label?: string
  trend?: 'up' | 'down' | 'flat'
}

const props = withDefaults(
  defineProps<{
    title: string
    value: string | number
    subtitle?: string
    baseline?: string
    icon?: string
    delta?: DeltaMeta | null
    loading?: boolean
  }>(),
  {
    subtitle: 'Snapshot',
    baseline: 'Updated moments ago',
    delta: null,
    loading: false,
  },
)

const deltaColor = computed(() => {
  if (!props.delta) return 'primary'
  if (props.delta.trend === 'down') return 'error'
  if (props.delta.trend === 'flat') return 'info'
  return 'success'
})

const deltaIcon = computed(() => {
  if (!props.delta) return ''
  const trend = props.delta.trend ?? 'up'
  if (trend === 'down') return 'mdi-arrow-bottom-right'
  if (trend === 'flat') return 'mdi-minus'
  return 'mdi-arrow-top-right'
})
</script>

<style scoped>
.metric-card {
  border-radius: var(--cf-radius-lg);
  border: 1px solid rgba(15, 23, 42, 0.06);
  padding: 1.5rem;
  background: var(--cf-surface);
  box-shadow: 0 15px 30px rgba(15, 23, 42, 0.06);
}

.metric-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.metric-card__icon {
  background: linear-gradient(135deg, rgba(10, 132, 255, 0.12), rgba(99, 102, 241, 0.08));
  border-radius: 16px;
  padding: 0.65rem;
}

.metric-card__subtitle {
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--cf-text-muted);
  margin: 0;
}

.metric-card__title {
  font-size: 1.4rem;
  font-weight: 600;
  margin: 0.15rem 0 0;
}

.metric-card__value {
  font-size: 2.6rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: 1.25rem;
}

.metric-card__footer {
  display: flex;
  align-items: center;
  min-height: 1.5rem;
}

.metric-card__delta {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--cf-text);
  padding: 0.15rem 0.35rem;
  border-radius: 999px;
  background: rgba(10, 132, 255, 0.08);
}

.metric-card__delta--down {
  background: rgba(255, 69, 58, 0.08);
}

.metric-card__delta--flat {
  background: rgba(100, 210, 255, 0.12);
}

.metric-card__delta-value {
  font-variant-numeric: tabular-nums;
}

.metric-card__delta-label {
  color: var(--cf-text-muted);
}

.metric-card__baseline {
  font-size: 0.9rem;
  color: var(--cf-text-muted);
  margin: 0;
}
</style>
