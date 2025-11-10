<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="flex justify-between items-center mb-6">
      <router-link to="/appointments">
        <v-btn icon variant="text" color="primary">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
      </router-link>

      <h1 class="text-3xl font-bold text-gray-800 flex-1">
        {{ isEditing ? 'Edit Appointment' : 'Create Appointment' }}
      </h1>
    </div>

    <!-- Form Card -->
    <v-card>
      <v-card-text class="pt-6">
        <v-form ref="form" @submit.prevent="submitForm">
          <v-row>
            <!-- Description -->
            <v-col cols="12">
              <v-text-field
                v-model="formData.description"
                label="Appointment Description"
                variant="outlined"
                density="compact"
                rules="required"
                required
              />
            </v-col>

            <!-- Status -->
            <v-col cols="12" sm="6">
              <v-select
                v-model="formData.status"
                :items="[
                  'proposed',
                  'pending',
                  'booked',
                  'arrived',
                  'fulfilled',
                  'cancelled',
                  'noshow',
                ]"
                label="Status"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Participant -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.participantDisplay"
                label="Participant / Patient"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Start Time -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.start"
                label="Start Time"
                type="datetime-local"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- End Time -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.end"
                label="End Time"
                type="datetime-local"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Notes -->
            <v-col cols="12">
              <v-textarea
                v-model="formData.notes"
                label="Notes"
                variant="outlined"
                density="compact"
                rows="4"
              />
            </v-col>
          </v-row>

          <!-- Form Actions -->
          <v-row class="mt-6">
            <v-col cols="12" class="d-flex gap-3">
              <v-btn type="submit" color="primary" size="large">
                <v-icon start>mdi-check</v-icon>
                {{ isEditing ? 'Update Appointment' : 'Create Appointment' }}
              </v-btn>

              <v-btn
                variant="outlined"
                size="large"
                @click="$router.push('/appointments')"
              >
                Cancel
              </v-btn>
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppointmentStore } from '@/stores/appointment'

const route = useRoute()
const router = useRouter()
const appointmentStore = useAppointmentStore()
const form = ref()

const isEditing = computed(() => !!route.params.id)

const formData = reactive({
  description: '',
  status: 'booked',
  participantDisplay: '',
  start: '',
  end: '',
  notes: '',
})

onMounted(async () => {
  if (isEditing.value) {
    const id = route.params.id as string
    await appointmentStore.fetchAppointmentById(id)

    if (appointmentStore.currentAppointment) {
      const appointment = appointmentStore.currentAppointment
      formData.description = appointment.description || ''
      formData.status = appointment.status || 'booked'
      formData.participantDisplay = appointment.participant?.[0]?.actor?.display || ''
      formData.start = appointment.start || ''
      formData.end = appointment.end || ''
    }
  }
})

const submitForm = async () => {
  if (form.value && await form.value.validate()) {
    const appointmentData = {
      resourceType: 'Appointment' as const,
      description: formData.description,
      status: formData.status as any,
      start: formData.start,
      end: formData.end,
      participant: [
        {
          actor: {
            reference: formData.participantDisplay,
            display: formData.participantDisplay,
          },
          status: 'accepted',
        },
      ],
    }

    try {
      if (isEditing.value) {
        await appointmentStore.updateAppointment(route.params.id as string, appointmentData as any)
      } else {
        await appointmentStore.createAppointment(appointmentData as any)
      }

      router.push('/appointments')
    } catch (error) {
      console.error('Failed to save appointment:', error)
    }
  }
}
</script>
