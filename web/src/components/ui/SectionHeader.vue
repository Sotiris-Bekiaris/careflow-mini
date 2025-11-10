<template>
  <div class="section-header">
    <div>
      <p v-if="eyebrow" class="section-header__eyebrow">{{ eyebrow }}</p>
      <h2>{{ title }}</h2>
      <p v-if="description" class="section-header__description">{{ description }}</p>
    </div>
    <div class="section-header__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
}>()
</script>

<style scoped>
.section-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.section-header__eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.75rem;
  color: var(--cf-text-muted);
  margin: 0;
}

.section-header h2 {
  margin: 0.25rem 0;
  font-size: 1.8rem;
}

.section-header__description {
  margin: 0;
  color: var(--cf-text-muted);
  max-width: 40rem;
}

.section-header__actions {
  display: flex;
  gap: 0.75rem;
}
</style>
