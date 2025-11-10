<template>
  <div class="empty-state">
    <div class="empty-state__icon" v-if="icon">
      <v-icon size="40" color="primary">{{ icon }}</v-icon>
    </div>
    <h3>{{ title }}</h3>
    <p>{{ description }}</p>

    <div class="empty-state__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    icon?: string
    title: string
    description: string
  }>(),
  {
    icon: 'mdi-information-outline',
  },
)
</script>

<style scoped>
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--cf-text-muted);
}

.empty-state__icon {
  display: inline-flex;
  padding: 1rem;
  border-radius: 20px;
  background: rgba(10, 132, 255, 0.08);
  margin-bottom: 1rem;
}

.empty-state h3 {
  margin: 0.5rem 0;
  font-size: 1.35rem;
  color: var(--cf-text);
}

.empty-state p {
  margin: 0 auto;
  max-width: 28rem;
}

.empty-state__actions {
  margin-top: 1.5rem;
  display: inline-flex;
  gap: 0.75rem;
}
</style>
