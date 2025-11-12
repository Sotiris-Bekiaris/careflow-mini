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
          <v-select
            v-model="formData.patientId"
            :items="patientOptions"
            item-title="display"
            item-value="id"
            label="Patient"
            :rules="[requiredRule]"
            :loading="patientStore.loading"
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
import { usePatientStore } from '@/stores/patient'
import { useUIStore } from '@/stores/ui'
import SectionHeader from '@/components/ui/SectionHeader.vue'

const route = useRoute()
const router = useRouter()
const appointmentStore = useAppointmentStore()
const patientStore = usePatientStore()
const uiStore = useUIStore()
const form = ref()

const isEditing = computed(() => !!route.params.id)

const formData = reactive({
  description: '',
  status: 'booked',
  patientId: '',
  start: '',
  end: '',
  notes: '',
})

const patientOptions = computed(() =>
  patientStore.patients.map(patient => ({
    id: patient.id,
    display: `${patient.name?.[0]?.given?.join(' ')} ${patient.name?.[0]?.family || patient.id}`.trim(),
  })),
)

const statusOptions = ['proposed', 'pending', 'booked', 'arrived', 'fulfilled', 'cancelled', 'noshow']
const requiredRule = (value: string) => (!!value && value.trim().length > 0) || 'Required'

// Convert RFC3339 format to datetime-local format for form input
const formatToDateTimeLocal = (rfc3339: string): string => {
  if (!rfc3339) return ''
  // Input format: "2025-10-31T01:19:00Z" or "2025-10-31T01:19:00+00:00"
  // Output format: "2025-10-31T01:19"
  try {
    const date = new Date(rfc3339)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day}T${hours}:${minutes}`
  } catch {
    return ''
  }
}

onMounted(async () => {
  // Always fetch patients for the dropdown
  await patientStore.fetchPatients()

  if (!isEditing.value) return
  await appointmentStore.fetchAppointmentById(route.params.id as string)
  const appointment = appointmentStore.currentAppointment
  if (!appointment) return
  formData.description = appointment.description ?? ''
  formData.status = appointment.status ?? 'booked'
  // Extract patient ID from reference like "Patient/123" -> "123"
  const reference = appointment.participant?.[0]?.actor?.reference ?? ''
  formData.patientId = reference.includes('/') ? reference.split('/').pop() || '' : reference
  formData.start = formatToDateTimeLocal(appointment.start ?? '')
  formData.end = formatToDateTimeLocal(appointment.end ?? '')
  formData.notes = ''
})

// Convert datetime-local format to RFC3339 format
const formatToRFC3339 = (dateTimeLocal: string): string => {
  if (!dateTimeLocal) return ''
  // Input format: "2025-10-31T01:19"
  // Output format: "2025-10-31T01:19:00Z"
  const [datePart, timePart] = dateTimeLocal.split('T')
  if (!datePart || !timePart) return ''

  // Add seconds if not present, and add timezone
  const timeWithSeconds = timePart.includes(':') ?
    (timePart.split(':').length === 2 ? `${timePart}:00` : timePart) :
    `${timePart}:00:00`

  return `${datePart}T${timeWithSeconds}Z`
}

const submitForm = async () => {
  const result = await form.value?.validate()
  if (!result?.valid) return

  // Format patient reference with "Patient/" prefix
  const patientReference = `Patient/${formData.patientId}`
  const selectedPatient = patientStore.patients.find(p => p.id === formData.patientId)
  const patientDisplay = selectedPatient
    ? `${selectedPatient.name?.[0]?.given?.join(' ')} ${selectedPatient.name?.[0]?.family || formData.patientId}`.trim()
    : formData.patientId

  const payload = {
    resourceType: 'Appointment' as const,
    description: formData.description,
    status: formData.status as any,
    start: formatToRFC3339(formData.start),
    end: formatToRFC3339(formData.end),
    participant: [
      {
        actor: {
          reference: patientReference,
          display: patientDisplay,
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
