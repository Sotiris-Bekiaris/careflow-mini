<template>
  <v-card class="notification" elevation="0">
    <v-card-text class="notification__content">
      <div class="notification__icon" :class="`notification__icon--${notification.type}`">
        <v-icon size="18">{{ notificationIcon }}</v-icon>
      </div>
      <span>{{ notification.message }}</span>
      <v-btn icon size="x-small" variant="text" @click="$emit('close')">
        <v-icon size="16">mdi-close</v-icon>
      </v-btn>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import type { UINotification } from '@/stores/types'

const props = defineProps<{
  notification: UINotification
}>()

defineEmits<{ close: [] }>()

const notification = toRef(props, 'notification')

const notificationIcon = computed(() => {
  const icons: Record<string, string> = {
    success: 'mdi-check-circle',
    error: 'mdi-alert-circle',
    warning: 'mdi-alert',
    info: 'mdi-information',
  }
  return icons[notification.value.type] || 'mdi-information'
})
</script>

<style scoped>
.notification {
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  margin: 0.75rem 1rem;
}

.notification__content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.notification__icon {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: rgba(10, 132, 255, 0.08);
}

.notification__icon--success {
  background: rgba(52, 199, 89, 0.15);
  color: #34c759;
}

.notification__icon--error {
  background: rgba(255, 69, 58, 0.15);
  color: #ff453a;
}

.notification__icon--warning {
  background: rgba(255, 159, 10, 0.15);
  color: #ff9f0a;
}

.notification__icon--info {
  background: rgba(100, 210, 255, 0.15);
  color: #64d2ff;
}
</style>
