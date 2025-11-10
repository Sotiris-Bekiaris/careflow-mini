<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-800">Patients</h1>
        <p class="text-gray-600 mt-1">Manage patient records</p>
      </div>

      <router-link to="/patients/new">
        <v-btn color="primary" prepend-icon="mdi-plus" size="large">
          Add Patient
        </v-btn>
      </router-link>
    </div>

    <!-- Search and Filter -->
    <v-card class="mb-6">
      <v-card-text class="pt-4">
        <v-row>
          <v-col cols="12" md="6">
            <v-text-field
              v-model="searchQuery"
              label="Search patients..."
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
              density="compact"
              clearable
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              v-model="filterGender"
              :items="['all', 'male', 'female', 'other', 'unknown']"
              label="Filter by gender"
              variant="outlined"
              density="compact"
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Loading State -->
    <v-card v-if="patientStore.loading" class="mb-6">
      <v-card-text class="text-center py-8">
        <v-progress-circular indeterminate color="primary" />
        <p class="mt-4 text-gray-600">Loading patients...</p>
      </v-card-text>
    </v-card>

    <!-- Error State -->
    <v-card v-else-if="patientStore.error" class="mb-6" color="error">
      <v-card-text class="d-flex align-center gap-3">
        <v-icon color="white">mdi-alert-circle</v-icon>
        <div>
          <p class="text-white font-medium">Error loading patients</p>
          <p class="text-white text-sm">{{ patientStore.error }}</p>
        </div>
      </v-card-text>
    </v-card>

    <!-- Patients Table -->
    <v-card v-else>
      <v-table v-if="patientStore.hasPatients">
        <thead>
          <tr>
            <th class="text-left">Name</th>
            <th class="text-left">Gender</th>
            <th class="text-left">Date of Birth</th>
            <th class="text-left">Contact</th>
            <th class="text-left">Status</th>
            <th class="text-center">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="patient in filteredPatients" :key="patient.id">
            <td>
              <strong>{{ getPatientName(patient) }}</strong>
            </td>
            <td>
              <v-chip size="small" variant="outlined">
                {{ capitalize(patient.gender) }}
              </v-chip>
            </td>
            <td>{{ formatDate(patient.birthDate) }}</td>
            <td>{{ getContactInfo(patient) }}</td>
            <td>
              <v-chip
                size="small"
                :color="patient.active ? 'success' : 'error'"
                variant="elevated"
              >
                {{ patient.active ? 'Active' : 'Inactive' }}
              </v-chip>
            </td>
            <td class="text-center">
              <router-link :to="`/patients/${patient.id}`">
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  color="primary"
                  title="View"
                >
                  <v-icon size="small">mdi-eye</v-icon>
                </v-btn>
              </router-link>

              <router-link :to="`/patients/${patient.id}/edit`">
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  color="info"
                  title="Edit"
                >
                  <v-icon size="small">mdi-pencil</v-icon>
                </v-btn>
              </router-link>

              <v-btn
                icon
                size="x-small"
                variant="text"
                color="error"
                title="Delete"
                @click="deletePatient(patient.id)"
              >
                <v-icon size="small">mdi-trash-can</v-icon>
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>

      <v-card-text v-else class="text-center py-12 text-gray-600">
        <v-icon size="48" class="mb-4 text-gray-400">mdi-hospital-box</v-icon>
        <p class="text-lg">No patients found</p>
        <router-link to="/patients/new" class="mt-4">
          <v-btn color="primary" variant="outlined">
            Create First Patient
          </v-btn>
        </router-link>
      </v-card-text>
    </v-card>

    <!-- Pagination (Placeholder) -->
    <div v-if="patientStore.hasPatients" class="mt-6 flex justify-center">
      <v-pagination
        v-model="currentPage"
        :length="totalPages"
        color="primary"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { usePatientStore } from '@/stores/patient'
import type { Patient } from '@/stores/types'

const patientStore = usePatientStore()
const searchQuery = ref('')
const filterGender = ref('all')
const currentPage = ref(1)
const itemsPerPage = 10

onMounted(async () => {
  await patientStore.fetchPatients()
})

const filteredPatients = computed(() => {
  let filtered = patientStore.sortedPatients

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(patient => {
      const name = getPatientName(patient).toLowerCase()
      const contact = getContactInfo(patient).toLowerCase()
      return name.includes(query) || contact.includes(query)
    })
  }

  if (filterGender.value !== 'all') {
    filtered = filtered.filter(patient => patient.gender === filterGender.value)
  }

  return filtered
})

const totalPages = computed(() => {
  return Math.ceil(filteredPatients.value.length / itemsPerPage)
})

const getPatientName = (patient: Patient) => {
  const name = patient.name[0]
  if (!name) return 'Unknown'
  return `${name.given?.join(' ') || ''} ${name.family || ''}`.trim()
}

const getContactInfo = (patient: Patient) => {
  const email = patient.telecom?.find(t => t.system === 'email')
  const phone = patient.telecom?.find(t => t.system === 'phone')
  return email?.value || phone?.value || 'N/A'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const capitalize = (text: string) => {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

const deletePatient = async (id: string) => {
  if (confirm('Are you sure you want to delete this patient?')) {
    try {
      await patientStore.deletePatient(id)
      patientStore.clearError()
    } catch (error) {
      console.error('Failed to delete patient:', error)
    }
  }
}
</script>
