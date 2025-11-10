<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="flex justify-between items-center mb-6">
      <router-link to="/patients">
        <v-btn icon variant="text" color="primary">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
      </router-link>

      <div class="flex-1">
        <h1 class="text-3xl font-bold text-gray-800">Patient Details</h1>
      </div>

      <router-link v-if="patientStore.currentPatient" :to="`/patients/${patientStore.currentPatient.id}/edit`">
        <v-btn color="primary" prepend-icon="mdi-pencil">
          Edit
        </v-btn>
      </router-link>
    </div>

    <!-- Loading State -->
    <v-card v-if="patientStore.loading" class="mb-6">
      <v-card-text class="text-center py-8">
        <v-progress-circular indeterminate color="primary" />
        <p class="mt-4 text-gray-600">Loading patient details...</p>
      </v-card-text>
    </v-card>

    <!-- Error State -->
    <v-card v-else-if="patientStore.error" class="mb-6" color="error">
      <v-card-text class="d-flex align-center gap-3">
        <v-icon color="white">mdi-alert-circle</v-icon>
        <div>
          <p class="text-white font-medium">Error loading patient</p>
          <p class="text-white text-sm">{{ patientStore.error }}</p>
        </div>
      </v-card-text>
    </v-card>

    <!-- Patient Details -->
    <div v-else-if="patientStore.currentPatient">
      <v-row>
        <!-- Main Information -->
        <v-col cols="12" md="8">
          <v-card class="mb-6">
            <v-card-title class="pb-4">Personal Information</v-card-title>
            <v-divider></v-divider>
            <v-card-text class="pt-6">
              <v-row>
                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Full Name</label>
                    <p class="text-lg font-medium text-gray-800">
                      {{ getPatientName(patientStore.currentPatient) }}
                    </p>
                  </div>
                </v-col>

                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Gender</label>
                    <p class="text-lg font-medium text-gray-800">
                      {{ capitalize(patientStore.currentPatient.gender) }}
                    </p>
                  </div>
                </v-col>

                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Date of Birth</label>
                    <p class="text-lg font-medium text-gray-800">
                      {{ formatDate(patientStore.currentPatient.birthDate) }}
                    </p>
                  </div>
                </v-col>

                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Status</label>
                    <v-chip
                      :color="patientStore.currentPatient.active ? 'success' : 'error'"
                      variant="elevated"
                      size="small"
                    >
                      {{ patientStore.currentPatient.active ? 'Active' : 'Inactive' }}
                    </v-chip>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Contact Information -->
          <v-card class="mb-6">
            <v-card-title class="pb-4">Contact Information</v-card-title>
            <v-divider></v-divider>
            <v-card-text class="pt-6">
              <v-row>
                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Email</label>
                    <p class="text-gray-800">
                      {{ getTelecom('email') || 'Not provided' }}
                    </p>
                  </div>
                </v-col>

                <v-col cols="12" sm="6">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Phone</label>
                    <p class="text-gray-800">
                      {{ getTelecom('phone') || 'Not provided' }}
                    </p>
                  </div>
                </v-col>

                <v-col cols="12">
                  <div class="mb-4">
                    <label class="text-xs font-semibold text-gray-600">Address</label>
                    <p class="text-gray-800">
                      {{
                        getAddress()
                          ? `${getAddress().line?.join(', ')}, ${getAddress().city}, ${getAddress().state}`
                          : 'Not provided'
                      }}
                    </p>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Identifiers -->
          <v-card v-if="patientStore.currentPatient.identifier?.length">
            <v-card-title class="pb-4">Identifiers</v-card-title>
            <v-divider></v-divider>
            <v-card-text class="pt-6">
              <v-list>
                <v-list-item
                  v-for="(id, idx) in patientStore.currentPatient.identifier"
                  :key="idx"
                >
                  <v-list-item-title>{{ id.system }}</v-list-item-title>
                  <v-list-item-subtitle>{{ id.value }}</v-list-item-subtitle>
                </v-list-item>
              </v-list>
            </v-card-text>
          </v-card>
        </v-col>

        <!-- Sidebar -->
        <v-col cols="12" md="4">
          <!-- Metadata -->
          <v-card class="mb-6">
            <v-card-title class="pb-4 text-base">Information</v-card-title>
            <v-divider></v-divider>
            <v-card-text class="pt-4">
              <div class="space-y-4">
                <div>
                  <label class="text-xs font-semibold text-gray-600">Patient ID</label>
                  <p class="text-sm text-gray-800 font-mono">
                    {{ patientStore.currentPatient.id }}
                  </p>
                </div>

                <v-divider></v-divider>

                <div>
                  <label class="text-xs font-semibold text-gray-600">Created</label>
                  <p class="text-sm text-gray-800">
                    {{ formatDateTime(patientStore.currentPatient.meta?.created) }}
                  </p>
                </div>

                <v-divider></v-divider>

                <div>
                  <label class="text-xs font-semibold text-gray-600">Last Updated</label>
                  <p class="text-sm text-gray-800">
                    {{ formatDateTime(patientStore.currentPatient.meta?.updated) }}
                  </p>
                </div>
              </div>
            </v-card-text>
          </v-card>

          <!-- Quick Actions -->
          <v-card>
            <v-card-title class="pb-4 text-base">Actions</v-card-title>
            <v-divider></v-divider>
            <v-card-text class="pt-4 space-y-2">
              <router-link :to="`/appointments/new`">
                <v-btn block color="primary" variant="elevated">
                  <v-icon start>mdi-plus</v-icon>
                  New Appointment
                </v-btn>
              </router-link>

              <v-btn block color="warning" variant="outlined">
                <v-icon start>mdi-download</v-icon>
                Export Record
              </v-btn>

              <v-btn
                block
                color="error"
                variant="text"
                @click="deletePatient(patientStore.currentPatient?.id || '')"
              >
                <v-icon start>mdi-trash-can</v-icon>
                Delete Patient
              </v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <!-- Empty State -->
    <v-card v-else class="text-center py-12">
      <v-icon size="48" class="mb-4 text-gray-400">mdi-hospital-box</v-icon>
      <p class="text-lg text-gray-600">Patient not found</p>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePatientStore } from '@/stores/patient'
import type { Patient } from '@/stores/types'

const route = useRoute()
const router = useRouter()
const patientStore = usePatientStore()

onMounted(async () => {
  const id = route.params.id as string
  await patientStore.fetchPatientById(id)
})

const getPatientName = (patient: Patient) => {
  const name = patient.name[0]
  if (!name) return 'Unknown'
  return `${name.given?.join(' ') || ''} ${name.family || ''}`.trim()
}

const getTelecom = (system: string) => {
  return patientStore.currentPatient?.telecom?.find(t => t.system === system)?.value
}

const getAddress = () => {
  return patientStore.currentPatient?.address?.[0]
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

const formatDateTime = (dateString?: string) => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleString('en-US')
}

const capitalize = (text: string) => {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

const deletePatient = async (id: string) => {
  if (confirm('Are you sure you want to delete this patient? This action cannot be undone.')) {
    try {
      await patientStore.deletePatient(id)
      router.push('/patients')
    } catch (error) {
      console.error('Failed to delete patient:', error)
    }
  }
}
</script>
