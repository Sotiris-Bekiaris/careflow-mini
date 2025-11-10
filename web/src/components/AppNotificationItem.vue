<template>
  <v-card class="ma-2" :color="`${notification.type}15`">
    <v-card-text class="d-flex justify-space-between align-center">
      <div class="d-flex align-center gap-2">
        <v-icon :color="notification.type">
          {{ notificationIcon }}
        </v-icon>
        <span class="text-sm">{{ notification.message }}</span>
      </div>
      <v-btn icon size="x-small" variant="text" @click="$emit('close')">
        <v-icon size="small">mdi-close</v-icon>
      </v-btn>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { UINotification } from '@/stores/types'

interface Props {
  notification: UINotification
}

defineProps<Props>()
defineEmits<{
  close: []
}>()

const notificationIcon = computed(() => {
  const icons: Record<string, string> = {
    success: 'mdi-check-circle',
    error: 'mdi-alert-circle',
    warning: 'mdi-alert',
    info: 'mdi-information',
  }
  return icons[props.notification.type] || 'mdi-information'
})

const props = defineProps<Props>()
</script>
