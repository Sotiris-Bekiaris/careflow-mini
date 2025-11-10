<template>
  <div class="page">
    <SectionHeader
      :title="isEditing ? 'Edit appointment' : 'Create appointment'"
      description="Send an appointment payload through the Go scheduler"
      eyebrow="Scheduling"
    >
      <template #actions>
        <router-link to="/appointments">
          <v-btn variant="text">Back to list</v-btn>
        </router-link>
      </template>
    </SectionHeader>

    <v-card class="panel">
      <v-form ref="form" @submit.prevent="submitForm">
        <div class="grid">
          <v-text-field
            v-model="formData.description"
            label="Description"
            :rules="[requiredRule]"
            variant="solo"
            density="comfortable"
          />
          <v-select
            v-model="formData.status"
            :items="statusOptions"
            label="Status"
            variant="solo"
            density="comfortable"
          />
          <v-text-field
            v-model="formData.patientReference"
            label="Patient reference (Patient/{id})"
            variant="solo"
            density="comfortable"
          />
          <v-text-field
            v-model="formData.start"
            label="Start"
            type="datetime-local"
            variant="solo"
            density="comfortable"
          />
          <v-text-field
            v-model="formData.end"
            label="End"
            type="datetime-local"
            variant="solo"
            density="comfortable"
          />
        </div>
        <v-textarea
          v-model="formData.notes"
          label="Notes"
          rows="4"
          variant="solo"
          class="mt-4"
        />

        <div class="actions">
          <v-btn type="submit" color="primary" size="large" :loading="appointmentStore.loading">
            <v-icon start>mdi-check</v-icon>
            {{ isEditing ? 'Update appointment' : 'Create appointment' }}
          </v-btn>
          <router-link to="/appointments">
            <v-btn size="large" variant="text">Cancel</v-btn>
          </router-link>
        </div>
      </v-form>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppointmentStore } from '@/stores/appointment'
import { useUIStore } from '@/stores/ui'
import SectionHeader from '@/components/ui/SectionHeader.vue'

const route = useRoute()
const router = useRouter()
const appointmentStore = useAppointmentStore()
const uiStore = useUIStore()
const form = ref()

const isEditing = computed(() => !!route.params.id)

const formData = reactive({
  description: '',
  status: 'booked',
  patientReference: '',
  start: '',
  end: '',
  notes: '',
})

const statusOptions = ['proposed', 'pending', 'booked', 'arrived', 'fulfilled', 'cancelled', 'noshow']
const requiredRule = (value: string) => (!!value && value.trim().length > 0) || 'Required'

onMounted(async () => {
  if (!isEditing.value) return
  await appointmentStore.fetchAppointmentById(route.params.id as string)
  const appointment = appointmentStore.currentAppointment
  if (!appointment) return
  formData.description = appointment.description ?? ''
  formData.status = appointment.status ?? 'booked'
  formData.patientReference = appointment.participant?.[0]?.actor?.reference ?? ''
  formData.start = appointment.start ?? ''
  formData.end = appointment.end ?? ''
  formData.notes = ''
})

const submitForm = async () => {
  const result = await form.value?.validate()
  if (!result?.valid) return

  const payload = {
    resourceType: 'Appointment' as const,
    description: formData.description,
    status: formData.status as any,
    start: formData.start,
    end: formData.end,
    participant: [
      {
        actor: {
          reference: formData.patientReference,
          display: formData.patientReference,
        },
        status: 'accepted',
      },
    ],
  }

  try {
    if (isEditing.value) {
      await appointmentStore.updateAppointment(route.params.id as string, payload as any)
    } else {
      await appointmentStore.createAppointment(payload as any)
    }
    uiStore.addNotification(`Appointment ${isEditing.value ? 'updated' : 'created'}`, 'success')
    router.push('/appointments')
  } catch (error) {
    uiStore.addNotification('Unable to save appointment', 'error')
    console.error('Failed to save appointment:', error)
  }
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel {
  border-radius: var(--cf-radius-lg);
  box-shadow: var(--cf-shadow-soft);
  padding: 2rem;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.actions {
  margin-top: 2rem;
  display: flex;
  gap: 1rem;
}
</style>
